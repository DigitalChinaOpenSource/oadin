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
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"intel.com/aog/internal/client"
	"intel.com/aog/internal/client/grpc/grpc_client"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils"
)

// Invoker defines the interface for service invocation
type Invoker interface {
	Invoke(sp *types.ServiceProvider, content types.HTTPContent) (*http.Response, error)
}

// GRPCInvoker implements the invocation of GRPC services
type GRPCInvoker struct {
	task *ServiceTask
}

// HTTPInvoker implements the invocation of HTTP services
type HTTPInvoker struct {
	task *ServiceTask
}

// NewInvoker creates the corresponding invoker
func NewInvoker(task *ServiceTask, providerType string) Invoker {
	switch providerType {
	case types.ProtocolGRPC:
		return &GRPCInvoker{task: task}
	case types.ProtocolHTTP:
		return &HTTPInvoker{task: task}
	default:
		return nil
	}
}

// Invoke implements GRPC service invocation
func (g *GRPCInvoker) Invoke(sp *types.ServiceProvider, content types.HTTPContent) (*http.Response, error) {
	// Get the corresponding service handler
	handler := NewGRPCServiceHandler(sp.ServiceName)
	if handler == nil {
		return nil, fmt.Errorf("unsupported service type: %s", sp.ServiceName)
	}

	// 检查是否为流式服务
	if g.task.Target.Protocol == types.ProtocolGRPC_STREAM {
		logger.LogicLogger.Debug("[Service] Invoking GRPC streaming service",
			"service", sp.ServiceName,
			"model", g.task.Target.Model,
			"taskid", g.task.Schedule.Id)
		return g.invokeStreamService(handler, content)
	}

	// 非流式服务处理
	logger.LogicLogger.Debug("[Service] Invoking GRPC non-streaming service",
		"service", sp.ServiceName,
		"model", g.task.Target.Model,
		"taskid", g.task.Schedule.Id)
	return g.invokeNonStreamService(handler, content)
}

