package op

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"urlAPI/internal/llm"
	"urlAPI/internal/model"
	"urlAPI/util"
)

/**
 * @brief 执行文本生成并绘制为图片。
 * @param task 待执行任务。
 * @param host 当前服务主机地址。
 * @param provider 文本提供方配置。
 * @param context 系统上下文提示词。
 * @return GenerateResult 生成结果。
 * @return error 生成文本、绘图或写文件失败时返回错误。
 */
func generateText(task *model.Task, host string, provider util.ProviderConfig, systemContext string) (GenerateResult, error) {
	if provider.Endpoint == "" {
		task.Status = "failed"
		task.Return = "Unknown API"
		return GenerateResult{}, errors.New("text generation unknown API")
	}

	client, err := llm.NewProvider(provider)
	if err != nil {
		task.Status = "failed"
		task.Return = err.Error()
		return GenerateResult{}, fmt.Errorf("text generation client: %w", err)
	}

	req := llm.ChatCompletionRequest{
		Model: task.Model,
		Messages: []llm.ChatMessage{
			{Role: "system", Content: systemContext},
			{Role: "user", Content: task.Target},
		},
		Temperature:      provider.Temperature,
		TopP:             provider.TopP,
		MaxTokens:        provider.MaxTokens,
		PresencePenalty:  provider.PresencePenalty,
		FrequencyPenalty: provider.FrequencyPenalty,
	}

	ctx := context.Background()
	resp, err := client.ChatCompletion(ctx, req)
	if err != nil {
		task.Status = "failed"
		task.Return = err.Error()
		return GenerateResult{}, fmt.Errorf("text generation: %w", err)
	}

	if resp == nil || len(resp.Choices) == 0 {
		task.Status = "failed"
		task.Return = "empty response"
		return GenerateResult{}, errors.New("text generation empty response")
	}

	response := resp.Choices[0].Message.Content

	img, err := util.DrawTxt(response)
	if err != nil {
		task.Status = "failed"
		task.Return = err.Error()
		return GenerateResult{}, fmt.Errorf("text image rendering: %w", err)
	}
	file, err := os.Create(ImgPath + task.UUID + ".png")
	if err != nil {
		task.Status = "failed"
		task.Return = err.Error()
		return GenerateResult{}, fmt.Errorf("text image create: %w", err)
	}
	defer file.Close()
	if _, err = io.Copy(file, bytes.NewReader(img)); err != nil {
		task.Status = "failed"
		task.Return = err.Error()
		return GenerateResult{}, fmt.Errorf("text image write: %w", err)
	}
	result := GenerateResult{Prompt: task.Target, Response: response, URL: host + "/download?img=" + task.UUID}
	if err := setTaskResult(task, result); err != nil {
		return GenerateResult{}, err
	}
	return result, nil
}
