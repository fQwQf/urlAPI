package op

import (
	"github.com/pkg/errors"
	"time"
	"urlAPI/internal/database"
	"urlAPI/internal/model"
	"urlAPI/util"
)

/**
 * @brief 处理后台登录鉴权。
 * @param info 会话请求与响应对象。
 * @param data 当前鉴权会话数据。
 * @return error 鉴权失败时返回错误。
 */
func login(info *Session, data *model.Session) error {
	var session model.Session
	if info.Operation == "login" && database.SettingsStore.Get().Security.DashboardPasswordHash == data.Token {
		session.Token = util.GetRandomString()
		info.SessionToken = session.Token
		session.Term = info.LoginTerm
		if info.LoginTerm {
			session.Expire = time.Now().AddDate(0, 0, 7)
		} else {
			session.Expire = time.Now().AddDate(0, 0, 1)
		}
		if err := db.CreateSession(&session); err != nil {
			return err
		}
		return nil
	}
	var ok bool
	session, ok = database.SessionMap[data.Token]
	switch {
	case !ok:
		return errors.WithStack(errors.New("Authentication failed"))
	case time.Now().After(session.Expire):
		return errors.New("Expired token")
	default:
		return nil
	}
}

/**
 * @brief 执行登出并删除会话。
 * @param data 当前会话数据。
 * @return error 删除会话失败时返回错误。
 */
func logout(data *model.Session) error {
	if err := db.DeleteSession(data); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

/**
 * @brief 处理退出操作。
 *
 * 非持久会话在退出时会被删除，长期会话保留。
 *
 * @param data 当前会话数据。
 * @return error 删除会话失败时返回错误。
 */
func exit(data *model.Session) error {
	session, _ := database.SessionMap[data.Token]
	if !session.Term {
		if err := db.DeleteSession(data); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