// 处理非流式服务
func (g *GRPCInvoker) invokeNonStreamService(handler GRPCServiceHandler, content types.HTTPContent) (*http.Response, error) {
	// Establish a gRPC connection
	conn, err := grpc.Dial(g.task.Target.ServiceProvider.URL, grpc.WithInsecure())
	if err != nil {
		logger.LogicLogger.Error("Couldn't connect to endpoint %s: %v", g.task.Target.ServiceProvider.URL, err)
		return nil, err
	}
	defer conn.Close()

	// Create a gRPC client
	client := grpc_client.NewGRPCInferenceServiceClient(conn)

	// Prepare the request
	grpcReq, err := handler.PrepareRequest(content, g.task.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	// Send the request
	inferResponse, err := client.ModelInfer(context.Background(), grpcReq)
	if err != nil {
		logger.LogicLogger.Error("[Service] Error processing InferRequest", "taskid", g.task.Schedule.Id, "error", err)
		return nil, err
	}

	// Process the response
	resp, err := handler.ProcessResponse(inferResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to process response: %w", err)
	}

	logger.LogicLogger.Debug("[Service] Response Receiving", "taskid", g.task.Schedule.Id, "header",
		fmt.Sprintf("%+v", resp.Header), "task", g.task)

	return resp, nil
}

// 处理流式服务
func (g *GRPCInvoker) invokeStreamService(handler GRPCServiceHandler, content types.HTTPContent) (*http.Response, error) {
	// 1. 解析WebSocket连接ID和任务类型
	wsConnID := content.Header.Get("X-WebSocket-ConnID")
	if wsConnID == "" {
		return nil, fmt.Errorf("missing WebSocket connection ID")
	}

	wsManager := client.GetGlobalWebSocketManager()
	wsConn, _ := wsManager.GetConnection(wsConnID)
	taskType := wsConn.GetTaskType(g.task.Schedule.Id)

	// 2. 根据任务类型分别处理
	switch taskType {
	case types.WSActionFinishTask:
		// 处理结束任务请求
		return g.handleFinishTask(handler, content, wsConnID)

	case types.WSActionRunTask:
		// 处理开始任务请求
		return g.handleRunTask(handler, content, wsConnID)

	default:
		// 处理音频数据或其他请求
		return g.handleStreamData(handler, content, wsConnID)
	}
}

// handleFinishTask 处理结束任务请求
func (g *GRPCInvoker) handleFinishTask(handler GRPCServiceHandler, content types.HTTPContent, wsConnID string) (*http.Response, error) {
	// 关闭与该WebSocket连接关联的gRPC流会话
	grpcStreamManager := client.GetGlobalGRPCStreamManager()
	session := grpcStreamManager.GetSessionByWSConnID(wsConnID)

	// 通知WebSocket管理器关闭流式会话
	wsManager := client.GetGlobalWebSocketManager()
	if wsConn, exists := wsManager.GetConnection(wsConnID); exists {
		wsConn.SetTaskFinished(g.task.Schedule.Id)
	}

	// 向ovms 发送空audio数据 代表发送完成
	grpcReq, err := handler.PrepareStreamRequest(content, g.task)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare stream request: %w", err)
	}

	err = session.Stream.Send(grpcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send empty audio data: %w", err)
	}

	// 成功发送数据，立即返回确认
	ackResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"status":"processing"}`))),
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}
	return ackResponse, nil
}

// handleRunTask 处理开始任务请求
func (g *GRPCInvoker) handleRunTask(handler GRPCServiceHandler, content types.HTTPContent, wsConnID string) (*http.Response, error) {
	// 创建新的gRPC流会话
	_, err := g.createStreamSession(handler, content, wsConnID)
	if err != nil {
		return nil, err
	}

	// 创建任务开始事件响应
	startedEvent := types.NewTaskStartedEvent(wsConnID)
	return createWebSocketEventResponse(startedEvent), nil
}

// handleStreamData 处理音频数据或其他流式请求
func (g *GRPCInvoker) handleStreamData(handler GRPCServiceHandler, content types.HTTPContent, wsConnID string) (*http.Response, error) {
	// 尝试获取现有gRPC流会话
	grpcStreamManager := client.GetGlobalGRPCStreamManager()
	session := grpcStreamManager.GetSessionByWSConnID(wsConnID)

	// 如果没有找到会话，创建新会话
	if session == nil {
		var err error
		session, err = g.createStreamSession(handler, content, wsConnID)
		if err != nil {
			return nil, err
		}

		// 立即返回一个启动成功响应
		startEvent := types.NewTaskStartedEvent(wsConnID)
		return createWebSocketEventResponse(startEvent), nil
	} else {
		// 准备现有会话的请求
		grpcReq, err := handler.PrepareStreamRequest(content, g.task)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare stream request: %w", err)
		}

		// 发送到现有流
		if err := session.Stream.Send(grpcReq); err != nil {
			grpcStreamManager.CloseSessionByWSConnID(wsConnID)
			logger.LogicLogger.Error("[Service] Error sending to existing stream",
				"taskid", g.task.Schedule.Id,
				"wsConnID", wsConnID,
				"error", err)
			return nil, err
		}

		// 成功发送数据，立即返回确认
		ackResponse := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"status":"processing"}`))),
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		}
		return ackResponse, nil
	}
}

// createStreamSession 创建新的gRPC流会话
func (g *GRPCInvoker) createStreamSession(handler GRPCServiceHandler, content types.HTTPContent, wsConnID string) (*client.GRPCStreamSession, error) {
	logger.LogicLogger.Info("[Service] Creating new GRPC stream for WebSocket connection",
		"wsConnID", wsConnID,
		"taskID", g.task.Schedule.Id)

	// 建立gRPC连接
	conn, err := grpc.Dial(g.task.Target.ServiceProvider.URL, grpc.WithInsecure())
	if err != nil {
		logger.LogicLogger.Error("[Service] Couldn't connect to endpoint",
			"url", g.task.Target.ServiceProvider.URL,
			"error", err)
		return nil, err
	}

	// 创建gRPC客户端
	gClient := grpc_client.NewGRPCInferenceServiceClient(conn)

	// 准备流式请求
	grpcReq, err := handler.PrepareStreamRequest(content, g.task)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare stream request: %w", err)
	}

	// 创建上下文和取消函数
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)

	// 创建双向流
	stream, err := gClient.ModelStreamInfer(ctx)
	if err != nil {
		cancel()
		logger.LogicLogger.Error("[Service] Error creating stream",
			"taskid", g.task.Schedule.Id,
			"wsConnID", wsConnID,
			"error", err)
		return nil, err
	}

	// 创建新会话
	grpcStreamManager := client.GetGlobalGRPCStreamManager()
	session := grpcStreamManager.CreateSession(
		wsConnID,
		gClient,
		stream,
		ctx,
		cancel,
		g.task.Request.Service,
		g.task.Target.Model,
	)

	// 发送初始请求
	if err := stream.Send(grpcReq); err != nil {
		grpcStreamManager.CloseSessionByWSConnID(wsConnID)
		logger.LogicLogger.Error("[Service] Error sending stream request",
			"taskid", g.task.Schedule.Id,
			"wsConnID", wsConnID,
			"error", err)
		return nil, err
	}

	return session, nil
}

