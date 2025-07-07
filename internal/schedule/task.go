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
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"intel.com/aog/internal/convert"
	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/types"
)

type TaskMiddleware interface {
	Handle(st *ServiceTask) error
}

type ServiceTask struct {
	Request  *types.ServiceRequest
	Target   *types.ServiceTarget
	Ch       chan *types.ServiceResult
	Error    error
	Schedule types.ScheduleDetails
}

func (st *ServiceTask) String() string {
	return fmt.Sprintf("ServiceTask{Id: %d, Request: %s, Target: %s}", st.Schedule.Id, st.Request, st.Target)
}

func NewStreamMode(header http.Header) *types.StreamMode {
	mode := types.StreamModeNonStream
	if contentType := header.Get("Content-Type"); contentType != "" {
		ct := strings.ToLower(contentType)
		if strings.Contains(ct, "text/event-stream") {
			mode = types.StreamModeEventStream
		} else if strings.Contains(ct, "application/x-ndjson") {
			mode = types.StreamModeNDJson
		}
	}
	return &types.StreamMode{Mode: mode, Header: header.Clone()}
}

// ServiceHandler defines the interface for handling service requests
type ServiceHandler interface {
	Handle(ctx *ServiceContext) error
}

// HandlerMetrics is used to record handler performance metrics
type HandlerMetrics struct {
	StartTime   time.Time
	Duration    time.Duration
	HandlerName string
}

// MetricsCollector collects handler performance metrics
type MetricsCollector interface {
	CollectMetrics(metrics HandlerMetrics)
}

// DefaultMetricsCollector is the default metrics collector
type DefaultMetricsCollector struct{}

func (c *DefaultMetricsCollector) CollectMetrics(metrics HandlerMetrics) {
	hours := int(metrics.Duration.Hours())
	minutes := int(metrics.Duration.Minutes()) % 60
	seconds := int(metrics.Duration.Seconds()) % 60
	durationStr := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)

	logger.LogicLogger.Info("[Service] Handler Metrics",
		"handler", metrics.HandlerName,
		"duration", durationStr,
		"start_time", metrics.StartTime)
}

// ServiceContext contains the context information needed for service processing
type ServiceContext struct {
	Task       *ServiceTask
	Request    *types.HTTPContent
	Response   *http.Response
	Error      error
	Metrics    MetricsCollector
	RetryCount int
	MaxRetries int
}

// NewServiceContext creates a service context
func NewServiceContext(task *ServiceTask) *ServiceContext {
	return &ServiceContext{
		Task:       task,
		Request:    &task.Request.HTTP,
		Metrics:    &DefaultMetricsCollector{},
		MaxRetries: 2, // default maximum retry count
	}
}

func (st *ServiceTask) Run() error {
	logger.LogicLogger.Debug("[Service] ServiceTask start run......")

	ctx := NewServiceContext(st)

	// Build the processing chain
	middleware := &MiddlewareHandler{BaseHandler: NewBaseHandler("Middleware")}
	reqConverter := &RequestConversionHandler{BaseHandler: NewBaseHandler("RequestConverter")}
	serviceInvoker := &ServiceInvokerHandler{BaseHandler: NewBaseHandler("ServiceInvoker")}
	respHandler := &ResponseHandler{BaseHandler: NewBaseHandler("Response")}

	middleware.SetNext(reqConverter)
	reqConverter.SetNext(serviceInvoker)
	serviceInvoker.SetNext(respHandler)

	return middleware.Handle(ctx)
}

func (st *ServiceTask) invokeGRPCServiceProvider(sp *types.ServiceProvider, content types.HTTPContent) (*http.Response, error) {
	invoker := NewInvoker(st, types.ProtocolGRPC)
	return invoker.Invoke(sp, content)
}

func (st *ServiceTask) invokeHTTPServiceProvider(sp *types.ServiceProvider, content types.HTTPContent) (*http.Response, error) {
	invoker := NewInvoker(st, types.ProtocolHTTP)
	return invoker.Invoke(sp, content)
}

// BaseHandler provides a basic processor implementation
type BaseHandler struct {
	next ServiceHandler
	name string
}

func NewBaseHandler(name string) BaseHandler {
	return BaseHandler{name: name}
}

func (h *BaseHandler) SetNext(handler ServiceHandler) ServiceHandler {
	h.next = handler
	return handler
}

func (h *BaseHandler) HandleNext(ctx *ServiceContext) error {
	if h.next != nil {
		return h.next.Handle(ctx)
	}
	return nil
}

