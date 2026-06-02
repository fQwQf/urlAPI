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

// openAICompatibleProvider implements Provider for OpenAI-compatible APIs.
type openAICompatibleProvider struct {
	config util.ProviderConfig
	name   string
}

func newOpenAICompatibleProvider(config util.ProviderConfig, name string) *openAICompatibleProvider {
	return &openAICompatibleProvider{
		config: config,
		name:   name,
	}
}

func (p *openAICompatibleProvider) Name() string {
	return p.name
}

func (p *openAICompatibleProvider) IsStreamingSupported() bool {
	return true
}

func (p *openAICompatibleProvider) IsEmbeddingsSupported() bool {
	return p.config.EmbeddingModel != ""
}

func (p *openAICompatibleProvider) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if p.config.Endpoint == "" {
		return nil, fmt.Errorf("endpoint not configured")
	}
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

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
			Error Error `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
			return &ChatCompletionResponse{Error: &errResp.Error}, fmt.Errorf("API error: %s", errResp.Error.Message)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result ChatCompletionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (p *openAICompatibleProvider) ChatCompletionStream(ctx context.Context, req ChatCompletionRequest) (<-chan StreamEvent, error) {
	if p.config.Endpoint == "" {
		return nil, fmt.Errorf("endpoint not configured")
	}
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	req.Stream = true
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Connection", "keep-alive")

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
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				events <- StreamEvent{Done: true}
				return
			}
			events <- StreamEvent{Data: data}
		}
		if err := scanner.Err(); err != nil {
			events <- StreamEvent{Error: err}
		}
	}()

	return events, nil
}

func (p *openAICompatibleProvider) Embeddings(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error) {
	if p.config.EmbeddingModel == "" {
		return nil, fmt.Errorf("embedding model not configured")
	}

	embeddingEndpoint := strings.Replace(p.config.Endpoint, "/chat/completions", "/embeddings", 1)
	if embeddingEndpoint == p.config.Endpoint {
		embeddingEndpoint = p.config.Endpoint + "/embeddings"
	}

	req.Model = p.config.EmbeddingModel
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", embeddingEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

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
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result EmbeddingResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (p *openAICompatibleProvider) Models(ctx context.Context) (*ModelsResponse, error) {
	modelsEndpoint := strings.Replace(p.config.Endpoint, "/chat/completions", "/models", 1)
	if modelsEndpoint == p.config.Endpoint {
		modelsEndpoint = p.config.Endpoint + "/models"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", modelsEndpoint, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

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
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result ModelsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