// createWebSocketEventResponse 创建一个WebSocket事件响应
func createWebSocketEventResponse(event types.WebSocketEventMessage) *http.Response {
	jsonData, _ := json.Marshal(event)
	header := make(http.Header)
	header.Set("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       io.NopCloser(bytes.NewReader(jsonData)),
	}
}

// Invoke implements HTTP service invocation
func (h *HTTPInvoker) Invoke(sp *types.ServiceProvider, content types.HTTPContent) (*http.Response, error) {
	// 1. Build the invocation URL
	invokeURL := sp.URL
	if strings.ToUpper(sp.Method) == http.MethodGet && len(content.Body) > 0 {
		u, err := utils.BuildGetRequestURL(sp.URL, content.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to build GET request URL: %w", err)
		}
		invokeURL = u
		content.Body = nil
	}

	// 2. Create the HTTP request
	req, err := http.NewRequest(sp.Method, invokeURL, bytes.NewReader(content.Body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 3. Set the request headers
	if err := h.setRequestHeaders(req, sp, content); err != nil {
		return nil, fmt.Errorf("failed to set request headers: %w", err)
	}

	// 4. Handle authentication
	if sp.AuthType != types.AuthTypeNone {
		if err := h.handleAuthentication(req, sp, content); err != nil {
			return nil, fmt.Errorf("failed to handle authentication: %w", err)
		}
	}

	// 5. Create HTTP client and send request
	client := h.createHTTPClient()

	// 6. Log the request
	h.logRequest(req, content)

	// 7. Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// 8. Handle error response
	if resp.StatusCode != http.StatusOK {
		return h.handleErrorResponse(resp)
	}

	// 9. Handle segmented request
	serviceDefaultInfo := GetProviderServiceDefaultInfo(h.task.Target.ToFavor, h.task.Request.Service)
	if serviceDefaultInfo.RequestSegments > 1 {
		return h.handleSegmentedRequest(client, resp, sp, serviceDefaultInfo)
	}

	// 10. Log the response
	logger.LogicLogger.Debug("[Service] Response Receiving",
		"taskid", h.task.Schedule.Id,
		"header", fmt.Sprintf("%+v", resp.Header),
		"task", h.task)

	return resp, nil
}

// handleAuthentication Handle authentication
func (h *HTTPInvoker) handleAuthentication(req *http.Request, sp *types.ServiceProvider, content types.HTTPContent) error {
	authParams := &AuthenticatorParams{
		Request:      req,
		ProviderInfo: sp,
		RequestBody:  string(content.Body),
	}

	authenticator := ChooseProviderAuthenticator(authParams)
	if authenticator == nil {
		logger.LogicLogger.Error("[Service] Failed to choose authenticator")
		return fmt.Errorf("failed to choose authenticator")
	}

	if err := authenticator.Authenticate(); err != nil {
		logger.LogicLogger.Error("[Service] Failed to authenticate",
			"taskid", h.task.Schedule.Id,
			"error", err)
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	return nil
}

func (h *HTTPInvoker) handleSegmentedRequest(client *http.Client, resp *http.Response, sp *types.ServiceProvider, serviceDefaultInfo ServiceDefaultInfo) (*http.Response, error) {
	defer resp.Body.Close()

	// 1. Read and parse initial response
	taskID, err := h.parseInitialResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse initial response: %w", err)
	}

	// 2. Poll task status
	return h.pollTaskStatus(client, sp, serviceDefaultInfo, taskID)
}

func (h *HTTPInvoker) parseInitialResponse(resp *http.Response) (string, error) {
	body, err := h.readResponseBody(resp)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var submitRespData struct {
		Output struct {
			TaskId string `json:"task_id"`
		} `json:"output"`
	}

	if err := json.Unmarshal(body, &submitRespData); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return submitRespData.Output.TaskId, nil
}

func (h *HTTPInvoker) readResponseBody(resp *http.Response) ([]byte, error) {
	var reader io.ReadCloser
	var err error

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	return io.ReadAll(reader)
}

// pollTaskStatus Poll the task status
func (h *HTTPInvoker) pollTaskStatus(client *http.Client, sp *types.ServiceProvider, serviceDefaultInfo ServiceDefaultInfo, taskID string) (*http.Response, error) {
	const (
		pollInterval = 500 * time.Millisecond
		maxRetries   = 100 // Add maximum retry count to prevent infinite loops
	)

	retryCount := 0
	for retryCount < maxRetries {
		resp, err := h.getTaskStatus(client, sp, serviceDefaultInfo, taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task status: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return h.handleErrorStatusResponse(resp)
		}

		status, body, err := h.parseTaskStatusResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse task status response: %w", err)
		}

		if h.isTaskComplete(status) {
			return &http.Response{
				StatusCode: resp.StatusCode,
				Header:     resp.Header.Clone(),
				Body:       io.NopCloser(bytes.NewReader(body)),
			}, nil
		}

		time.Sleep(pollInterval)
		retryCount++
	}

	return nil, fmt.Errorf("exceeded maximum number of retries (%d) while polling task status", maxRetries)
}

// getTaskStatus Get the task status
func (h *HTTPInvoker) getTaskStatus(client *http.Client, sp *types.ServiceProvider, serviceDefaultInfo ServiceDefaultInfo, taskID string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", serviceDefaultInfo.RequestExtraUrl, taskID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create status request: %w", err)
	}

	if err := h.authenticateStatusRequest(req, sp); err != nil {
		return nil, fmt.Errorf("failed to authenticate status request: %w", err)
	}

	return client.Do(req)
}

