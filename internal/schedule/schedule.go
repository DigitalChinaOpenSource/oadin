// Apache v2 license
// Copyright (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package schedule

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"

	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/event"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils"
)

type ServiceTaskEventType int

const (
	ServiceTaskEnqueue ServiceTaskEventType = iota
	ServiceTaskFailed
	ServiceTaskDone
)

type ServiceTaskEvent struct {
	Type  ServiceTaskEventType
	Task  *ServiceTask
	Error error // only for ServiceTaskFailed
}

type ServiceScheduler interface {
	// Enqueue itself is a non-blocking API. It returns the task ID and
	// a chan to return results as the Restful service
	Enqueue(*types.ServiceRequest) (uint64, chan *types.ServiceResult)
	Start()
	TaskComplete(*ServiceTask, error)
}

type BasicServiceScheduler struct {
	curID       uint64
	WaitingList *utils.SafeList
	RunningList *utils.SafeList
	ChEvent     chan *ServiceTaskEvent
}

func NewBasicServiceScheduler() *BasicServiceScheduler {
	return &BasicServiceScheduler{
		WaitingList: utils.NewSafeList(),
		RunningList: utils.NewSafeList(),
		ChEvent:     make(chan *ServiceTaskEvent, 600),
	}
}

func (ss *BasicServiceScheduler) Enqueue(req *types.ServiceRequest) (uint64, chan *types.ServiceResult) {
	ch := make(chan *types.ServiceResult, 600)
	ss.curID += 1
	// we don't close ch here. It should be closed when the task is done
	task := &ServiceTask{Request: req, Ch: ch}
	task.Schedule.Id = ss.curID
	ss.ChEvent <- &ServiceTaskEvent{Type: ServiceTaskEnqueue, Task: task}
	return task.Schedule.Id, ch
}

func (ss *BasicServiceScheduler) TaskComplete(task *ServiceTask, err error) {
	if task.Schedule.ListMark == nil {
		panic("[Schedule] See a task without a list mark")
	}
	if err == nil {
		ss.ChEvent <- &ServiceTaskEvent{Type: ServiceTaskDone, Task: task}
	} else {
		ss.ChEvent <- &ServiceTaskEvent{Type: ServiceTaskFailed, Task: task, Error: err}
	}
}

func (ss *BasicServiceScheduler) Start() {
	logger.LogicLogger.Info("[Init] Start basic service scheduler ...")
	go func() {
		for taskEvent := range ss.ChEvent {
			task := taskEvent.Task
			switch taskEvent.Type {
			case ServiceTaskEnqueue:
				ss.onTaskEnqueue(task)
			case ServiceTaskDone:
				ss.onTaskDone(task)
			case ServiceTaskFailed:
				ss.onTaskFailed(task, taskEvent.Error)
			}
			ss.schedule()
		}
	}()
}

func (ss *BasicServiceScheduler) onTaskEnqueue(task *ServiceTask) {
	logger.LogicLogger.Info("[Schedule] Enqueue", "task", task)
	ss.addToList(task, "waiting")
	task.Schedule.TimeEnqueue = time.Now()
}

func (ss *BasicServiceScheduler) onTaskDone(task *ServiceTask) {
	logger.LogicLogger.Info("[Schedule] Task Done", "since queued", time.Since(task.Schedule.TimeEnqueue),
		"since run", time.Since(task.Schedule.TimeRun), "task", task)
	task.Schedule.TimeComplete = time.Now()
	close(task.Ch)
	ss.removeFromList(task)
}

func (ss *BasicServiceScheduler) onTaskFailed(task *ServiceTask, err error) {
	logger.LogicLogger.Error("[Service] Task Failed", "error", err.Error(), "since queued",
		time.Since(task.Schedule.TimeEnqueue), "since run", time.Since(task.Schedule.TimeRun), "task", task)
	task.Error = err
	task.Schedule.TimeComplete = time.Now()
	close(task.Ch)
	ss.removeFromList(task)
}

func (ss *BasicServiceScheduler) addToList(task *ServiceTask, list string) {
	switch list {
	case "waiting":
		mark := ss.WaitingList.PushBack(task)
		task.Schedule.ListMark = mark
	case "running":
		mark := ss.RunningList.PushBack(task)
		task.Schedule.ListMark = mark
	default:
		panic("[Schedule] Invalid list name: " + list)
	}
}

func (ss *BasicServiceScheduler) removeFromList(task *ServiceTask) {
	if task.Schedule.IsRunning {
		ss.RunningList.Remove(task.Schedule.ListMark)
	} else {
		ss.WaitingList.Remove(task.Schedule.ListMark)
	}
}

