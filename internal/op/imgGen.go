package op

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"urlAPI/internal/model"
	"urlAPI/util"
)

/**
 * @brief 执行图像生成并将结果写入本地文件。
 * @param task 待执行任务。
 * @param host 当前服务主机地址。
 * @param provider 图像提供方配置。
 * @return GenerateResult 生成结果。
 * @return error 调用接口、写文件或序列化结果失败时返回错误。
 */
func generateImage(task *model.Task, host string, provider util.ProviderConfig) (GenerateResult, error) {
	var img []byte
	prompt := task.Target
	var err error
	switch task.API {
	case "alibaba":
		img, prompt, err = util.AlibabaImg(provider.APIKey, task.Target, task.Model, task.Size)
	case "openai":
		img, err = util.OpenaiImg(provider.Endpoint, provider.APIKey, task.Target, task.Model, task.Size)
	default:
		task.Status = "failed"
		task.Return = "Imggen Process invalid API"
		return GenerateResult{}, fmt.Errorf("image generation invalid API")
	}
	if err != nil {
		task.Status = "failed"
		task.Return = err.Error()
		return GenerateResult{}, fmt.Errorf("image generation: %w", err)
	}
	file, err := os.Create(ImgPath + task.UUID + ".png")
	if err != nil {
		task.Status = "failed"
		task.Return = err.Error()
		return GenerateResult{}, fmt.Errorf("image create: %w", err)
	}
	defer file.Close()
	if _, err = io.Copy(file, bytes.NewReader(img)); err != nil {
		task.Status = "failed"
		task.Return = err.Error()
		return GenerateResult{}, fmt.Errorf("image write: %w", err)
	}
	result := GenerateResult{OriginalPrompt: task.Target, ActualPrompt: prompt, URL: host + "/download?img=" + task.UUID}
	if err := setTaskResult(task, result); err != nil {
		return GenerateResult{}, err
	}
	return result, nil
}
