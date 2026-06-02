package llm

import (
	"context"
	"encoding/json"
	"io"
)

// ChatMessage represents a message in a chat conversation.
type ChatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolCall represents a tool call in a message.
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// ChatCompletionRequest represents a request to the chat completion API.
type ChatCompletionRequest struct {
	Model            string        `json:"model"`
	Messages         []ChatMessage `json:"messages"`
	Temperature      float64       `json:"temperature,omitempty"`
	TopP             float64       `json:"top_p,omitempty"`
	MaxTokens        int           `json:"max_tokens,omitempty"`
	PresencePenalty  float64       `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64       `json:"frequency_penalty,omitempty"`
	Stream           bool          `json:"stream,omitempty"`
	Stop             []string      `json:"stop,omitempty"`
	Seed             int           `json:"seed,omitempty"`
	Tools            []Tool        `json:"tools,omitempty"`
	ToolChoice       any           `json:"tool_choice,omitempty"`
	ResponseFormat   any           `json:"response_format,omitempty"`
}

// Tool represents a tool definition.
type Tool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Parameters  any    `json:"parameters"`
	} `json:"function"`
}

// ChatCompletionResponse represents a response from the chat completion API.
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
	Error   *Error   `json:"error,omitempty"`
}

// Choice represents a choice in a chat completion response.
type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
	Delta        ChatMessage `json:"delta,omitempty"`
}

// Usage represents token usage information.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Error represents an API error.
type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// EmbeddingRequest represents a request to the embeddings API.
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input any      `json:"input"`
	User  string   `json:"user,omitempty"`
}

// EmbeddingResponse represents a response from the embeddings API.
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  Usage       `json:"usage"`
	Error  *Error      `json:"error,omitempty"`
}

// Embedding represents a single embedding.
type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// Model represents a model information.
type Model struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Created    int64  `json:"created"`
	OwnedBy    string `json:"owned_by"`
}

// ModelsResponse represents a response from the models API.
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
	Error  *Error  `json:"error,omitempty"`
}

// StreamEvent represents a single SSE stream event.
type StreamEvent struct {
	Data  string
	Error error
	Done  bool
}

// Provider defines the interface for LLM providers.
type Provider interface {
	// Name returns the provider name.
	Name() string

	// ChatCompletion sends a chat completion request and returns the response.
	ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error)

	// ChatCompletionStream sends a chat completion request and returns a stream of events.
	ChatCompletionStream(ctx context.Context, req ChatCompletionRequest) (<-chan StreamEvent, error)

	// Embeddings sends an embedding request and returns the response.
	Embeddings(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error)

	// Models returns the list of available models.
	Models(ctx context.Context) (*ModelsResponse, error)

	// IsStreamingSupported returns whether the provider supports streaming.
	IsStreamingSupported() bool

	// IsEmbeddingsSupported returns whether the provider supports embeddings.
	IsEmbeddingsSupported() bool
}

// SSEWriter is a helper to write SSE events.
type SSEWriter struct {
	Writer io.Writer
}

// WriteEvent writes a single SSE event.
func (w *SSEWriter) WriteEvent(data string) error {
	_, err := w.Writer.Write([]byte("data: " + data + "\n\n"))
	return err
}

// WriteDone writes the SSE done event.
func (w *SSEWriter) WriteDone() error {
	_, err := w.Writer.Write([]byte("data: [DONE]\n\n"))
	return err
}

// ToJSON marshals a value to JSON string.
func ToJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
