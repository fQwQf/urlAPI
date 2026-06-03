package handles

import (
	"log"
	"net/http"
	"urlAPI/internal/model"
	"urlAPI/internal/op"
	"urlAPI/util"

	"github.com/gin-gonic/gin"
)

/**
 * @brief 处理后台会话与配置管理请求。
 * @param c Gin 请求上下文。
 */
func SessionHandler(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	var request op.Session
	if err := c.ShouldBindJSON(&request); err != nil {
		util.ErrorPrinter(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	authSession := model.Session{
		Token: c.Request.Header.Get("Authorization"),
		Term:  request.LoginTerm,
	}
	request.SessionIP = c.ClientIP()
	request.SessionToken = c.Request.Header.Get("Authorization")

	response, err := op.HandleSession(request, authSession)
	if err != nil {
		log.Printf("%s from %s\n", err, c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}
