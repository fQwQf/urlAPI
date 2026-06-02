package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"urlAPI/util"
)

// anthropicProvider implements Provider for Anthropic Claude API.
type anthropicProvider struct {
	config util.ProviderConfig
}

func newAnthropicProvider(config util.ProviderConfig) *anthropicProvider {
	return &anthropicProvider{config: config}
}

func (p *anthropicProvider) Name() string {
	return "anthropic"
}

func (p *anthropicProvider) IsStreamingSupported() bool {
	return true
}

func (p *anthropicProvider) IsEmbeddingsSupported() bool {
	return false
}

// anthropicRequest is the request format for Anthropic API.
type anthropicRequest struct {
	Model         string              `json:"model"`
	Messages      []anthropicMessage  `json:"messages"`
	System        string              `json:"system,omitempty"`
	MaxTokens     int                 `json:"max_tokens"`
	Temperature   float64             `json:"temperature,omitempty"`
	TopP          float64             `json:"top_p,omitempty"`
	Stream        bool                `json:"stream,omitempty"`
	StopSequences []string            `json:"stop_sequences,omitempty"`
}

// anthropicMessage is a message in Anthropic format.
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse is the response format for Anthropic API.
type anthropicResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Role         string `json:"role"`
	Content      []anthropicContent `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// anthropicContent is a content block in Anthropic response.
type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// anthropicStreamEvent is a stream event from Anthropic.
type anthropicStreamEvent struct {
	Type    string              `json:"type"`
	Index   int                 `json:"index,omitempty"`
	Delta   anthropicStreamDelta `json:"delta,omitempty"`
	Message anthropicStreamMessage `json:"message,omitempty"`
}

type anthropicStreamDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicStreamMessage struct {
	Usage struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (p *anthropicProvider) toAnthropicMessages(req ChatCompletionRequest) ([]anthropicMessage, string) {
	var messages []anthropicMessage
	var system string
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			system = msg.Content
			continue
		}
		role := msg.Role
		if role == "assistant" {
			role = "assistant"
		} else {
			role = "user"
		}
		messages = append(messages, anthropicMessage{
			Role:    role,
			Content: msg.Content,
		})
	}
	return messages, system
}

func (p *anthropicProvider) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if p.config.Endpoint == "" {
		return nil, fmt.Errorf("endpoint not configured")
	}
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	messages, system := p.toAnthropicMessages(req)
	anthropicReq := anthropicRequest{
		Model:       req.Model,
		Messages:    messages,
		System:      system,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}
	if anthropicReq.MaxTokens == 0 {
		anthropicReq.MaxTokens = 4096
	}

	payload, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	for k, v := range p.config.CustomHeaders {
		httpReq.Header.Set(k, v)
	}

	resp, err := util.GlobalHTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil {
			return nil, fmt.Errorf("API error: %s", errResp.Error.Message)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, err
	}

	var content string
	for _, c := range anthropicResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &ChatCompletionResponse{
		ID:      anthropicResp.ID,
		Object:  "chat.completion",
		Created: nowUnix(),
		Model:   anthropicResp.Model,
		Choices: []Choice{{
			Index: 0,
			Message: ChatMessage{
				Role:    "assistant",
				Content: content,
			},
			FinishReason: anthropicResp.StopReason,
		}},
		Usage: Usage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
	}, nil
}

func (p *anthropicProvider) ChatCompletionStream(ctx context.Context, req ChatCompletionRequest) (<-chan StreamEvent, error) {
	if p.config.Endpoint == "" {
		return nil, fmt.Errorf("endpoint not configured")
	}
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	messages, system := p.toAnthropicMessages(req)
	anthropicReq := anthropicRequest{
		Model:       req.Model,
		Messages:    messages,
		System:      system,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      true,
	}
	if anthropicReq.MaxTokens == 0 {
		anthropicReq.MaxTokens = 4096
	}

	payload, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")

	for k, v := range p.config.CustomHeaders {
		httpReq.Header.Set(k, v)
	}

	resp, err := util.GlobalHTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	events := make(chan StreamEvent)
	go func() {
		defer close(events)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			events <- StreamEvent{Error: fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		id := generateID("chatcmpl")
		model := req.Model

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" || !strings.HasPrefix(line, "event: ") {
				continue
			}
			eventType := strings.TrimPrefix(line, "event: ")
			if !scanner.Scan() {
				break
			}
			dataLine := scanner.Text()
			if !strings.HasPrefix(dataLine, "data: ") {
				continue
			}
			data := strings.TrimPrefix(dataLine, "data: ")

			if eventType == "message_stop" {
				events <- StreamEvent{Done: true}
				return
			}
			if eventType != "content_block_delta" {
				continue
			}

			var streamEvent anthropicStreamEvent
			if err := json.Unmarshal([]byte(data), &streamEvent); err != nil {
				continue
			}

			chunk := ChatCompletionResponse{
				ID:      id,
				Object:  "chat.completion.chunk",
				Created: nowUnix(),
				Model:   model,
				Choices: []Choice{{
					Index: 0,
					Delta: ChatMessage{
						Role:    "assistant",
						Content: streamEvent.Delta.Text,
					},
					FinishReason: "",
				}},
			}
			events <- StreamEvent{Data: ToJSON(chunk)}
		}
		if err := scanner.Err(); err != nil {
			events <- StreamEvent{Error: err}
		}
	}()

	return events, nil
}

func (p *anthropicProvider) Embeddings(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error) {
	return nil, fmt.Errorf("embeddings not supported by Anthropic")
}

func (p *anthropicProvider) Models(ctx context.Context) (*ModelsResponse, error) {
	return &ModelsResponse{
		Object: "list",
		Data: []Model{
			{ID: "claude-3-5-sonnet-20241022", Object: "model", OwnedBy: "anthropic"},
			{ID: "claude-3-5-haiku-20241022", Object: "model", OwnedBy: "anthropic"},
			{ID: "claude-3-opus-20240229", Object: "model", OwnedBy: "anthropic"},
			{ID: "claude-3-sonnet-20240229", Object: "model", OwnedBy: "anthropic"},
			{ID: "claude-3-haiku-20240307", Object: "model", OwnedBy: "anthropic"},
		},
	}, nil
}
