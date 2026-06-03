package handles

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"urlAPI/internal/database"
	"urlAPI/internal/llm"
	"urlAPI/internal/model"
	"urlAPI/internal/server/middleware"
	"urlAPI/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ChatCompletionHandler handles /v1/chat/completions requests.
func ChatCompletionHandler(c *gin.Context) {
	var req llm.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}

	providerName := c.GetHeader("X-Provider")
	if providerName == "" {
		settings := database.SettingsStore.Get()
		providerName = settings.Text.GenerationAPI
	}

	providerConfig, ok := database.SettingsStore.Get().Providers.ByName(providerName)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown provider"})
		return
	}

	if providerConfig.Enabled == false {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "provider disabled"})
		return
	}

	if req.Model == "" && providerConfig.TextModel != "" {
		req.Model = providerConfig.TextModel
	}

	client, err := llm.NewProvider(providerConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	task := model.Task{
		UUID:   uuid.New().String(),
		Time:   time.Now(),
		IP:     c.ClientIP(),
		Type:   "txt.gen",
		Target: req.Messages[0].Content,
		API:    providerName,
		Model:  req.Model,
		Status: "running",
	}

	// Record API Key usage if authenticated
	if key, ok := middleware.GetAPIKey(c); ok {
		database.APIKeyStore.IncrementUsage(key.KeyHash)
	}

	ctx := context.Background()

	if req.Stream {
		if !client.IsStreamingSupported() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "streaming not supported by provider"})
			return
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		stream, err := client.ChatCompletionStream(ctx, req)
		if err != nil {
			task.Status = "failed"
			task.Return = err.Error()
			database.CreateTask(&task)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Stream(func(w io.Writer) bool {
			for event := range stream {
				if event.Error != nil {
					fmt.Fprintf(w, "data: %s\n\n", llm.ToJSON(llm.ChatCompletionResponse{
						Object: "chat.completion.chunk",
						Error:  &llm.Error{Message: event.Error.Error(), Type: "api_error"},
					}))
					return false
				}
				if event.Done {
					fmt.Fprintf(w, "data: [DONE]\n\n")
					task.Status = "success"
					database.CreateTask(&task)
					return false
				}
				fmt.Fprintf(w, "data: %s\n\n", event.Data)
			}
			return false
		})
		return
	}

	resp, err := client.ChatCompletion(ctx, req)
	if err != nil {
		task.Status = "failed"
		task.Return = err.Error()
		database.CreateTask(&task)
		if resp != nil && resp.Error != nil {
			c.JSON(http.StatusBadGateway, resp)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	task.Status = "success"
	body, _ := json.Marshal(resp)
	task.Return = string(body)
	database.CreateTask(&task)

	c.JSON(http.StatusOK, resp)
}

// EmbeddingsHandler handles /v1/embeddings requests.
func EmbeddingsHandler(c *gin.Context) {
	var req llm.EmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	providerName := c.GetHeader("X-Provider")
	if providerName == "" {
		settings := database.SettingsStore.Get()
		providerName = settings.Text.GenerationAPI
	}

	providerConfig, ok := database.SettingsStore.Get().Providers.ByName(providerName)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown provider"})
		return
	}

	if providerConfig.Enabled == false {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "provider disabled"})
		return
	}

	if req.Model == "" && providerConfig.EmbeddingModel != "" {
		req.Model = providerConfig.EmbeddingModel
	}

	client, err := llm.NewProvider(providerConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !client.IsEmbeddingsSupported() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "embeddings not supported by provider"})
		return
	}

	ctx := context.Background()
	resp, err := client.Embeddings(ctx, req)
	if err != nil {
		if resp != nil && resp.Error != nil {
			c.JSON(http.StatusBadGateway, resp)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ModelsHandler handles /v1/models requests.
func ModelsHandler(c *gin.Context) {
	providerName := c.Query("provider")
	if providerName == "" {
		providerName = c.GetHeader("X-Provider")
	}

	var models []llm.Model

	if providerName != "" {
		providerConfig, ok := database.SettingsStore.Get().Providers.ByName(providerName)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown provider"})
			return
		}

		client, err := llm.NewProvider(providerConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx := context.Background()
		resp, err := client.Models(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		models = resp.Data
	} else {
		settings := database.SettingsStore.Get()
		providers := map[string]util.ProviderConfig{
			"openai":    settings.Providers.OpenAI,
			"deepseek":  settings.Providers.DeepSeek,
			"alibaba":   settings.Providers.Alibaba,
			"anthropic": settings.Providers.Anthropic,
			"gemini":    settings.Providers.Gemini,
			"azure":     settings.Providers.Azure,
			"moonshot":  settings.Providers.Moonshot,
			"otherapi":  settings.Providers.OtherAPI,
		}

		for name, config := range providers {
			if !config.Enabled {
				continue
			}
			client, err := llm.NewProvider(config)
			if err != nil {
				continue
			}
			ctx := context.Background()
			resp, err := client.Models(ctx)
			if err != nil {
				continue
			}
			for _, m := range resp.Data {
				m.ID = name + "/" + m.ID
				models = append(models, m)
			}
		}
	}

	c.JSON(http.StatusOK, llm.ModelsResponse{
		Object: "list",
		Data:   models,
	})
}
