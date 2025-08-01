package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"oadin/extension/api/dto"
	"oadin/extension/server"
	"oadin/internal/datastore"
	"oadin/internal/logger"
	"oadin/internal/utils/bcode"

	"github.com/gin-gonic/gin"
)

type PlaygroundApi struct {
	Playground server.Playground
	MCP        server.MCPService
	DataStore  datastore.Datastore
}

func NewPlaygroundApi() *PlaygroundApi {
	return &PlaygroundApi{
		Playground: server.NewPlayground(),
		MCP:        server.NewMcpService(),
		DataStore:  datastore.GetDefaultDatastore(),
	}
}

func (e *PlaygroundApi) InjectRoutes(api *gin.RouterGroup) {
	api.POST("/session", e.CreateSession)
	api.GET("/sessions", e.GetSessions)
	api.DELETE("/session", e.DeleteSession)
	api.POST("/message", e.SendMessage)
	api.POST("/message/stream", e.SendMessageStream)
	api.POST("/session/genTitle", e.GenSessionTitle)
	api.GET("/messages", e.GetMessages)
	api.POST("/session/model", e.ChangeSessionModel)
	api.POST("/session/thinking", e.ToggleSessionThinking)
	api.POST("/file", e.UploadFile)
	api.GET("/files", e.GetFiles)
	api.DELETE("/file", e.DeleteFile)
	api.POST("/file/process", e.ProcessFile)
}

