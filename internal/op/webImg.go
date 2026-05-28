package op

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"urlAPI/internal/database"
	"urlAPI/internal/model"
	"urlAPI/util"
)

/**
 * @brief 从 Bilibili URL 中提取视频标识。
 * @param URL 原始链接。
 * @return string 视频 BV 或 av 标识。
 */
func getBiliABV(URL string) string {
	for i := 31; i < len(URL); i++ {
		if URL[i] == '/' || URL[i] == '?' {
			return URL[31:i]
		}
	}
	return URL[31:]
}

/**
 * @brief 从 YouTube URL 中提取视频 ID。
 * @param URL 原始链接。
 * @return string 视频 ID。
 */
func getYtbID(URL string) string {
	for i := 32; i < len(URL); i++ {
		if URL[i] == '&' {
			return URL[32:i]
		}
	}
	return URL[32:]
}

/**
 * @brief 根据网页地址生成对应的封面图。
 * @param task 待执行任务。
 * @param host 当前服务主机地址。
 * @return GenerateResult 生成结果。
 * @return error 目标站点不支持或处理失败时返回错误。
 */
func generateWebImage(task *model.Task, host string) (GenerateResult, error) {
	var img []byte
	var err error
	settings := database.SettingsStore.Get()
	switch task.API {
	case "www.bilibili.com":
		img, err = util.Bili(getBiliABV(task.Target))
	case "www.youtube.com":
		img, err = util.Ytb(getYtbID(task.Target), settings.Web.YouTubeToken)
	case "arxiv.org":
		img, err = util.Arxiv(task.Target)
	case "www.ithome.com":
		api := settings.Text.SummaryAPI
		provider, ok := settings.Providers.ByName(api)
		if !ok {
			task.Status = "failed"
			task.Return = "Invalid API"
			return GenerateResult{}, fmt.Errorf("web image invalid summary API")
		}
		img, err = util.ITHome(task.Target, provider.Endpoint, provider.APIKey, provider.SummaryModel, settings.Prompts.SummaryContext)
	case "github.com", "gitee.com":
		img, err = util.Repo(task.Target, settings.Web.RepoToken)
	default:
		task.Status = "failed"
		task.Return = "Invalid URL"
		return GenerateResult{}, fmt.Errorf("web image invalid URL")
	}
	if err != nil {
		task.Status = "failed"
		task.Return = err.Error()
		return GenerateResult{}, fmt.Errorf("web image generation: %w", err)
	}
	file, err := os.Create(ImgPath + task.UUID + ".png")
	if err != nil {
		task.Status = "failed"
		task.Return = err.Error()
		return GenerateResult{}, fmt.Errorf("web image create: %w", err)
	}
	defer file.Close()
	if _, err = io.Copy(file, bytes.NewReader(img)); err != nil {
		task.Status = "failed"
		task.Return = err.Error()
		return GenerateResult{}, fmt.Errorf("web image write: %w", err)
	}
	result := GenerateResult{URL: host + "/download?img=" + task.UUID}
	if err := setTaskResult(task, result); err != nil {
		return GenerateResult{}, err
	}
	return result, nil
}
