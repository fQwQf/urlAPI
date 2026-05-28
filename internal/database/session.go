package database

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"urlAPI/internal/model"
)

/**
 * @brief 创建登录会话记录。
 * @param session 待创建的会话对象。
 * @return error 写入失败时返回错误。
 */
func (adapter *SQLiteAdapter) CreateSession(session *model.Session) error {
	SessionMap[session.Token] = *session
	return errors.WithStack(adapter.db.Create(&session).Error)

}

/**
 * @brief 更新登录会话记录。
 * @param session 待更新的会话对象。
 * @return error 更新失败时返回错误。
 */
func (adapter *SQLiteAdapter) UpdateSession(session *model.Session) error {
	SessionMap[session.Token] = *session
	return errors.WithStack(adapter.db.Save(session).Error)
}

/**
 * @brief 查询登录会话记录。
 * @param session 查询条件，当前仅使用 token。
 * @return *model.DBList 查询结果集合。
 * @return error 查询失败时返回错误。
 */
func (adapter *SQLiteAdapter) ReadSession(session model.Session) (*model.DBList, error) {
	var sessions []model.Session
	err := adapter.db.Where("token=?", session.Token).Find(&sessions).Error
	if len(sessions) == 0 {
		err = gorm.ErrRecordNotFound
	}
	ret := model.DBList{
		SessionList: sessions,
	}
	return &ret, errors.WithStack(err)
}

/**
 * @brief 删除登录会话记录。
 * @param session 待删除的会话对象。
 * @return error 删除失败时返回错误。
 */
func (adapter *SQLiteAdapter) DeleteSession(session *model.Session) error {
	delete(SessionMap, session.Token)
	return errors.WithStack(adapter.db.Delete(session).Error)
}

/** @brief 创建会话记录的包级代理函数。 */
func CreateSession(session *model.Session) error { return localDB.CreateSession(session) }

/** @brief 更新会话记录的包级代理函数。 */
func UpdateSession(session *model.Session) error { return localDB.UpdateSession(session) }

/** @brief 查询会话记录的包级代理函数。 */
func ReadSession(session model.Session) (*model.DBList, error) { return localDB.ReadSession(session) }

/** @brief 删除会话记录的包级代理函数。 */
func DeleteSession(session *model.Session) error { return localDB.DeleteSession(session) }
