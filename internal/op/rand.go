package op

import (
	"fmt"
	"math/rand"
	"urlAPI/internal/database"
	"urlAPI/internal/model"
)

/**
 * @brief 从缓存仓库内容中随机选择一张图片地址。
 * @param task 待执行任务。
 * @return GenerateResult 随机选择的结果。
 * @return error 仓库不存在或结果写回失败时返回错误。
 */
func generateRandom(task *model.Task) (GenerateResult, error) {
	content, ok := database.RepoMap[task.API+";"+task.Target]
	if !ok || len(content) == 0 {
		task.Status = "failed"
		task.Return = "Repo not found"
		return GenerateResult{}, fmt.Errorf("random repository not found")
	}
	result := GenerateResult{URL: content[rand.Intn(len(content))]}
	if err := setTaskResult(task, result); err != nil {
		return GenerateResult{}, err
	}
	return result, nil
}