// authenticateStatusRequest Authenticate the status request
func (h *HTTPInvoker) authenticateStatusRequest(req *http.Request, sp *types.ServiceProvider) error {
	authParams := &AuthenticatorParams{
		Request:      req,
		ProviderInfo: sp,
	}

	authenticator := ChooseProviderAuthenticator(authParams)
	if authenticator == nil {
		return fmt.Errorf("failed to choose authenticator for status request")
	}

	return authenticator.Authenticate()
}

// handleErrorStatusResponse Handle error status responses
func (h *HTTPInvoker) handleErrorStatusResponse(resp *http.Response) (*http.Response, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read error response body: %w", err)
	}

	logger.LogicLogger.Warn("[Service] Service Provider returns Error",
		"taskid", h.task.Schedule.Id,
		"status_code", resp.StatusCode,
		"body", string(body))

	return nil, &types.HTTPErrorResponse{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       body,
	}
}

// parseTaskStatusResponse Parse the task status response
func (h *HTTPInvoker) parseTaskStatusResponse(resp *http.Response) (string, []byte, error) {
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var respData struct {
		Output struct {
			TaskStatus string `json:"task_status"`
		} `json:"output"`
	}

	if err := json.Unmarshal(body, &respData); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return respData.Output.TaskStatus, body, nil
}

// isTaskComplete Check if the task is complete
func (h *HTTPInvoker) isTaskComplete(status string) bool {
	return status == "FAILED" || status == "SUCCEEDED" || status == "UNKNOWN"
}

// logRequest Log the request
func (h *HTTPInvoker) logRequest(req *http.Request, content types.HTTPContent) {
	logger.LogicLogger.Info("[Service] Request Sending to Service Provider ...",
		"taskid", h.task.Schedule.Id,
		"url", req.URL.String())

	logger.LogicLogger.Debug("[Service] Request Sending to Service Provider ...",
		"taskid", h.task.Schedule.Id,
		"method", req.Method,
		"url", req.URL.String(),
		"header", fmt.Sprintf("%+v", req.Header),
		"body", string(content.Body))
}

func (h *HTTPInvoker) handleErrorResponse(resp *http.Response) (*http.Response, error) {
	var sbody string
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		sbody = string(b)
	}
	logger.LogicLogger.Warn("[Service] Service Provider returns Error", "taskid", h.task.Schedule.Id,
		"status_code", resp.StatusCode, "body", sbody)
	resp.Body.Close()
	return nil, errors.New("[Service] Service Provider API returns Error err: \n" + sbody)
}

// createHTTPClient Create an HTTP client
func (h *HTTPInvoker) createHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		},
	}
}

// setRequestHeaders Set request headers
func (h *HTTPInvoker) setRequestHeaders(req *http.Request, sp *types.ServiceProvider, content types.HTTPContent) error {
	// set content headers
	for k, v := range content.Header {
		if k != "Content-Length" {
			req.Header.Set(k, v[0])
		}
	}

	// set extra headers
	if sp.ExtraHeaders != "{}" {
		var extraHeader map[string]interface{}
		if err := json.Unmarshal([]byte(sp.ExtraHeaders), &extraHeader); err != nil {
			logger.LogicLogger.Error("Error parsing extra headers:", err)
			return fmt.Errorf("failed to parse extra headers: %w", err)
		}
		for k, v := range extraHeader {
			if strVal, ok := v.(string); ok {
				req.Header.Set(k, strVal)
			}
		}
	}
	return nil
}