// 创建会话
func (e *PlaygroundApi) CreateSession(c *gin.Context) {
	var req dto.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := e.Playground.CreateSession(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 获取会话列表
func (e *PlaygroundApi) GetSessions(c *gin.Context) {
	resp, err := e.Playground.GetSessions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 删除会话
func (e *PlaygroundApi) DeleteSession(c *gin.Context) {
	var req dto.DeleteSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := e.Playground.DeleteSession(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 发送消�?
func (e *PlaygroundApi) SendMessage(c *gin.Context) {
	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.McpIds) > 0 {
		req.Tools = make([]dto.Tool, 0)
		for _, id := range req.McpIds {
			tools, err := e.MCP.ClientGetTools(c, id)
			if err != nil {
				continue
			}

			newTools := make([]dto.Tool, 0, len(tools))
			for _, tool := range tools {
				newTools = append(newTools, dto.Tool{Type: "function", Function: dto.TypeFunction{Name: tool.Name, Description: tool.Description, Parameters: tool.InputSchema}})
			}
			req.Tools = append(req.Tools, newTools...)
		}
	}
	resp, err := e.Playground.SendMessage(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 发送消息并流式返回响应
func (e *PlaygroundApi) SendMessageStream(c *gin.Context) {
	var req dto.SendStreamMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.McpIds) > 0 {
		req.Tools = make([]dto.Tool, 0)
		for _, id := range req.McpIds {
			tools, err := e.MCP.ClientGetTools(c, id)
			if err != nil {
				continue
			}

			newTools := make([]dto.Tool, 0, len(tools))
			for _, tool := range tools {
				newTools = append(newTools, dto.Tool{Type: "function", Function: dto.TypeFunction{Name: tool.Name, Description: tool.Description, Parameters: tool.InputSchema}})
			}
			req.Tools = append(req.Tools, newTools...)
		}
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	ctx := c.Request.Context()

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Printf("[API] Flusher不支持\n")
		http.NotFound(w, c.Request)
		return
	}

	// 开始流式处�?
	respChan, errChan := e.Playground.SendMessageStream(ctx, &req)

	for {
		select {
		case chunk, ok := <-respChan:
			if !ok {
				return // 流结�?
			}

			if chunk.Type == "error" {
				response := dto.StreamMessageResponse{
					Bcode: bcode.ErrServer,
					Data: dto.MessageChunk{
						ID:        chunk.ID,
						SessionID: req.SessionID,
						Content:   chunk.Content,
						Type:      "error",
					},
				}
				data, err := json.Marshal(response)
				if err == nil {
					n, err := fmt.Fprintf(w, "data: %s\n\n", data)
					fmt.Printf("[API] 错误消息写入结果: 字节�?%d, 错误=%v\n", n, err)
					flusher.Flush()
				}
				return
			}

			// 转换 ToolCalls 类型
			var dtoToolCalls []dto.ToolCall
			for _, tc := range chunk.ToolCalls {
				dtoToolCalls = append(dtoToolCalls, dto.ToolCall{
					Function: struct {
						Name      string                 `json:"name"`
						Arguments map[string]interface{} `json:"arguments"`
					}{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				})
			}

			// issues 84 如果是工具调用，Content会偶发出�?n，会导致前端出现空白�?
			if chunk.ToolCalls != nil && chunk.Content != "" {
				chunk.Content = ""
			}

			response := dto.StreamMessageResponse{
				Bcode: bcode.SuccessCode,
				Data: dto.MessageChunk{
					ID:            chunk.ID,
					SessionID:     req.SessionID,
					Content:       chunk.Content,
					IsComplete:    chunk.IsComplete,
					Thoughts:      chunk.Thoughts,
					Type:          chunk.Type,
					ToolCalls:     dtoToolCalls,
					TotalDuration: chunk.TotalDuration,
					ToolGroupID:   chunk.ToolGroupID,
				},
			}
			data, err := json.Marshal(response)
			if err != nil {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			if chunk.IsComplete {
				return
			}

		case err, ok := <-errChan:
			if !ok || err == nil {
				return
			}
			flusher.Flush()
			return
		}
	}
}

func (e *PlaygroundApi) GenSessionTitle(c *gin.Context) {
	var req struct {
		SessionID string `json:"sessionId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "400", "message": err.Error()})
		return
	}
	err := e.Playground.UpdateSessionTitle(c.Request.Context(), req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "404", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "message": "success"})
}

// 获取消息
func (e *PlaygroundApi) GetMessages(c *gin.Context) {
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	req := &dto.GetMessagesRequest{
		SessionId: sessionID,
	}

	resp, err := e.Playground.GetMessages(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// 切换会话模型
func (e *PlaygroundApi) ChangeSessionModel(c *gin.Context) {
	var req dto.ChangeSessionModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := e.Playground.ChangeSessionModel(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// 切换会话深度思考模�?
func (e *PlaygroundApi) ToggleSessionThinking(c *gin.Context) {
	var req dto.ToggleThinkingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := e.Playground.ToggleThinking(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 上传文件
func (e *PlaygroundApi) UploadFile(c *gin.Context) {
	// 解析表单
	sessionID := c.PostForm("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	// 获取文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// 检查文件大小限�?(50MB)
	const maxFileSize = 50 * 1024 * 1024 // 50MB in bytes
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file size exceeds the maximum limit of 50MB"})
		return
	}

	// 检查文件格�?
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedFormats := []string{".txt", ".md", ".html", ".pdf", ".xlsx", ".docx"}
	formatAllowed := false
	for _, format := range allowedFormats {
		if ext == format {
			formatAllowed = true
			break
		}
	}
	if !formatAllowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file format not allowed. Allowed formats: txt, md, html, pdf, xlsx, docx"})
		return
	}

	// 检查当前会话的文件数量限制
	fileQuery := &dto.GetFilesRequest{
		SessionID: sessionID,
	}
	filesResp, err := e.Playground.GetFiles(c.Request.Context(), fileQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	const maxFileCount = 10
	if len(filesResp.Data) >= maxFileCount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "maximum file count reached (10 files per session)"})
		return
	}

	req := &dto.UploadFileRequest{
		SessionID: sessionID,
	}

	resp, err := e.Playground.UploadFile(c.Request.Context(), req, file, header.Filename, header.Size)
	if err != nil {
		logger.ApiLogger.Error("API: 文件上传到存储失败", "sessionID", sessionID, "filename", header.Filename, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.ApiLogger.Info("API: 文件上传成功", "sessionID", sessionID, "filename", header.Filename, "fileID", resp.Data.ID)

	// 提交数据库事务
	if err := e.DataStore.Commit(c.Request.Context()); err != nil {
		logger.ApiLogger.Warn("API: 提交数据库操作失败，但将继续处理", "error", err)
	}

	// 检�?embedding 服务可用�?
	hasEmbeddingService, embeddingError := e.Playground.CheckEmbeddingService(c.Request.Context(), sessionID)
	if !hasEmbeddingService {
		logger.ApiLogger.Error("API: embed 服务不可用，无法生成向量", "sessionID", sessionID, "fileID", resp.Data.ID, "error", embeddingError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件已上传，但无法生成向量嵌入，请检�?embedding 服务是否可用"})
		return
	}

	// 自动触发文件向量生成，最多重试3次
	logger.ApiLogger.Info("API: 开始生成文件向量", "sessionID", sessionID, "fileID", resp.Data.ID)
	embReq := &dto.GenerateEmbeddingRequest{FileID: resp.Data.ID}
	var embErr error
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Millisecond * 200 * time.Duration(i))
			logger.ApiLogger.Info("API: 重试生成向量", "attempt", i+1, "fileID", resp.Data.ID)
		}
		_, embErr = e.Playground.ProcessFile(c.Request.Context(), embReq)
		if embErr == nil {
			break
		}
		logger.ApiLogger.Error("API: 第次生成向量失败", "attempt", i+1, "fileID", resp.Data.ID, "error", embErr)
	}

	if embErr != nil {
		logger.ApiLogger.Error("API: 自动向量生成最终失败", "fileID", resp.Data.ID, "error", embErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件已上传，但向量生成失败"})
		return
	}

	logger.ApiLogger.Info("API: 文件向量生成并存储成功", "fileID", resp.Data.ID)
	c.JSON(http.StatusOK, gin.H{
		"bcode":             resp.Bcode,
		"data":              resp.Data,
		"embedding_success": true,
	})
}

// 获取文件列表
func (e *PlaygroundApi) GetFiles(c *gin.Context) {
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	req := &dto.GetFilesRequest{
		SessionID: sessionID,
	}

	resp, err := e.Playground.GetFiles(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 删除文件
func (e *PlaygroundApi) DeleteFile(c *gin.Context) {
	var req dto.DeleteFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := e.Playground.DeleteFile(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 处理文件生成嵌入向量
func (e *PlaygroundApi) ProcessFile(c *gin.Context) {
	var req dto.GenerateEmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := e.Playground.ProcessFile(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
