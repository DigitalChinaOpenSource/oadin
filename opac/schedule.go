// Apache v2 license
// Copyright (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package opac

import (
	"container/list"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"intel.com/opac/helpers"
)

type ScheduleDetails struct {
	Id           uint64
	IsRunning    bool
	ListMark     *list.Element
	TimeEnqueue  time.Time
	TimeRun      time.Time
	TimeComplete time.Time
}

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
	Enqueue(*ServiceRequest) (uint64, chan *ServiceResult)
	Start()
	TaskComplete(*ServiceTask, error)
}

type BasicServiceScheduler struct {
	curID       uint64
	WaitingList *helpers.SafeList
	RunningList *helpers.SafeList
	ChEvent     chan *ServiceTaskEvent
}

func NewBasicServiceScheduler() *BasicServiceScheduler {
	return &BasicServiceScheduler{
		WaitingList: helpers.NewSafeList(),
		RunningList: helpers.NewSafeList(),
		ChEvent:     make(chan *ServiceTaskEvent, 600),
	}
}

func (ss *BasicServiceScheduler) Enqueue(req *ServiceRequest) (uint64, chan *ServiceResult) {
	ch := make(chan *ServiceResult, 600)
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
	slog.Info("[Init] Start basic service scheduler ...")
	go func() {
		for event := range ss.ChEvent {
			task := event.Task
			switch event.Type {
			case ServiceTaskEnqueue:
				ss.onTaskEnqueue(task)
			case ServiceTaskDone:
				ss.onTaskDone(task)
			case ServiceTaskFailed:
				ss.onTaskFailed(task, event.Error)
			}
			ss.schedule()
		}
	}()
}

func (ss *BasicServiceScheduler) onTaskEnqueue(task *ServiceTask) {
	slog.Info("[Schedule] Enqueue", "task", task)
	ss.addToList(task, "waiting")
	task.Schedule.TimeEnqueue = time.Now()
}
func (ss *BasicServiceScheduler) onTaskDone(task *ServiceTask) {
	slog.Info("[Schedule] Task Done", "since queued", time.Since(task.Schedule.TimeEnqueue),
		"since run", time.Since(task.Schedule.TimeRun), "task", task)
	task.Schedule.TimeComplete = time.Now()
	close(task.Ch)
	ss.removeFromList(task)
}
func (ss *BasicServiceScheduler) onTaskFailed(task *ServiceTask, err error) {
	slog.Error("[Service] Task Failed", "error", err.Error(), "since queued",
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
func (ss *BasicServiceScheduler) dispatch(task *ServiceTask) (*ServiceTarget, error) {

	// Location Selection
	// ================
	// TODO: so far we all dispatch to local, unless force
	location := "local"
	if task.Request.HybridPolicy == "always_local" {
		location = "local"
	} else if task.Request.HybridPolicy == "always_remote" {
		location = "remote"
	}
	sp := GetPlatformInfo().GetServiceProviderInfo(task.Request.Service, location)
	if sp == nil {
		return nil, fmt.Errorf("service provider not found for %s of Service %s", location, task.Request.Service)
	}

	// Model Selection
	// ================
	// pick the smallest priority number, which means the most preferred
	// if more than one candidates for the same priority, pick the 1st one
	model := task.Request.Model
	if task.Request.Model != "" && len(sp.Properties.Models) > 0 {
		curPriority := 8
		for _, m := range sp.Properties.Models {
			priority := modelPriority(task.Request.Model, m)
			if priority < curPriority {
				curPriority = priority
				model = m
			}
		}
		if model != task.Request.Model {
			slog.Warn("[Schedule] Model mismatch between Request and Service Provider",
				"expect_model", task.Request.Model, "selected_model", model,
				"id_service_provider", sp.Id, "all_models", sp.Properties.Models)
		}
	}

	// Stream Mode Selection
	// ================
	stream := task.Request.AskStreamMode
	// assume it supports stream mode if not specified supported_response_mode
	if stream && len(sp.Properties.SupportedResponseMode) > 0 {
		stream = false
		for _, mode := range sp.Properties.SupportedResponseMode {
			if mode == "stream" {
				stream = true
				break
			}
		}
		if !stream {
			slog.Warn("[Schedule] Asks for stream mode but it is not supported by the service provider",
				"id_service_provider", sp.Id, "supported_response_mode", sp.Properties.SupportedResponseMode)
		}
	}

	// Stream Mode Selection
	// ================
	// TODO: XPU selection

	return &ServiceTarget{
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
			task.Ch <- &ServiceResult{Type: ServiceResultFailed, TaskId: task.Schedule.Id, Error: err}
			ss.onTaskFailed(task, err)
			continue
		}
		task.Target = target
		ss.removeFromList(task)
		ss.addToList(task, "running")
		task.Schedule.IsRunning = true
		task.Schedule.TimeRun = time.Now()
		slog.Info("[Schedule] Start to run the task", "taskid", task.Schedule.Id, "service", task.Request.Service,
			"location", task.Target.Location, "service_provider", task.Target.ServiceProvider)
		// REALLY run the task
		go func() {
			err := task.Run()
			// need to send back error to the client
			if err != nil {
				task.Ch <- &ServiceResult{Type: ServiceResultFailed, TaskId: task.Schedule.Id, Error: err}
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