// GRPCServiceHandler define the interface for handling GRPC service requests
type GRPCServiceHandler interface {
	PrepareRequest(content types.HTTPContent, target *types.ServiceTarget) (*grpc_client.ModelInferRequest, error)
	ProcessResponse(inferResponse *grpc_client.ModelInferResponse) (*http.Response, error)

	// 流式请求处理方法
	PrepareStreamRequest(content types.HTTPContent, st *ServiceTask) (*grpc_client.ModelInferRequest, error)
	ProcessStreamResponse(streamResponse *grpc_client.ModelStreamInferResponse, wsConnID string) (*http.Response, bool, error)
}

type BaseGRPCHandler struct{}

// 默认的流式请求处理方法，返回未实现错误
func (h *BaseGRPCHandler) PrepareStreamRequest(content types.HTTPContent, st *ServiceTask) (*grpc_client.ModelInferRequest, error) {
	return nil, fmt.Errorf("streaming not implemented")
}

// 默认的流式响应处理方法，返回未实现错误
// 返回值：HTTP响应，是否继续处理流，错误
func (h *BaseGRPCHandler) ProcessStreamResponse(streamResponse *grpc_client.ModelStreamInferResponse, wsConnID string) (*http.Response, bool, error) {
	return nil, true, fmt.Errorf("streaming not implemented")
}

// TextToImageHandler handles text-to-image service
type TextToImageHandler struct {
	BaseGRPCHandler
}

func NewTextToImageHandler() *TextToImageHandler {
	return &TextToImageHandler{
		BaseGRPCHandler: BaseGRPCHandler{},
	}
}

// SpeechToTextHandler handles speech-to-text service
type SpeechToTextHandler struct {
	BaseGRPCHandler
}

func NewSpeechToTextHandler() *SpeechToTextHandler {
	return &SpeechToTextHandler{
		BaseGRPCHandler: BaseGRPCHandler{},
	}
}

type SpeechToTextWSHandler struct {
	BaseGRPCHandler
}

func NewSpeechToTextWSHandler() *SpeechToTextWSHandler {
	return &SpeechToTextWSHandler{
		BaseGRPCHandler: BaseGRPCHandler{},
	}
}

// PrepareRequest 实现非流式请求准备
func (h *SpeechToTextWSHandler) PrepareRequest(content types.HTTPContent, target *types.ServiceTarget) (*grpc_client.ModelInferRequest, error) {
	// 流式处理器不支持非流式请求
	return nil, fmt.Errorf("speech-to-text stream handler only supports streaming")
}

// ProcessResponse 实现非流式响应处理
func (h *SpeechToTextWSHandler) ProcessResponse(inferResponse *grpc_client.ModelInferResponse) (*http.Response, error) {
	// 流式处理器不支持非流式响应
	return nil, fmt.Errorf("speech-to-text stream handler only supports streaming")
}

// PrepareStreamRequest 准备流式请求
func (h *SpeechToTextWSHandler) PrepareStreamRequest(content types.HTTPContent, st *ServiceTask) (*grpc_client.ModelInferRequest, error) {
	var actionMessage types.WebSocketActionMessage

	// 首先尝试从WebSocket连接中获取任务类型
	wsConnID := content.Header.Get("X-WebSocket-ConnID")
	wsManager := client.GetGlobalWebSocketManager()
	wsConn, _ := wsManager.GetConnection(wsConnID)
	taskType := wsConn.GetTaskType(st.Schedule.Id)

	isBinary := taskType == types.WSTaskTypeAudio
	sttParams := wsConn.GetSTTParams()
	actionMessage.Parameters = &types.WebSocketParameters{
		Service:    st.Target.ServiceProvider.ServiceName,
		Format:     sttParams.AudioFormat,
		SampleRate: sttParams.SampleRate,
		Language:   sttParams.Language,
		UseVAD:     sttParams.UseVAD,
	}
	actionMessage.Model = st.Target.Model
	actionMessage.Action = taskType
	actionMessage.Task = st.Request.Service
	actionMessage.TaskID = wsConnID

	// 准备参数对象
	params := h.prepareParams(content, isBinary, taskType, actionMessage)
	// 获取音频数据
	audioBytes := h.getAudioData(content, isBinary, taskType, actionMessage)
	// 构建并返回模型推理请求
	return h.buildModelInferRequest(st.Target.Model, audioBytes, params), nil
}

