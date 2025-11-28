package models

import "time"

// ===== OpenAI 格式 =====

type OpenAIChatRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIChatResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   *OpenAIUsage   `json:"usage,omitempty"`
}

type OpenAIChoice struct {
	Index        int            `json:"index"`
	Message      *OpenAIMessage `json:"message,omitempty"`
	Delta        *OpenAIDelta   `json:"delta,omitempty"`
	FinishReason *string        `json:"finish_reason"`
}

type OpenAIDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAI Models 响应
type OpenAIModelsResponse struct {
	Object string        `json:"object"`
	Data   []OpenAIModel `json:"data"`
}

type OpenAIModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ===== Cosine 格式 =====

type CosineChatRequest struct {
	ID         string          `json:"id"`
	Messages   []CosineMessage `json:"messages"`
	Model      string          `json:"model"`
	TeamID     string          `json:"teamId"`
	Visibility string          `json:"visibility"`
}

type CosineMessage struct {
	Content   string `json:"content"`
	Role      string `json:"role"`
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
}

// Cosine 流式响应中的结束标记
type CosineFinishEvent struct {
	FinishReason string `json:"finishReason"`
	Usage        struct {
		PromptTokens     *int `json:"promptTokens"`
		CompletionTokens *int `json:"completionTokens"`
	} `json:"usage"`
	IsContinued bool `json:"isContinued,omitempty"`
}

// ===== 数据库模型 =====

type Account struct {
	ID        int       `json:"id"`
	Auth      string    `json:"auth"`
	TeamID    string    `json:"team_id"`
	LinuxdoID *string   `json:"linuxdo_id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ===== 通用响应 =====

type HealthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}
