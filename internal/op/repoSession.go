package op

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"urlAPI/internal/database"
	"urlAPI/internal/model"
	"urlAPI/util"
)

/**
 * @brief 新建仓库源配置并抓取内容列表。
 * @param info 会话请求与响应对象。
 * @return error 抓取或保存失败时返回错误。
 */
func newRepo(info *Session) error {
	var err error
	var content []string
	switch info.RepoAPI {
	case "github":
		content, err = util.GetRepo("https://api.github.com/repos/" + info.RepoInfo + "/contents")
		if err != nil {
			return errors.WithStack(err)
		}
		util.ListReplacer(&content, "https://raw.githubusercontent.com", database.SettingsStore.Get().Random.SourceRewriteFrom)
	case "gitee":
		content, err = util.GetRepo("https://gitee.com/api/v5/repos/" + info.RepoInfo + "/contents")
		if err != nil {
			return errors.WithStack(err)
		}
	default:
		err = errors.WithStack(errors.New(info.RepoAPI + " is not supported"))
	}
	content = util.LinkFilter(content)
	jsonString, err := json.Marshal(content)
	if err != nil {
		return err
	}
	repoDB := model.Repo{
		UUID:    uuid.New().String(),
		API:     info.RepoAPI,
		Info:    info.RepoInfo,
		Content: string(jsonString),
	}
	return errors.WithStack(db.CreateRepo(&repoDB))
}

/**
 * @brief 刷新指定仓库源的缓存内容。
 * @param info 会话请求与响应对象。
 * @return error 查询、抓取或保存失败时返回错误。
 */
func refreshRepo(info *Session) error {
	repoFinder := model.Repo{
		UUID: info.RepoUUID,
	}
	repoDBList, err := db.ReadRepo(repoFinder)
	if err != nil {
		return errors.WithStack(err)
	}
	repoDB := (*repoDBList).RepoList[0]
	info.RepoAPI = repoDB.API
	info.RepoInfo = repoDB.Info
	var content []string
	switch info.RepoAPI {
	case "github":
		content, err = util.GetRepo("https://api.github.com/repos/" + info.RepoInfo + "/contents")
		if err != nil {
			return errors.WithStack(err)
		}
		util.ListReplacer(&content, "https://raw.githubusercontent.com", database.SettingsStore.Get().Random.SourceRewriteFrom)
	case "gitee":
		content, err = util.GetRepo("https://gitee.com/api/v5/repos/" + info.RepoInfo + "/contents")
		if err != nil {
			return errors.WithStack(err)
		}
	default:
		err = errors.WithStack(errors.New(info.RepoAPI + " is not supported"))
	}
	content = util.LinkFilter(content)
	jsonString, err := json.Marshal(content)
	if err != nil {
		return errors.WithStack(err)
	}
	repoDB.Content = string(jsonString)
	return errors.WithStack(db.UpdateRepo(&repoDB))
}

/**
 * @brief 删除仓库源配置。
 * @param info 会话请求与响应对象。
 * @return error 删除失败时返回错误。
 */
func delRepo(info *Session) error {
	repoDB := model.Repo{
		UUID: info.RepoUUID,
	}
	return errors.WithStack(db.DeleteRepo(&repoDB))
}

/**
 * @brief 查询全部仓库源配置。
 * @param info 会话请求与响应对象。
 * @return error 查询失败时返回错误。
 */
func fetchRepo(info *Session) error {
	repoFinder := model.Repo{}
	repoDBList, err := db.ReadRepo(repoFinder)
	info.RepoData = repoDBList.RepoList
	if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
		return errors.WithStack(err)
	}
	return nil
}