// returns priority, smaller more preferred to pick
// 1 - if exactly match
// 2 - if ask is prefix of got, e.g. asks llama3.1, got llama3.1-int8
// 3 - if got is prefix of ask, e.g. asks llama3.1, got llama3
// 4 - if ask is part of got but not prefix, e.g. asks llama3.1, got my-llama3.1-int8
// 5 - if got is part of ask but not prefix, e.g. asks llama3.1, got lama
// 6 - otherwise
func modelPriority(ask, got string) int {
	if ask == got {
		return 1
	}
	if strings.HasPrefix(got, ask) {
		return 2
	}
	if strings.HasPrefix(ask, got) {
		return 3
	}
	if strings.Contains(got, ask) {
		return 4
	}
	if strings.Contains(ask, got) {
		return 5
	}
	return 6
}

// Decide the running details - local or remote, which model, which xpu etc.
// It will fill in the task.Target field if need to run now
// So if task.Target is still nil, it means the task is not ready to run
func (ss *BasicServiceScheduler) dispatch(task *ServiceTask) (*types.ServiceTarget, error) {
	// Location Selection
	// ================
	// TODO: so far we all dispatch to local, unless force

	location := types.ServiceSourceLocal
	model := task.Request.Model
	if task.Request.HybridPolicy == "always_local" {
		location = types.ServiceSourceLocal
	} else if task.Request.HybridPolicy == "always_remote" {
		location = types.ServiceSourceRemote
	} else if task.Request.HybridPolicy == "default" {
		if model == "" {
			gpuUtilization, err := utils.GetGpuInfo()
			if err != nil {
				cpuTotalPercent, _ := cpu.Percent(15*time.Second, false)
				if cpuTotalPercent[0] > 80.0 {
					location = types.ServiceSourceRemote
				}
			}
			if gpuUtilization >= 80.0 {
				location = types.ServiceSourceRemote
			}
		}
	}
	ds := datastore.GetDefaultDatastore()
	service := &types.Service{
		Name: task.Request.Service,
	}

	err := ds.Get(context.Background(), service)
	if err != nil {
		return nil, fmt.Errorf("service not found: %s", task.Request.Service)
	}

	if service.LocalProvider == "" && service.RemoteProvider == "" {
		return nil, fmt.Errorf("service %s does not have local or remote provider", task.Request.Service)
	}
	// Provider Selection
	// ================
	providerName := service.LocalProvider
	if location == types.ServiceSourceRemote {
		if service.RemoteProvider == "" {
			providerName = service.LocalProvider
		} else {
			providerName = service.RemoteProvider
		}
	} else if service.LocalProvider == "" {
		providerName = service.RemoteProvider
	}

	sp := &types.ServiceProvider{
		ProviderName: providerName,
	}
	err = ds.Get(context.Background(), sp)
	if err != nil {
		return nil, fmt.Errorf("service provider not found for %s of Service %s", location, task.Request.Service)
	}
	providerProperties := &types.ServiceProviderProperties{}
	err = json.Unmarshal([]byte(sp.Properties), providerProperties)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal service provider properties: %v", err)
	}
	// Non-query model services do not require model validation
	if task.Request.Service != types.ServiceModels {
		if model == "" {
			switch location {
			case types.ServiceSourceLocal:
				m := &types.Model{
					ProviderName: sp.ProviderName,
				}
				sortOption := []datastore.SortOption{
					{Key: "updated_at", Order: -1},
				}
				ms, err := ds.List(context.Background(), m, &datastore.ListOptions{
					FilterOptions: datastore.FilterOptions{
						Queries: []datastore.FuzzyQueryOption{
							{Key: "status", Query: "downloaded"},
						},
					},
					SortBy: sortOption})
				if err != nil {
					return nil, fmt.Errorf("model not found for %s of Service %s", location, task.Request.Service)
				}
				if len(ms) == 0 {
					return nil, fmt.Errorf("model not found for %s of Service %s", location, task.Request.Service)
				}
				model = ms[0].(*types.Model).ModelName
			case types.ServiceSourceRemote:
				defaultInfo := GetProviderServiceDefaultInfo(sp.Flavor, task.Request.Service)
				model = defaultInfo.DefaultModel
			}
		} else {
			m := &types.Model{
				ModelName: task.Request.Model,
			}
			err := ds.Get(context.Background(), m)
			if err != nil {
				return nil, fmt.Errorf("model not found for %s of Service %s", location, task.Request.Service)
			}
			if m.Status != "downloaded" {
				return nil, fmt.Errorf("model installing %s of Service %s, please wait", location, task.Request.Service)
			}
		}
	}

	// Model Selection
	// ================
	// pick the smallest priority number, which means the most preferred
	// if more than one candidate for the same priority, pick the 1st one
	// TODO(Strategies to be discussed later)
	//model := task.Request.Model
	//if task.Request.Model != "" && len(ms) > 0 {
	//	curPriority := 8
	//	for _, m := range ms {
	//		priority := modelPriority(task.Request.Model, m.(*types.Model).ModelName)
	//		if priority < curPriority {
	//			curPriority = priority
	//			model = m.(*types.Model).ModelName
	//		}
	//	}
	//	if model != task.Request.Model {
	//		slog.Warn("[Schedule] Model mismatch between Request and Service Provider",
	//			"expect_model", task.Request.Model, "selected_model", model,
	//			"id_service_provider", sp.ProviderName, "all_models")
	//	}
	//}

	// Stream Mode Selection
	// ================
	stream := task.Request.AskStreamMode
	// assume it supports stream mode if not specified supported_response_mode
	if stream && len(providerProperties.SupportedResponseMode) > 0 {
		stream = false
		for _, mode := range providerProperties.SupportedResponseMode {
			if mode == "stream" {
				stream = true
				break
			}
		}
		if !stream {
			slog.Warn("[Schedule] Asks for stream mode but it is not supported by the service provider",
				"id_service_provider", sp.ProviderName, "supported_response_mode", providerProperties.SupportedResponseMode)
		}
	}

	// Stream Mode Selection
	// ================
	// TODO: XPU selection

	return &types.ServiceTarget{
		Location:        location,
		Stream:          stream,
		Model:           model,
		ToFavor:         sp.Flavor,
		ServiceProvider: sp,
	}, nil
}