// MiddlewareHandler handles middleware logic
type MiddlewareHandler struct {
	BaseHandler
}

func (h *MiddlewareHandler) Handle(ctx *ServiceContext) error {
	metrics := HandlerMetrics{
		StartTime:   time.Now(),
		HandlerName: h.name,
	}
	defer func() {
		metrics.Duration = time.Since(metrics.StartTime)
		ctx.Metrics.CollectMetrics(metrics)
	}()

	if err := ExecuteMiddleware(ctx.Task); err != nil {
		return err
	}

	if ctx.Task.Target == nil || ctx.Task.Target.ServiceProvider == nil {
		return fmt.Errorf("[Service] ServiceTask is not dispatched before it goes to Run() %s", ctx.Task.String())
	}
	return h.HandleNext(ctx)
}

// RequestConversionHandler handles request conversion
type RequestConversionHandler struct {
	BaseHandler
}

func (h *RequestConversionHandler) Handle(ctx *ServiceContext) error {
	ds := datastore.GetDefaultDatastore()
	sp := &types.ServiceProvider{
		Flavor:        ctx.Task.Target.ToFavor,
		ServiceSource: ctx.Task.Target.Location,
		ServiceName:   ctx.Task.Request.Service,
		Status:        1,
	}

	if err := ds.Get(context.Background(), sp); err != nil {
		return fmt.Errorf("service Provider not found for %s of Service %s", ctx.Task.Target.Location, ctx.Task.Request.Service)
	}

	// 判断请求头是否为二进制数据，若是则跳过转换
	if ctx.Request.Header.Get("Content-Type") != "application/octet-stream" ||
		ctx.Request.Header.Get("Content-Type") != "application/x-binary" {
		requestFlavor, targetFlavor, err := h.getFlavors(ctx.Task)
		if err != nil {
			return err
		}

		if err := h.convertRequest(ctx, requestFlavor, targetFlavor); err != nil {
			return err
		}
	}

	return h.HandleNext(ctx)
}

func (h *RequestConversionHandler) getFlavors(st *ServiceTask) (requestFlavor, targetFlavor APIFlavor, err error) {
	requestFlavor, err = GetAPIFlavor(st.Request.FromFlavor)
	if err != nil {
		return nil, nil, fmt.Errorf("[Service] Unsupported API Flavor %s for Request: %s", st.Request.FromFlavor, err.Error())
	}

	targetFlavor, err = GetAPIFlavor(st.Target.ServiceProvider.Flavor)
	if err != nil {
		return nil, nil, fmt.Errorf("[Service] Unsupported API Flavor %s for Service Provider: %s", st.Target.ServiceProvider.Flavor, err.Error())
	}

	return requestFlavor, targetFlavor, nil
}

func (h *RequestConversionHandler) convertRequest(ctx *ServiceContext, requestFlavor, targetFlavor APIFlavor) error {
	if targetFlavor.Name() == requestFlavor.Name() {
		return nil
	}

	requestCtx := convert.ConvertContext{"stream": ctx.Task.Target.Stream}
	if ctx.Task.Target.Model != "" {
		requestCtx["model"] = ctx.Task.Target.Model
	}

	content, err := ConvertBetweenFlavors(requestFlavor, targetFlavor, ctx.Task.Request.Service, "request", *ctx.Request, requestCtx)
	if err != nil {
		return fmt.Errorf("[Service] Failed to convert request: %s", err.Error())
	}

	ctx.Request = &content
	return nil
}

// RetryableError defines a retryable error
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("retryable error: %v", e.Err)
}

// ServiceInvokerHandler handles service invocation
type ServiceInvokerHandler struct {
	BaseHandler
}

func (h *ServiceInvokerHandler) Handle(ctx *ServiceContext) error {
	var lastErr error
	for ctx.RetryCount < ctx.MaxRetries {
		err := h.tryInvoke(ctx)
		if err == nil {
			return h.HandleNext(ctx)
		}

		var retryableError *RetryableError
		if !errors.As(err, &retryableError) {
			return err
		}

		lastErr = err
		ctx.RetryCount++
		logger.LogicLogger.Warn("[Service] Retrying request",
			"attempt", ctx.RetryCount,
			"max_retries", ctx.MaxRetries,
			"error", err)

		// 指数退避
		time.Sleep(time.Duration(1<<uint(ctx.RetryCount)) * time.Second)
	}

	return fmt.Errorf("max retries exceeded: %v", lastErr)
}

