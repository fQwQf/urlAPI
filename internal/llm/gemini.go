package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"urlAPI/util"
)

// geminiProvider implements Provider for Google Gemini API.
type geminiProvider struct {
	config util.ProviderConfig
}

func newGeminiProvider(config util.ProviderConfig) *geminiProvider {
	return &geminiProvider{config: config}
}

func (p *geminiProvider) Name() string {
	return "gemini"
}

func (p *geminiProvider) IsStreamingSupported() bool {
	return true
}

func (p *geminiProvider) IsEmbeddingsSupported() bool {
	return true
}

// geminiRequest is the request format for Gemini API.
type geminiRequest struct {
	Contents         []geminiContent      `json:"contents"`
	SystemInstruction *geminiContent      `json:"systemInstruction,omitempty"`
	GenerationConfig  geminiGenConfig     `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}
type geminiPart struct {
	Text string `json:"text"`
}
type geminiGenConfig struct {
	Temperature      float64  `json:"temperature,omitempty"`
	TopP             float64  `json:"topP,omitempty"`
	MaxOutputTokens  int      `json:"maxOutputTokens,omitempty"`
	StopSequences    []string `json:"stopSequences,omitempty"`
}

// geminiResponse is the response format for Gemini API.
type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func (p *geminiProvider) toGeminiMessages(req ChatCompletionRequest) ([]geminiContent, *geminiContent) {
	var contents []geminiContent
	var system *geminiContent
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			system = &geminiContent{
				Parts: []geminiPart{{Text: msg.Content}},
			}
			continue
		}
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: msg.Content}},
		})
	}
	return contents, system
}

func (p *geminiProvider) buildEndpoint(model, action string) string {
	endpoint := p.config.Endpoint
	if strings.HasSuffix(endpoint, "/") {
		endpoint = strings.TrimSuffix(endpoint, "/")
	}
	if !strings.Contains(endpoint, "/models/") {
		endpoint = endpoint + "/models/" + model
	} else {
		endpoint = strings.Replace(endpoint, "/models/"+p.config.TextModel, "/models/"+model, 1)
	}
	return endpoint + ":" + action + "?key=" + url.QueryEscape(p.config.APIKey)
}

func (p *geminiProvider) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	contents, system := p.toGeminiMessages(req)
	geminiReq := geminiRequest{
		Contents: contents,
		GenerationConfig: geminiGenConfig{
			Temperature:     req.Temperature,
			TopP:            req.TopP,
			MaxOutputTokens: req.MaxTokens,
			StopSequences:   req.Stop,
		},
	}
	if system != nil {
		geminiReq.SystemInstruction = system
	}

	payload, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, err
	}

	endpoint := p.buildEndpoint(req.Model, "generateContent")
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
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
				Code    int    `json:"code"`
				Message string `json:"message"`
				Status  string `json:"status"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil {
			return nil, fmt.Errorf("API error: %s", errResp.Error.Message)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, err
	}

	var content string
	var finishReason string
	if len(geminiResp.Candidates) > 0 {
		for _, part := range geminiResp.Candidates[0].Content.Parts {
			content += part.Text
		}
		finishReason = geminiResp.Candidates[0].FinishReason
	}

	return &ChatCompletionResponse{
		ID:      generateID("chatcmpl"),
		Object:  "chat.completion",
		Created: nowUnix(),
		Model:   req.Model,
		Choices: []Choice{{
			Index: 0,
			Message: ChatMessage{
				Role:    "assistant",
				Content: content,
			},
			FinishReason: finishReason,
		}},
		Usage: Usage{
			PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
		},
	}, nil
}

func (p *geminiProvider) ChatCompletionStream(ctx context.Context, req ChatCompletionRequest) (<-chan StreamEvent, error) {
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	contents, system := p.toGeminiMessages(req)
	geminiReq := geminiRequest{
		Contents: contents,
		GenerationConfig: geminiGenConfig{
			Temperature:     req.Temperature,
			TopP:            req.TopP,
			MaxOutputTokens: req.MaxTokens,
			StopSequences:   req.Stop,
		},
	}
	if system != nil {
		geminiReq.SystemInstruction = system
	}

	payload, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, err
	}

	endpoint := p.buildEndpoint(req.Model, "streamGenerateContent")
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
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

		id := generateID("chatcmpl")
		model := req.Model

		decoder := json.NewDecoder(resp.Body)
		for decoder.More() {
			var geminiResp geminiResponse
			if err := decoder.Decode(&geminiResp); err != nil {
				if err == io.EOF {
					break
				}
				continue
			}

			var content string
			var finishReason string
			if len(geminiResp.Candidates) > 0 {
				for _, part := range geminiResp.Candidates[0].Content.Parts {
					content += part.Text
				}
				finishReason = geminiResp.Candidates[0].FinishReason
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
						Content: content,
					},
					FinishReason: finishReason,
				}},
			}
			events <- StreamEvent{Data: ToJSON(chunk)}
		}
		events <- StreamEvent{Done: true}
	}()

	return events, nil
}

func (p *geminiProvider) Embeddings(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error) {
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	embeddingModel := p.config.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = "text-embedding-004"
	}

	var inputs []string
	switch v := req.Input.(type) {
	case string:
		inputs = []string{v}
	case []string:
		inputs = v
	default:
		return nil, fmt.Errorf("invalid input type")
	}

	var embeddings []Embedding
	for i, text := range inputs {
		payload, _ := json.Marshal(map[string]any{
			"content": map[string]any{"parts": []map[string]string{{"text": text}}},
		})

		endpoint := p.buildEndpoint(embeddingModel, "embedContent")
		httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}

		httpReq.Header.Set("Content-Type", "application/json")
		resp, err := util.GlobalHTTPClient.Do(httpReq)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var result struct {
			Embedding struct {
				Values []float64 `json:"values"`
			} `json:"embedding"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}

		embeddings = append(embeddings, Embedding{
			Object:    "embedding",
			Embedding: result.Embedding.Values,
			Index:     i,
		})
	}

	return &EmbeddingResponse{
		Object: "list",
		Data:   embeddings,
		Model:  embeddingModel,
	}, nil
}

func (p *geminiProvider) Models(ctx context.Context) (*ModelsResponse, error) {
	return &ModelsResponse{
		Object: "list",
		Data: []Model{
			{ID: "gemini-2.0-flash", Object: "model", OwnedBy: "google"},
			{ID: "gemini-2.0-flash-lite", Object: "model", OwnedBy: "google"},
			{ID: "gemini-2.0-pro-exp-02-05", Object: "model", OwnedBy: "google"},
			{ID: "gemini-1.5-pro", Object: "model", OwnedBy: "google"},
			{ID: "gemini-1.5-flash", Object: "model", OwnedBy: "google"},
		},
	}, nil
}
