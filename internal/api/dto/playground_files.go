package dto

import "byze/internal/utils/bcode"

// 文件上传请求
type UploadFileRequest struct {
	SessionID string `json:"session_id" form:"session_id"`
}

// 文件上传响应
type UploadFileResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  FileInfo     `json:"data"`
}

// 文件信息
type FileInfo struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
}

// 获取会话文件请求
type GetFilesRequest struct {
	SessionID string `json:"session_id"`
}

// 获取会话文件响应
type GetFilesResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  []FileInfo   `json:"data"`
}

// 删除文件请求
type DeleteFileRequest struct {
	FileID string `json:"fileId"`
}

// 删除文件响应
type DeleteFileResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
}

// 嵌入请求
type GenerateEmbeddingRequest struct {
	FileID string `json:"file_id"`
	Model  string `json:"model"`
}

// 嵌入响应
type GenerateEmbeddingResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
}