// this is invoked by schedule goroutine
func (ss *BasicServiceScheduler) schedule() {
	// TODO: currently, we run all of the
	for e := ss.WaitingList.Front(); e != nil; e = e.Next() {
		task := e.Value.(*ServiceTask)
		target, err := ss.dispatch(task)
		if err != nil {
			task.Ch <- &types.ServiceResult{Type: types.ServiceResultFailed, TaskId: task.Schedule.Id, Error: err}
			ss.onTaskFailed(task, err)
			continue
		}
		task.Target = target
		ss.removeFromList(task)
		ss.addToList(task, "running")
		task.Schedule.IsRunning = true
		task.Schedule.TimeRun = time.Now()
		logger.LogicLogger.Info("[Schedule] Start to run the task", "taskid", task.Schedule.Id, "service", task.Request.Service,
			"location", task.Target.Location, "service_provider", task.Target.ServiceProvider)
		// REALLY run the task
		go func() {
			err := task.Run()
			// need to send back error to the client
			if err != nil {
				task.Ch <- &types.ServiceResult{Type: types.ServiceResultFailed, TaskId: task.Schedule.Id, Error: err}
			}
			ss.TaskComplete(task, err)
		}()
	}
}

var scheduler ServiceScheduler

func StartScheduler(s string) {
	if scheduler != nil {
		panic("Default scheduler is already set")
	}
	switch s {
	case "basic":
		scheduler = NewBasicServiceScheduler()
		scheduler.Start()
	default:
		panic(fmt.Sprintf("Invalid scheduler type: %s", s))
	}
}

func GetScheduler() ServiceScheduler {
	if scheduler == nil {
		panic("Scheduler is not started yet")
	}
	return scheduler
}

func InvokeService(fromFlavor string, service string, request *http.Request) (uint64, chan *types.ServiceResult, error) {
	logger.LogicLogger.Info("[Service] Invoking Service", "fromFlavor", fromFlavor, "service", service)

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return 0, nil, err
	}

	event.SysEvents.NotifyHTTPRequest("receive_service_request", request.Method,
		request.URL.String(), request.Header, body)

	if request.Method == http.MethodGet {
		queryParams := request.URL.Query()
		queryParamsJSON, err := json.Marshal(queryParams)
		if err != nil {
			logger.LogicLogger.Error("[Service] Failed to unmarshal GET request", "error", err, "body", string(body))
			return 0, nil, err
		}
		logger.LogicLogger.Debug("[Service] GET Request Query Params", "params", string(queryParamsJSON))

		body = queryParamsJSON
	} // TODO: handle the case that the body is not json and not text
	if request.Method == http.MethodPost &&
		!strings.Contains(request.Header.Get("Content-Type"), "application/json") &&
		!strings.Contains(request.Header.Get("Content-Type"), "text/plain") {
		panic("TO SUPPORT non JSON or non text request")
	}
	hybridPolicy := "default"
	if service != "" {
		ds := datastore.GetDefaultDatastore()
		sp := &types.Service{
			Name:   service,
			Status: 1,
		}
		err = ds.Get(context.Background(), sp)
		if err != nil {
			logger.LogicLogger.Error("[Schedule] Failed to get service", "error", err, "service", service)
		}
		hybridPolicy = sp.HybridPolicy
	}

	serviceRequest := types.ServiceRequest{
		FromFlavor:      fromFlavor,
		Service:         service,
		Priority:        0,
		HTTP:            types.HTTPContent{Body: body, Header: request.Header},
		OriginalRequest: request,
		HybridPolicy:    hybridPolicy,
	}

	err = json.Unmarshal(body, &serviceRequest)
	if err != nil {
		logger.LogicLogger.Error("[Service] Failed to unmarshal POST request", "error", err, "body", string(body))
		return 0, nil, err
	}

	taskid, ch := GetScheduler().Enqueue(&serviceRequest)

	return taskid, ch, err
}
