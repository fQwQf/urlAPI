package op

import (
	"encoding/json"
	"errors"
	"net/url"
	"urlAPI/internal/database"
	"urlAPI/internal/model"
)

/**
 * @brief 生成文本类图片结果。
 * @param task 待执行任务。
 * @param host 当前服务主机地址。
 * @param skipDB 是否跳过数据库记录。
 * @return model.Task 更新后的任务对象。
 * @return GenerateResult 生成结果。
 * @return error 执行失败时返回错误。
 */
func GenerateTextImage(task model.Task, host string, skipDB bool) (model.Task, GenerateResult, error) {
	settings := database.SettingsStore.Get()
	if _, ok := database.PromptMap[task.Target]; ok {
		task.Target = settings.Prompts.Templates[task.Target]
	}
	if task.API == "" {
		task.API = settings.Text.GenerationAPI
	}
	provider, ok := settings.Providers.ByName(task.API)
	if !ok {
		return failTask(task, "Unknown API"), GenerateResult{URL: settings.Text.FallbackImageURL}, errors.New("text generation unknown API")
	}
	if task.Model == "" {
		task.Model = provider.TextModel
	}
	filter := TaskQueueFilter{Type: "txt.gen", Target: task.Target, API: task.API}
	return ExecuteCachedTask(task, filter, skipDB, func(task *model.Task) (GenerateResult, error) {
		return generateText(task, host, provider, settings.Prompts.GenerationContext)
	})
}

/**
 * @brief 生成图像结果。
 * @param task 待执行任务。
 * @param host 当前服务主机地址。
 * @param skipDB 是否跳过数据库记录。
 * @return model.Task 更新后的任务对象。
 * @return GenerateResult 生成结果。
 * @return error 执行失败时返回错误。
 */
func GenerateImage(task model.Task, host string, skipDB bool) (model.Task, GenerateResult, error) {
	settings := database.SettingsStore.Get()
	if task.API == "" {
		task.API = settings.Image.API
	}
	provider, ok := settings.Providers.ByName(task.API)
	if !ok {
		return failTask(task, "Imggen Process invalid API"), GenerateResult{URL: settings.Image.FallbackImageURL}, errors.New("image generation invalid API")
	}
	if task.Model == "" {
		task.Model = provider.ImageModel
	}
	if task.Size == "" {
		task.Size = provider.ImageSize
	}
	filter := TaskQueueFilter{Type: "img.gen", Size: task.Size, Target: task.Target, API: task.API}
	return ExecuteCachedTask(task, filter, skipDB, func(task *model.Task) (GenerateResult, error) {
		return generateImage(task, host, provider)
	})
}

/**
 * @brief 生成网页封面图结果。
 * @param task 待执行任务。
 * @param host 当前服务主机地址。
 * @param skipDB 是否跳过数据库记录。
 * @return model.Task 更新后的任务对象。
 * @return GenerateResult 生成结果。
 * @return error 执行失败时返回错误。
 */
func GenerateWebImage(task model.Task, host string, skipDB bool) (model.Task, GenerateResult, error) {
	settings := database.SettingsStore.Get()
	parsedURL, err := url.Parse(task.Target)
	if err != nil {
		return failTask(task, err.Error()), GenerateResult{URL: settings.Web.FallbackImageURL}, err
	}
	task.API = parsedURL.Host
	filter := TaskQueueFilter{Type: "web.img", Target: task.Target, API: task.API}
	return ExecuteCachedTask(task, filter, skipDB, func(task *model.Task) (GenerateResult, error) {
		return generateWebImage(task, host)
	})
}

/**
 * @brief 生成随机图片结果。
 * @param task 待执行任务。
 * @param skipDB 是否跳过数据库记录。
 * @return model.Task 更新后的任务对象。
 * @return GenerateResult 生成结果。
 * @return error 执行失败时返回错误。
 */
func GenerateRandom(task model.Task, skipDB bool) (model.Task, GenerateResult, error) {
	settings := database.SettingsStore.Get()
	if task.API == "" {
		task.API = settings.Random.DefaultAPI
	}
	result, err := generateRandom(&task)
	SaveTask(task, skipDB)
	if err != nil {
		return task, GenerateResult{URL: settings.Random.FallbackImageURL}, err
	}
	return task, result, nil
}

/**
 * @brief 下载指定目标图片。
 * @param target 目标图片地址。
 * @return []byte 图片二进制内容。
 * @return string 失败时使用的回退地址。
 * @return error 下载失败时返回错误。
 */
func DownloadImage(target string) ([]byte, string, error) {
	return downloadImage(target)
}

/**
 * @brief 将任务标记为失败。
 * @param task 原始任务对象。
 * @param msg 失败原因。
 * @return model.Task 更新后的任务对象。
 */
func failTask(task model.Task, msg string) model.Task {
	task.Status = "failed"
	task.Return = msg
	return task
}

/**
 * @brief 将生成结果写回任务记录。
 * @param task 待更新的任务对象。
 * @param result 生成结果。
 * @return error 结果序列化失败时返回错误。
 */
func setTaskResult(task *model.Task, result GenerateResult) error {
	body, err := json.Marshal(result)
	if err != nil {
		task.Status = "failed"
		task.Return = err.Error()
		return err
	}
	task.Return = string(body)
	task.Status = "success"
	return nil
}