func (h *ServiceInvokerHandler) tryInvoke(ctx *ServiceContext) error {
	var err error
	if ctx.Task.Target.ServiceProvider.Flavor == types.FlavorOpenvino {
		ctx.Response, err = ctx.Task.invokeGRPCServiceProvider(ctx.Task.Target.ServiceProvider, *ctx.Request)
	} else {
		ctx.Response, err = ctx.Task.invokeHTTPServiceProvider(ctx.Task.Target.ServiceProvider, *ctx.Request)
	}

	if err != nil {
		// 判断是否是可重试的错误
		if isTemporaryError(err) {
			return &RetryableError{Err: err}
		}
		return err
	}

	return nil
}

// isTemporaryError checks if the error is temporary
func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	// 网络超时、连接重置等错误可以重试
	if netErr, ok := err.(net.Error); ok {
		return netErr.Temporary()
	}

	// Server-side errors (500 series) can be retried
	if httpErr, ok := err.(*types.HTTPErrorResponse); ok {
		return httpErr.StatusCode >= 500
	}

	return false
}

// ResponseHandler optimizes stream response processing
type ResponseHandler struct {
	BaseHandler
}

func (h *ResponseHandler) Handle(ctx *ServiceContext) error {
	if ctx.Response == nil {
		return fmt.Errorf("no response to handle")
	}

	defer ctx.Response.Body.Close()

	respStreamMode := NewStreamMode(ctx.Response.Header)
	if !respStreamMode.IsStream() {
		return h.handleNonStreamResponse(ctx)
	}
	return h.handleStreamResponse(ctx)
}

func (h *ResponseHandler) handleNonStreamResponse(ctx *ServiceContext) error {
	body, err := io.ReadAll(ctx.Response.Body)
	if err != nil {
		return fmt.Errorf("[Service] Failed to read response body: %s", err.Error())
	}

	content := types.HTTPContent{Body: body, Header: ctx.Response.Header.Clone()}

	requestFlavor, err := GetAPIFlavor(ctx.Task.Request.FromFlavor)
	if err != nil {
		return err
	}

	targetFlavor, err := GetAPIFlavor(ctx.Task.Target.ServiceProvider.Flavor)
	if err != nil {
		return err
	}

	if targetFlavor.Name() != requestFlavor.Name() {
		respConvertCtx := convert.ConvertContext{
			"id": fmt.Sprintf("%d%d", rand.Uint64(), ctx.Task.Schedule.Id),
		}
		content, err = ConvertBetweenFlavors(targetFlavor, requestFlavor, ctx.Task.Request.Service, "response", content, respConvertCtx)
		if err != nil {
			return fmt.Errorf("[Service] Failed to convert response: %s", err.Error())
		}
	}

	ctx.Task.Ch <- &types.ServiceResult{
		Type:       types.ServiceResultDone,
		TaskId:     ctx.Task.Schedule.Id,
		StatusCode: ctx.Response.StatusCode,
		HTTP:       content,
	}

	return nil
}

func (h *ResponseHandler) handleStreamResponse(ctx *ServiceContext) error {
	respStreamMode := NewStreamMode(ctx.Response.Header)
	reader := bufio.NewReader(ctx.Response.Body)

	requestFlavor, err := GetAPIFlavor(ctx.Task.Request.FromFlavor)
	if err != nil {
		return err
	}

	targetFlavor, err := GetAPIFlavor(ctx.Task.Target.ServiceProvider.Flavor)
	if err != nil {
		return err
	}

	streamProcessor := NewStreamProcessor(ctx, respStreamMode, requestFlavor, targetFlavor)
	return streamProcessor.Process(reader)
}

// StreamProcessor handles stream response processing
type StreamProcessor struct {
	ctx                 *ServiceContext
	streamMode          *types.StreamMode
	requestFlavor       APIFlavor
	targetFlavor        APIFlavor
	isFirstChunk        bool
	convertedStreamMode *types.StreamMode
}

func NewStreamProcessor(ctx *ServiceContext, mode *types.StreamMode, reqFlavor, targetFlavor APIFlavor) *StreamProcessor {
	return &StreamProcessor{
		ctx:           ctx,
		streamMode:    mode,
		requestFlavor: reqFlavor,
		targetFlavor:  targetFlavor,
		isFirstChunk:  true,
	}
}

