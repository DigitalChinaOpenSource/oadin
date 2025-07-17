package dto

import (
	"oadin/extension/entity"
	"oadin/internal/utils/bcode"
	"github.com/mark3labs/mcp-go/mcp"
)

// 创建会话请求
type CreateSessionRequest struct {
	Title        string `json:"title"`
	ModelId      string `json:"modelId"`
	ModelName    string `json:"modelName"`
	EmbedModelId string `json:"embedModelId"`
}

type CreateSessionResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  Session      `json:"data"`
}

type Session struct {
	Id              string `json:"id"`
	Title           string `json:"title"`
	ModelId         string `json:"modelId"`
	ModelName       string `json:"modelName"`
	EmbedModelId    string `json:"embedModelId"`
	ThinkingEnabled bool   `json:"thinkingEnabled"`
	ThinkingActive  bool   `json:"thinkingActive"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type GetSessionsResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  []Session    `json:"data"`
}

type SendMessageRequest struct {
	SessionId string   `json:"sessionId"`
	Content   string   `json:"content"`
	Tools     []Tool   `json:"tools,omitempty"`
	McpIds    []string `json:"mcpIds,omitempty"`
}

// 发送消息响应
type SendMessageResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  []Message    `json:"data"`
}

// 消息信息
type Message struct {
	Id        string               `json:"id"`
	SessionId string               `json:"sessionId"`
	Role      string               `json:"role"`
	Content   string               `json:"content"`
	CreatedAt string               `json:"createdAt"`
	Thoughts  string               `json:"thoughts,omitempty"`
	Type      string               `json:"type,omitempty"`
	ModelId   string               `json:"modelId,omitempty"`
	ModelName string               `json:"modelName,omitempty"`
	ToolCalls []entity.ToolMessage `json:"toolCalls,omitempty"`
}

// 获取会话消息请求
type GetMessagesRequest struct {
	SessionId string `json:"sessionId"`
}

// 获取会话消息响应
type GetMessagesResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  []Message    `json:"data"`
}

// 切换会话模型请求
type ChangeSessionModelRequest struct {
	SessionId    string `json:"sessionId"`
	ModelId      string `json:"modelId"`
	EmbedModelId string `json:"embedModelId,omitempty"`
}

// 删除会话请求
type DeleteSessionRequest struct {
	SessionId string `json:"sessionId"`
}

// 删除会话响应
type DeleteSessionResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
}

type ChangeSessionModelResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  *Session     `json:"data"`
}


type TypeFunction struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Parameters  mcp.ToolInputSchema `json:"parameters"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function TypeFunction `json:"function"`
}

// 聊天请求模型
type ChatRequest struct {
	Model       string              `json:"model"`
	Messages    []map[string]string `json:"messages"`
	Temperature float32             `json:"temperature,omitempty"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Stream      bool                `json:"stream,omitempty"`
	Think       bool                `json:"think"`
	Tools       []Tool              `json:"tools,omitempty"` // 新增，支持Ollama工具调用
}

// 聊天响应模型
type ChatResponse struct {
	ID            string     `json:"id"`
	Object        string     `json:"object"`
	Model         string     `json:"model"`
	ModelName     string     `json:"model_name,omitempty"` // 新增字段
	Content       string     `json:"content"`
	ToolCalls     []ToolCall `json:"tool_calls,omitempty"`     // 新增，支持Ollama工具调用
	IsComplete    bool       `json:"is_complete"`              // 流式输出时，是否是最后一个块
	Thoughts      string     `json:"thinking,omitempty"`       // 深度思考的结果
	Type          string     `json:"type,omitempty"`           // "answer"、"thoughts"等
	TotalDuration int64      `json:"total_duration,omitempty"` // 总耗时，单位秒
	ToolGroupID   string     `json:"tool_group_id,omitempty"`  // 工具组ID，用于关联工具调用
}

type ToolCall struct {
	Function struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	} `json:"function"`
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

type OadinAPIResponse struct {
	BusinessCode int         `json:"business_code"`
	Message      string      `json:"message"`
	Data         interface{} `json:"data"`
}

type OadinChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role      string           `json:"role"`
			Content   string           `json:"content"`
			Thinking  string           `json:"thinking,omitempty"`
			ToolCalls []ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type OadinChatStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
		Delta        struct {
			Role      string           `json:"role"`
			Content   string           `json:"content"`
			Thinking  string           `json:"thinking,omitempty"`
			ToolCalls []ToolCall `json:"tool_calls,omitempty"`
		} `json:"delta"`
	} `json:"choices"`
}

type OadinEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string         `json:"model"`
	Usage map[string]int `json:"usage"`
}