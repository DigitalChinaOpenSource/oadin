//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

package schedule

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils"
	"intel.com/aog/internal/utils/bcode"
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
			//gpuUtilization, err := utils.GetGpuInfo()
			//if err != nil {
			//cpuTotalPercent, _ := cpu.Percent(15*time.Second, false)
			//if cpuTotalPercent[0] > 80.0 {
			//	location = types.ServiceSourceRemote
			//}
			//}
			//if gpuUtilization >= 80.0 {
			//	location = types.ServiceSourceRemote
			//}
		}
	}
	ds := datastore.GetDefaultDatastore()
	service := &types.Service{
		Name: task.Request.Service,
	}

	err := ds.Get(context.Background(), service)
	if err != nil {
		logger.LogicLogger.Error("[Schedule] Failed to get service", "error", err, "service", task.Request.Service)
		return nil, bcode.ErrServiceRecordNotFound
	}

	if service.LocalProvider == "" && service.RemoteProvider == "" {
		logger.LogicLogger.Error("[Schedule] Service ", task.Request.Service, " does not have local or remote provider")
		return nil, bcode.ErrNotExistDefaultProvider
	}

	m := &types.Model{
		ModelName:   task.Request.Model,
		ServiceName: task.Request.Service,
	}

	// Provider Selection
	// ================
	providerName := service.LocalProvider
	if model == "" {
		if location == types.ServiceSourceRemote {
			if service.RemoteProvider == "" {
				providerName = service.LocalProvider
			} else {
				providerName = service.RemoteProvider
			}
		} else if service.LocalProvider == "" {
			providerName = service.RemoteProvider
		}
	} else {
		err := ds.Get(context.Background(), m)
		if err != nil {
			logger.LogicLogger.Error("[Schedule] model not found", "error", err, "model", task.Request.Model)
			return nil, bcode.ErrModelRecordNotFound
		}
		if m.Status != "downloaded" {
			logger.LogicLogger.Error("[Schedule] model installing", "model", task.Request.Model, "status", m.Status)
			return nil, bcode.ErrModelUnDownloaded
		}

		providerName = m.ProviderName
	}

	sp := &types.ServiceProvider{
		ProviderName: providerName,
	}
	err = ds.Get(context.Background(), sp)
	if err != nil {
		logger.LogicLogger.Error("[Schedule] service provider not found for ", location, " of Service ", task.Request.Service)
		return nil, bcode.ErrProviderNotExist
	}

	location = sp.ServiceSource
	providerProperties := &types.ServiceProviderProperties{}
	err = json.Unmarshal([]byte(sp.Properties), providerProperties)
	if err != nil {
		logger.LogicLogger.Error("[Schedule] failed to unmarshal service provider properties", "error", err, "properties", sp.Properties)
		return nil, bcode.ErrUnmarshalProviderProperties
	}
	// Non-query model services do not require model validation
	if task.Request.Service != types.ServiceModels {
		if model == "" {
			switch location {
			case types.ServiceSourceLocal:
				m := &types.Model{
					ProviderName: sp.ProviderName,
				}
				// 先查找 is_default=true 且 downloaded 的模型
				ms, err := ds.List(context.Background(), m, &datastore.ListOptions{
					FilterOptions: datastore.FilterOptions{
						Queries: []datastore.FuzzyQueryOption{
							{Key: "status", Query: "downloaded"},
							{Key: "is_default", Query: "true"},
						},
					},
				})

				if err == nil && len(ms) > 0 {
					model = ms[0].(*types.Model).ModelName
				} else {
					// 没有default，再查找第一个downloaded模型
					sortOption := []datastore.SortOption{
						{Key: "updated_at", Order: -1},
					}
					ms, err := ds.List(context.Background(), m, &datastore.ListOptions{
						FilterOptions: datastore.FilterOptions{
							Queries: []datastore.FuzzyQueryOption{
								{Key: "status", Query: "downloaded"},
							},
						},
						SortBy: sortOption,
					})
					if err != nil {
						return nil, fmt.Errorf("model not found for %s of Service %s", location, task.Request.Service)
					}
					if len(ms) == 0 {
						return nil, fmt.Errorf("model not found for %s of Service %s", location, task.Request.Service)
					}
					model = ms[0].(*types.Model).ModelName
				}
			case types.ServiceSourceRemote:
				defaultInfo := GetProviderServiceDefaultInfo(sp.Flavor, task.Request.Service)
				model = defaultInfo.DefaultModel
			}
		}
	}

	// Model Selection
	// ================
	// pick the smallest priority number, which means the most preferred
	// if more than one candidate for the same priority, pick the 1st one
	// TODO(Strategies to be discussed later)

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

	// 从YAML配置中获取服务协议类型
	flavorDef := GetFlavorDef(sp.Flavor)
	protocol := ""
	if serviceDef, exists := flavorDef.Services[task.Request.Service]; exists {
		protocol = serviceDef.Protocol
	}

	return &types.ServiceTarget{
		Location:        location,
		Stream:          stream,
		Model:           model,
		ToFavor:         sp.Flavor,
		Protocol:        protocol, // 设置Protocol字段
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
		logger.LogicLogger.Info("[Schedule] Start to run the task", "taskID", task.Schedule.Id, "service", task.Request.Service,
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

	if request.Method != http.MethodGet && request.Method != http.MethodPost {
		logger.LogicLogger.Error("[Service] Unsupported request method", "method", request.Method)
		return 0, nil, bcode.ErrUnSupportRequestMethod
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		logger.LogicLogger.Error("[Service] Failed to read request body", "error", err)
		return 0, nil, bcode.ErrReadRequestBody
	}

	// 重新设置请求体，以便后续处理可能需要再次读取
	request.Body = io.NopCloser(bytes.NewReader(body))

	wsConnID := request.Header.Get("X-WebSocket-ConnID")
	isWebSocket := wsConnID != ""

	if request.Method == http.MethodGet && !isWebSocket {
		queryParams := request.URL.Query()
		queryParamsJSON, err := json.Marshal(queryParams)
		if err != nil {
			logger.LogicLogger.Error("[Service] Failed to unmarshal GET request", "error", err, "body", string(body))
			return 0, nil, bcode.ErrUnmarshalRequestBody
		}
		logger.LogicLogger.Debug("[Service] GET Request Query Params", "params", string(queryParamsJSON))

		body = queryParamsJSON
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

	// 根据Content-Type决定如何处理请求体
	contentType := request.Header.Get("Content-Type")
	isBinary := strings.Contains(contentType, "application/octet-stream") ||
		strings.Contains(contentType, "audio/") ||
		strings.Contains(contentType, "video/") ||
		strings.Contains(contentType, "image/")

	// 处理WebSocket相关信息
	if isWebSocket {
		// 如果是WebSocket请求，添加相关标记
		serviceRequest.WebSocketConnID = wsConnID

		// 记录WebSocket请求特有的日志
		logger.LogicLogger.Debug("[Service] Processing WebSocket request",
			"wsConnID", wsConnID,
			"contentType", contentType,
			"isBinary", isBinary,
			"bodySize", len(body))

		// 如果是二进制数据，使用更适合的日志格式
		if isBinary {
			logger.LogicLogger.Debug("[Service] Binary WebSocket data received",
				"wsConnID", wsConnID,
				"bodySize", len(body),
				"firstBytes", fmt.Sprintf("%x", body[:utils.Min(20, len(body))]))
		}
	}

	// 只有当不是二进制数据时，才尝试解析JSON
	if !isBinary {
		err = json.Unmarshal(body, &serviceRequest)
		if err != nil {
			// 对于WebSocket文本消息，可能是特定指令而不是完整的JSON结构
			if isWebSocket {
				logger.LogicLogger.Debug("[Service] WebSocket message is not valid JSON, treating as raw message",
					"wsConnID", wsConnID,
					"error", err)
			} else {
				logger.LogicLogger.Error("[Service] Failed to unmarshal POST request", "error", err, "body", string(body))
				return 0, nil, err
			}
		}
	}

	taskID, ch := GetScheduler().Enqueue(&serviceRequest)

	// 对于WebSocket请求，记录分配的taskID
	if isWebSocket {
		logger.LogicLogger.Info("[Service] WebSocket request enqueued",
			"wsConnID", wsConnID,
			"taskID", taskID)
	}

	return taskID, ch, err
}
