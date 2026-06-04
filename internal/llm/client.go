package llm

import (
	"fmt"
	"time"
	"urlAPI/util"
)

// clientFactory creates provider clients based on configuration.
type clientFactory struct{}

// NewProvider creates a new provider client based on the configuration.
func NewProvider(config util.ProviderConfig) (Provider, error) {
	apiType := config.APIType
	if apiType == "" {
		apiType = "openai"
	}

	switch apiType {
	case "openai", "azure", "deepseek", "alibaba", "moonshot":
		return newOpenAICompatibleProvider(config, apiType), nil
	case "anthropic":
		return newAnthropicProvider(config), nil
	case "gemini":
		return newGeminiProvider(config), nil
	default:
		return newOpenAICompatibleProvider(config, "openai"), nil
	}
}

// GetDefaultModels returns default models for a provider.
func GetDefaultModels(providerName string) (text, summary, image, embedding string) {
	switch providerName {
	case "openai":
		return "gpt-4o", "gpt-4o-mini", "dall-e-3", "text-embedding-3-small"
	case "deepseek":
		return "deepseek-chat", "deepseek-chat", "", ""
	case "alibaba":
		return "deepseek-v3", "qwen-turbo", "wanx2.1-t2i-turbo", ""
	case "anthropic":
		return "claude-3-5-sonnet-20241022", "claude-3-haiku-20240307", "", ""
	case "gemini":
		return "gemini-2.0-flash", "gemini-2.0-flash-lite", "", ""
	case "azure":
		return "gpt-4o", "gpt-4o-mini", "", "text-embedding-3-small"
	case "moonshot":
		return "moonshot-v1-8k", "moonshot-v1-8k", "", ""
	default:
		return "", "", "", ""
	}
}

// IsProviderSupported checks if a provider is supported.
func IsProviderSupported(name string) bool {
	switch name {
	case "openai", "deepseek", "alibaba", "anthropic", "gemini", "azure", "moonshot", "otherapi":
		return true
	default:
		return false
	}
}

// buildChatCompletionRequest builds a ChatCompletionRequest from provider config and messages.
func buildChatCompletionRequest(config util.ProviderConfig, messages []ChatMessage, stream bool) ChatCompletionRequest {
	req := ChatCompletionRequest{
		Model:       config.TextModel,
		Messages:    messages,
		Temperature: config.Temperature,
		TopP:        config.TopP,
		Stream:      stream,
	}
	if config.MaxTokens > 0 {
		req.MaxTokens = config.MaxTokens
	}
	if config.PresencePenalty != 0 {
		req.PresencePenalty = config.PresencePenalty
	}
	if config.FrequencyPenalty != 0 {
		req.FrequencyPenalty = config.FrequencyPenalty
	}
	return req
}

// nowUnix returns the current Unix timestamp.
func nowUnix() int64 {
	return time.Now().Unix()
}

// generateID generates a unique ID for chat completions.
func generateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, nowUnix())
}

// newErrorResponse creates an error response.
func newErrorResponse(msg, typ, code string) *ChatCompletionResponse {
	return &ChatCompletionResponse{
		Object: "chat.completion",
		Error:  &Error{Message: msg, Type: typ, Code: code},
	}
}

// newStreamError creates a stream error event.
func newStreamError(msg, typ, code string) StreamEvent {
	resp := &ChatCompletionResponse{
		Object: "chat.completion.chunk",
		Error:  &Error{Message: msg, Type: typ, Code: code},
	}
	return StreamEvent{Data: ToJSON(resp), Error: fmt.Errorf("%s", msg)}
}