// prepareParams 准备所有参数
func (h *SpeechToTextWSHandler) prepareParams(content types.HTTPContent, isBinary bool, taskType string, actionMessage types.WebSocketActionMessage) *types.SpeechToTextParams {
	params := types.NewSpeechToTextParams()
	if actionMessage.Parameters.Format != "" {
		params.AudioFormat = actionMessage.Parameters.Format
	}
	if actionMessage.Parameters.SampleRate > 0 {
		params.SampleRate = actionMessage.Parameters.SampleRate
	}
	if actionMessage.Parameters.Language != "" {
		params.Language = actionMessage.Parameters.Language
	}
	params.UseVAD = actionMessage.Parameters.UseVAD
	if actionMessage.Parameters.ReturnFormat != "" {
		params.ReturnFormat = actionMessage.Parameters.ReturnFormat
	}
	if actionMessage.TaskID != "" {
		params.TaskID = actionMessage.TaskID
	}
	if actionMessage.Model != "" {
		params.Model = actionMessage.Model
	}
	if actionMessage.Parameters.Service != "" {
		params.Service = actionMessage.Parameters.Service
	}

	return params
}

func (h *SpeechToTextWSHandler) ProcessStreamResponse(streamResponse *grpc_client.ModelStreamInferResponse, wsConnID string) (*http.Response, bool, error) {
	// 检查响应是否包含错误信息
	if streamResponse.ErrorMessage != "" {
		return nil, false, fmt.Errorf("stream error from model: %s", streamResponse.ErrorMessage)
	}

	// 检查是否有推理响应
	inferResponse := streamResponse.GetInferResponse()
	if inferResponse == nil || len(inferResponse.RawOutputContents) == 0 {
		return nil, true, nil // 返回nil响应，但继续接收流
	}

	// 解析SRT格式的文本
	srtText := string(inferResponse.RawOutputContents[0])
	lines := strings.Split(strings.TrimSpace(srtText), "\n")

	// 确保有足够的行来获取有效数据
	if len(lines) < 3 {
		return nil, true, nil // 数据不完整，继续接收流
	}

	// 解析ID
	id := 0
	fmt.Sscanf(lines[0], "%d", &id)

	// 解析时间戳
	timeRegex := regexp.MustCompile(`(\d{2}:\d{2}:\d{2},\d{3}) --> (\d{2}:\d{2}:\d{2},\d{3})`)
	matches := timeRegex.FindStringSubmatch(lines[1])
	if len(matches) != 3 {
		return nil, true, nil // 时间戳格式不正确，继续接收流
	}

	// 获取文本内容
	text := strings.Join(lines[2:], " ")

	// 构建单个段落响应（非数组）
	segment := map[string]interface{}{
		"id":    id,
		"start": matches[1],
		"end":   matches[2],
		"text":  text,
	}

	// 创建自定义响应消息
	resultMsg := map[string]interface{}{
		"header": map[string]string{
			"task_id": wsConnID,
			"event":   types.WSEventResultGenerated,
		},
		"payload": segment,
	}

	// 序列化响应
	jsonData, err := json.Marshal(resultMsg)
	if err != nil {
		return nil, false, fmt.Errorf("failed to marshal response to JSON: %v", err)
	}

	// 创建HTTP响应
	header := make(http.Header)
	header.Set("Content-Type", "application/json")

	// 确定是否继续接收流（根据是否有EOS标记）
	continueStream := true
	if inferResponse.Parameters != nil {
		if eosParam, exists := inferResponse.Parameters["eos"]; exists {
			if eosParam.GetBoolParam() {
				continueStream = false // 如果收到EOS标记，不再继续接收
			}
		}
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       io.NopCloser(bytes.NewReader(jsonData)),
	}, continueStream, nil
}

// getAudioData 获取音频数据
func (h *SpeechToTextWSHandler) getAudioData(content types.HTTPContent, isBinary bool, taskType string, actionMessage types.WebSocketActionMessage) []byte {
	// 处理二进制音频数据
	if isBinary {
		logger.LogicLogger.Debug("[Service] Processing binary audio data",
			"contentType", content.Header.Get("Content-Type"),
			"dataSize", len(content.Body))
		return content.Body
	}

	return []byte{}
}

// buildModelInferRequest 构建模型推理请求
func (h *SpeechToTextWSHandler) buildModelInferRequest(modelName string, audioBytes []byte, params *types.SpeechToTextParams) *grpc_client.ModelInferRequest {
	// 将参数序列化为JSON字符串
	paramsJSON, err := params.ToJSON()
	if err != nil {
		logger.LogicLogger.Error("[Service] Failed to marshal params to JSON", "error", err)
		paramsJSON = []byte("{}")
	}

	// 准备输入数据
	rawContents := [][]byte{
		audioBytes,
		paramsJSON,
	}

	// 准备输入张量
	inferTensorInputs := []*grpc_client.ModelInferRequest_InferInputTensor{
		{
			Name:     "audio",
			Datatype: "BYTES",
			Shape:    []int64{1},
		},
		{
			Name:     "params",
			Datatype: "BYTES",
			Shape:    []int64{1},
		},
	}

	// 准备输出张量
	inferOutputs := []*grpc_client.ModelInferRequest_InferRequestedOutputTensor{
		{
			Name: "result",
		},
	}

	return &grpc_client.ModelInferRequest{
		ModelName:        modelName,
		Inputs:           inferTensorInputs,
		Outputs:          inferOutputs,
		RawInputContents: rawContents,
	}
}

