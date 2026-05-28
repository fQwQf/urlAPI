package database

import (
	"encoding/json"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"urlAPI/internal/model"
)

/**
 * @brief 创建仓库缓存记录。
 * @param repo 待写入的仓库记录。
 * @return error 写入失败时返回错误。
 */
func (adapter *SQLiteAdapter) CreateRepo(repo *model.Repo) error {
	return errors.WithStack(adapter.db.Create(repo).Error)
}

/**
 * @brief 更新仓库缓存记录并同步内存缓存。
 * @param repo 待更新的仓库记录。
 * @return error 更新失败时返回错误。
 */
func (adapter *SQLiteAdapter) UpdateRepo(repo *model.Repo) error {
	if err := adapter.db.Save(repo).Error; err != nil {
		return errors.WithStack(err)
	}
	var tmp []string
	if err := json.Unmarshal([]byte(repo.Content), &tmp); err != nil {
		return errors.WithStack(err)
	}
	RepoMap[repo.API+";"+repo.Info] = tmp
	return nil
}

/**
 * @brief 查询仓库缓存记录。
 * @param repo 查询条件。
 * @return *model.DBList 查询结果集合。
 * @return error 查询失败时返回错误。
 */
func (adapter *SQLiteAdapter) ReadRepo(repo model.Repo) (*model.DBList, error) {
	var repos []model.Repo
	var err error
	switch {
	case repo.UUID != "":
		err = adapter.db.Where("uuid = ?", repo.UUID).Find(&repos).Error
	case repo.API != "":
		err = adapter.db.Where("api=? AND info=?", repo.API, repo.Info).Find(&repos).Error
	default:
		err = adapter.db.Find(&repos).Error
	}
	if len(repos) == 0 {
		err = gorm.ErrRecordNotFound
	}
	ret := model.DBList{
		RepoList: repos,
	}
	return &ret, errors.WithStack(err)
}

/**
 * @brief 删除仓库缓存记录并清理内存缓存。
 * @param repo 待删除的仓库记录。
 * @return error 删除失败时返回错误。
 */
func (adapter *SQLiteAdapter) DeleteRepo(repo *model.Repo) error {
	delete(RepoMap, repo.API+";"+repo.Info)
	return errors.WithStack(adapter.db.Delete(repo).Error)
}

/** @brief 创建仓库记录的包级代理函数。 */
func CreateRepo(repo *model.Repo) error { return localDB.CreateRepo(repo) }

/** @brief 更新仓库记录的包级代理函数。 */
func UpdateRepo(repo *model.Repo) error { return localDB.UpdateRepo(repo) }

/** @brief 查询仓库记录的包级代理函数。 */
func ReadRepo(repo model.Repo) (*model.DBList, error) { return localDB.ReadRepo(repo) }

/** @brief 删除仓库记录的包级代理函数。 */
func DeleteRepo(repo *model.Repo) error { return localDB.DeleteRepo(repo) }