func (sp *StreamProcessor) Process(reader *bufio.Reader) error {
	prolog := sp.requestFlavor.GetStreamResponseProlog(sp.ctx.Task.Request.Service)
	epilog := sp.requestFlavor.GetStreamResponseEpilog(sp.ctx.Task.Request.Service)

	for {
		chunk, err := sp.processChunk(reader)
		if err != nil {
			if err == io.EOF {
				return sp.handleEOF(chunk, epilog)
			}
			return err
		}

		if sp.isFirstChunk {
			if err := sp.sendProlog(prolog); err != nil {
				return err
			}
			sp.isFirstChunk = false
		}

		if err := sp.sendChunk(chunk); err != nil {
			return err
		}
	}
}

func (sp *StreamProcessor) processChunk(reader *bufio.Reader) ([]byte, error) {
	chunk, err := sp.streamMode.ReadChunk(reader)
	if err != nil && err != io.EOF {
		return nil, err
	}

	if len(bytes.TrimSpace(chunk)) == 0 {
		return nil, &types.DropAction{}
	}

	// 需要转换
	if sp.targetFlavor.Name() != sp.requestFlavor.Name() {
		content := types.HTTPContent{
			Body:   sp.streamMode.UnwrapChunk(chunk),
			Header: sp.ctx.Response.Header.Clone(),
		}

		convertCtx := convert.ConvertContext{
			"id": fmt.Sprintf("%d%d", rand.Uint64(), sp.ctx.Task.Schedule.Id),
		}

		content, err = ConvertBetweenFlavors(sp.targetFlavor, sp.requestFlavor,
			sp.ctx.Task.Request.Service, "stream_response", content, convertCtx)
		if err != nil {
			if types.IsDropAction(err) {
				return nil, err
			}
			return nil, fmt.Errorf("[Service] Failed to convert response: %s", err.Error())
		}

		// 初始化转换后的流模式
		if sp.convertedStreamMode == nil {
			sp.convertedStreamMode = NewStreamMode(content.Header)
		}

		chunk = sp.convertedStreamMode.WrapChunk(content.Body)
	}

	return chunk, err
}

func (sp *StreamProcessor) sendProlog(prolog []string) error {
	if len(prolog) == 0 {
		return nil
	}

	logger.LogicLogger.Info("[Service] Stream: Send Prolog",
		"taskid", sp.ctx.Task.Schedule.Id,
		"prolog", prolog)

	for i := len(prolog) - 1; i >= 0; i-- {
		content := types.HTTPContent{
			Body:   []byte(prolog[i]),
			Header: sp.getStreamHeader(),
		}
		if err := sp.sendResult(content, types.ServiceResultChunk); err != nil {
			return err
		}
	}
	return nil
}

func (sp *StreamProcessor) handleEOF(lastChunk []byte, epilog []string) error {
	if len(lastChunk) > 0 {
		content := types.HTTPContent{
			Body:   lastChunk,
			Header: sp.getStreamHeader(),
		}
		if err := sp.sendResult(content, types.ServiceResultChunk); err != nil {
			return err
		}
	}

	if len(epilog) > 0 {
		logger.LogicLogger.Info("[Service] Stream: Send Epilog",
			"taskid", sp.ctx.Task.Schedule.Id,
			"epilog", epilog)

		for _, line := range epilog {
			content := types.HTTPContent{
				Body:   []byte(line),
				Header: sp.getStreamHeader(),
			}
			if err := sp.sendResult(content, types.ServiceResultChunk); err != nil {
				return err
			}
		}
	}

	// 发送结束标记
	return sp.sendResult(types.HTTPContent{
		Header: sp.getStreamHeader(),
	}, types.ServiceResultDone)
}

func (sp *StreamProcessor) sendChunk(chunk []byte) error {
	return sp.sendResult(types.HTTPContent{
		Body:   chunk,
		Header: sp.getStreamHeader(),
	}, types.ServiceResultChunk)
}

func (sp *StreamProcessor) sendResult(content types.HTTPContent, resultType types.ServiceResultType) error {
	sp.ctx.Task.Ch <- &types.ServiceResult{
		Type:       resultType,
		TaskId:     sp.ctx.Task.Schedule.Id,
		StatusCode: sp.ctx.Response.StatusCode,
		HTTP:       content,
	}
	return nil
}

func (sp *StreamProcessor) getStreamHeader() http.Header {
	if sp.convertedStreamMode != nil {
		return sp.convertedStreamMode.Header
	}
	return sp.ctx.Response.Header.Clone()
}