func (h *TextToImageHandler) PrepareRequest(content types.HTTPContent, target *types.ServiceTarget) (*grpc_client.ModelInferRequest, error) {
	var requestMap map[string]interface{}
	if err := json.Unmarshal(content.Body, &requestMap); err != nil {
		logger.LogicLogger.Error("[Service] Failed to unmarshal request body", "error", err)
		return nil, err
	}

	// Parse parameters
	prompt, ok := requestMap["prompt"].(string)
	if !ok {
		logger.LogicLogger.Error("[Service] Failed to get prompt from request body")
		return nil, fmt.Errorf("failed to get prompt from request body")
	}

	batch, height, width := h.parseImageParams(requestMap)

	// Prepare input data
	rawContents := [][]byte{
		[]byte(prompt),
		[]byte(fmt.Sprintf("%d", int(batch))),
		[]byte(strconv.Itoa(height)),
		[]byte(strconv.Itoa(width)),
	}

	// Prepare input tensors
	inferTensorInputs := []*grpc_client.ModelInferRequest_InferInputTensor{
		{
			Name:     "prompt",
			Datatype: "BYTES",
			Shape:    []int64{1},
		},
		{
			Name:     "batch",
			Datatype: "BYTES",
			Shape:    []int64{1},
		},
		{
			Name:     "height",
			Datatype: "BYTES",
		},
		{
			Name:     "width",
			Datatype: "BYTES",
		},
	}

	// Prepare output tensors
	inferOutputs := []*grpc_client.ModelInferRequest_InferRequestedOutputTensor{
		{
			Name: "image",
		},
	}

	return &grpc_client.ModelInferRequest{
		ModelName:        target.Model,
		Inputs:           inferTensorInputs,
		Outputs:          inferOutputs,
		RawInputContents: rawContents,
	}, nil
}

func (h *TextToImageHandler) parseImageParams(requestMap map[string]interface{}) (batch float64, height, width int) {
	batch = 1
	height = 1024
	width = 1024

	if batchVal, ok := requestMap["batch"].(float64); ok {
		batch = batchVal
	}

	if size, ok := requestMap["size"].(string); ok {
		sizeStr := strings.Split(size, "x")
		if len(sizeStr) == 2 {
			if num, err := strconv.Atoi(sizeStr[0]); err == nil {
				height = num
			}
			if num, err := strconv.Atoi(sizeStr[1]); err == nil {
				width = num
			}
		}
	}

	return batch, height, width
}

func (h *TextToImageHandler) ProcessResponse(inferResponse *grpc_client.ModelInferResponse) (*http.Response, error) {
	imageList, err := utils.ParseImageData(inferResponse.RawOutputContents[0])
	if err != nil {
		logger.LogicLogger.Error("[Service] Failed to parse image data", "error", err)
		return nil, err
	}

	outputList := make([]string, 0)
	for i, imageData := range imageList {
		imagePath, err := h.saveImage(imageData, i)
		if err != nil {
			logger.LogicLogger.Error("[Service] Failed to write image file", "error", err)
			continue
		}
		outputList = append(outputList, imagePath)
	}

	return h.createResponse(outputList)
}

func (h *TextToImageHandler) saveImage(imageData []byte, index int) (string, error) {
	now := time.Now()
	randNum := rand.Intn(10000)
	downloadPath, _ := utils.GetDownloadDir()
	imageName := fmt.Sprintf("%d%02d%02d%02d%02d%02d%04d%01d.png",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		randNum, index)
	imagePath := fmt.Sprintf("%s/%s", downloadPath, imageName)

	if err := os.WriteFile(imagePath, imageData, 0o644); err != nil {
		return "", err
	}
	return imagePath, nil
}

func (h *TextToImageHandler) createResponse(outputList []string) (*http.Response, error) {
	respHeader := make(http.Header)
	respHeader.Set("Content-Type", "application/json")

	respBody := map[string]interface{}{
		"local_path": outputList,
	}
	respBodyBytes, err := json.Marshal(respBody)
	if err != nil {
		return nil, err
	}

	return &http.Response{
		Header: respHeader,
		Body:   io.NopCloser(strings.NewReader(string(respBodyBytes))),
	}, nil
}

