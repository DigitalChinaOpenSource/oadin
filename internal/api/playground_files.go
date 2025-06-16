package api

import (
	"net/http"
	"path/filepath"
	"strings"

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
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
