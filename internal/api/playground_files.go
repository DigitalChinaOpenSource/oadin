package api

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"byze/internal/api/dto"

	"github.com/gin-gonic/gin"
)

// 上传文件
func (t *ByzeCoreServer) UploadFile(c *gin.Context) {
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

	// 检查文件大小限制 (50MB)
	const maxFileSize = 50 * 1024 * 1024 // 50MB in bytes
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file size exceeds the maximum limit of 50MB"})
		return
	}

	// 检查文件格式
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
	filesResp, err := t.Playground.GetFiles(c.Request.Context(), fileQuery)
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

	resp, err := t.Playground.UploadFile(c.Request.Context(), req, file, header.Filename, header.Size)
	if err != nil {
		slog.Error("API: 文件上传到存储失败", "sessionID", sessionID, "filename", header.Filename, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	slog.Info("API: 文件上传成功", "sessionID", sessionID, "filename", header.Filename, "fileID", resp.Data.ID)

	// 提交数据库事务
	if err := t.DataStore.Commit(c.Request.Context()); err != nil {
		slog.Warn("API: 提交数据库操作失败，但将继续处理", "error", err)
	}

	// 检查 embedding 服务可用性
	hasEmbeddingService, embeddingError := t.Playground.CheckEmbeddingService(c.Request.Context(), sessionID)
	if !hasEmbeddingService {
		slog.Error("API: embed 服务不可用，无法生成向量", "sessionID", sessionID, "fileID", resp.Data.ID, "error", embeddingError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件已上传，但无法生成向量嵌入，请检查 embedding 服务是否可用"})
		return
	}

	// 自动触发文件向量生成，最多重试 3 次
	slog.Info("API: 开始生成文件向量", "sessionID", sessionID, "fileID", resp.Data.ID)
	embReq := &dto.GenerateEmbeddingRequest{FileID: resp.Data.ID}
	var embErr error
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Millisecond * 200 * time.Duration(i))
			slog.Info("API: 重试生成向量", "attempt", i+1, "fileID", resp.Data.ID)
		}
		_, embErr = t.Playground.ProcessFile(c.Request.Context(), embReq)
		if embErr == nil {
			break
		}
		slog.Error("API: 第次生成向量失败", "attempt", i+1, "fileID", resp.Data.ID, "error", embErr)
	}

	if embErr != nil {
		slog.Error("API: 自动向量生成最终失败", "fileID", resp.Data.ID, "error", embErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件已上传，但向量生成失败。"})
		return
	}

	slog.Info("API: 文件向量生成并存储成功", "fileID", resp.Data.ID)
	c.JSON(http.StatusOK, gin.H{
		"bcode":             resp.Bcode,
		"data":              resp.Data,
		"embedding_success": true,
	})
}

// 获取文件列表
func (t *ByzeCoreServer) GetFiles(c *gin.Context) {
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	req := &dto.GetFilesRequest{
		SessionID: sessionID,
	}

	resp, err := t.Playground.GetFiles(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 删除文件
func (t *ByzeCoreServer) DeleteFile(c *gin.Context) {
	var req dto.DeleteFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := t.Playground.DeleteFile(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 处理文件生成嵌入向量
func (t *ByzeCoreServer) ProcessFile(c *gin.Context) {
	var req dto.GenerateEmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := t.Playground.ProcessFile(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
