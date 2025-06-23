package types

import (
	"strings"
	"time"
)

type FileChunk struct {
	ID         string      `json:"id"`
	FileID     string      `json:"file_id"`
	Content    string      `json:"content"`
	ChunkIndex int         `json:"index"`
	Embedding  Float32List `json:"embedding"`
	CreatedAt  time.Time   `json:"created_at"`
}

func (f *FileChunk) SetCreateTime(t time.Time) { f.CreatedAt = t }
func (f *FileChunk) SetUpdateTime(t time.Time) {}
func (f *FileChunk) PrimaryKey() string        { return f.ID }
func (f *FileChunk) TableName() string         { return "file_chunks" }
func (f *FileChunk) Index() map[string]interface{} {
	return map[string]interface{}{
		"id":      f.ID,
		"file_id": f.FileID,
	}
}

type File struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	Type      string    `json:"type"`
	ChunkSize int       `json:"chunk_size"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (f *File) SetCreateTime(t time.Time) { f.CreatedAt = t }
func (f *File) SetUpdateTime(t time.Time) { f.UpdatedAt = t }
func (f *File) PrimaryKey() string        { return f.ID }
func (f *File) TableName() string         { return "files" }
func (f *File) Index() map[string]interface{} {
	index := map[string]interface{}{}

	if f.ID != "" {
		index["id"] = f.ID
	}
	if f.SessionID != "" {
		// 确保session_id完全一致
		// 添加trim防止空格问题，sqlite查询是大小写敏感的
		index["session_id"] = strings.TrimSpace(f.SessionID)
	}

	return index
}

type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbeddingResponse struct {
	Object     string         `json:"object"`
	Embeddings [][]float32    `json:"embeddings"`
	Model      string         `json:"model"`
	Usage      EmbeddingUsage `json:"usage"`
}

type EmbeddingData struct {
	Object     string    `json:"object"`
	Embedding  []float32 `json:"embedding"`
	EmbedIndex int       `json:"index"`
}

type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