func (h *SpeechToTextHandler) PrepareRequest(content types.HTTPContent, target *types.ServiceTarget) (*grpc_client.ModelInferRequest, error) {
	var requestMap map[string]interface{}
	if err := json.Unmarshal(content.Body, &requestMap); err != nil {
		logger.LogicLogger.Error("[Service] Failed to unmarshal request body", "error", err)
		return nil, err
	}

	// 解析音频文件路径
	audioPath, ok := requestMap["audio"].(string)
	if !ok {
		logger.LogicLogger.Error("[Service] Failed to get audio path from request body")
		return nil, fmt.Errorf("failed to get audio path from request body")
	}

	// 解析语言参数（可选）
	language := "<|zh|>" // 默认中文
	if lang, ok := requestMap["language"].(string); ok {
		language = fmt.Sprintf("<|%s|>", lang)
	}

	params := &types.SpeechToTextParams{
		Service:      target.ServiceProvider.ServiceName,
		Language:     language,
		ReturnFormat: "text",
	}

	// 将参数序列化为JSON字符串
	paramsJSON, err := params.ToJSON()
	if err != nil {
		logger.LogicLogger.Error("[Service] Failed to marshal params to JSON", "error", err)
		paramsJSON = []byte("{}")
	}

	// 读取音频文件
	audioBytes, err := os.ReadFile(audioPath)
	if err != nil {
		logger.LogicLogger.Error("[Service] Failed to read audio file", "error", err)
		return nil, err
	}

	// 准备输入数据
	rawContents := [][]byte{
		audioBytes,
		paramsJSON,
	}

	// 准备输入张量
	inferTensorInputs := []*grpc_client.ModelInferRequest_InferInputTensor{
		{
			Name:     "audio",
			Datatype: "BYTES",
			Shape:    []int64{1},
		},
		{
			Name:     "params",
			Datatype: "BYTES",
			Shape:    []int64{1},
		},
	}

	// 准备输出张量
	inferOutputs := []*grpc_client.ModelInferRequest_InferRequestedOutputTensor{
		{
			Name: "result",
		},
	}

	return &grpc_client.ModelInferRequest{
		ModelName:        target.Model,
		Inputs:           inferTensorInputs,
		Outputs:          inferOutputs,
		RawInputContents: rawContents,
	}, nil
}

func (h *SpeechToTextHandler) ProcessResponse(inferResponse *grpc_client.ModelInferResponse) (*http.Response, error) {
	if len(inferResponse.RawOutputContents) == 0 {
		return nil, fmt.Errorf("empty response from model")
	}

	// 解析时间戳格式的文本
	srtText := string(inferResponse.RawOutputContents[0])
	lines := strings.Split(strings.TrimSpace(srtText), "\n")

	segments := make([]map[string]interface{}, 0, len(lines))
	// 匹配格式: [开始时间, 结束时间] 文本内容
	timeRegex := regexp.MustCompile(`^\[(\d+\.\d+),\s*(\d+\.\d+)\]\s*(.+)$`)

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析时间戳和文本
		matches := timeRegex.FindStringSubmatch(line)
		if len(matches) != 4 {
			continue
		}

		startTime := matches[1]
		endTime := matches[2]
		text := strings.TrimSpace(matches[3])

		// 将秒数转换为时:分:秒,毫秒格式
		startTimeFormatted := utils.FormatSecondsToSRT(startTime)
		endTimeFormatted := utils.FormatSecondsToSRT(endTime)

		segment := map[string]interface{}{
			"id":    i + 1,
			"start": startTimeFormatted,
			"end":   endTimeFormatted,
			"text":  text,
		}

		segments = append(segments, segment)
	}

	// 构建JSON响应
	response := map[string]interface{}{
		"segments": segments,
	}

	// 转换为JSON
	jsonData, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response to JSON: %v", err)
	}

	// 创建HTTP响应
	header := make(http.Header)
	header.Set("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       io.NopCloser(bytes.NewReader(jsonData)),
	}, nil
}

func NewGRPCServiceHandler(serviceType string) GRPCServiceHandler {
	switch serviceType {
	case types.ServiceTextToImage:
		return NewTextToImageHandler()
	case types.ServiceSpeechToText:
		return NewSpeechToTextHandler()
	case types.ServiceSpeechToTextWS:
		return NewSpeechToTextWSHandler()
	default:
		return nil
	}
}
