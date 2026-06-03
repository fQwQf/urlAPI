package op

import (
	"github.com/pkg/errors"
	"urlAPI/internal/model"
)

/**
 * @brief 统一处理后台会话相关操作。
 * @param request 前端提交的会话请求。
 * @param authSession 当前鉴权会话。
 * @return Session 处理后的响应对象。
 * @return error 鉴权或具体操作失败时返回错误。
 */
func HandleSession(request Session, authSession model.Session) (Session, error) {
	response := request
	if err := login(&response, &authSession); err != nil {
		return response, errors.WithStack(err)
	}

	var err error
	switch response.Operation {
	case "login":
	case "logout":
		err = logout(&authSession)
	case "exit":
		err = exit(&authSession)
	case "newRepo":
		err = newRepo(&response)
	case "refreshRepo":
		err = refreshRepo(&response)
	case "delRepo":
		err = delRepo(&response)
	case "fetchRepo":
		err = fetchRepo(&response)
	case "fetchTask":
		err = fetchTask(&response)
	case "fetchSettings":
		err = fetchSettings(&response)
	case "editSettings":
		err = editSettings(&response)
	case "fetchProviderModels":
		err = fetchProviderModels(&response)
	case "fetchAPIKeys":
		err = fetchAPIKeys(&response)
	case "createAPIKey":
		err = createAPIKey(&response)
	case "deleteAPIKey":
		err = deleteAPIKey(&response)
	case "updateAPIKey":
		err = updateAPIKey(&response)
	}
	if err != nil {
		return response, errors.WithStack(err)
	}
	return response, nil
}
