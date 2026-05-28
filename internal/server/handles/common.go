package handles

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"urlAPI/internal/op"
)

/**
 * @brief 返回成功结果。
 *
 * 当请求指定 `format=json` 时返回 JSON，否则重定向到结果 URL。
 *
 * @param c Gin 请求上下文。
 * @param result 生成结果。
 */
func returner(c *gin.Context, result op.GenerateResult) {
	if c.Query("format") == "json" {
		c.JSON(http.StatusOK, result)
	} else {
		c.Redirect(http.StatusFound, result.URL)
	}
}

/**
 * @brief 返回统一错误响应。
 * @param c Gin 请求上下文。
 * @param err 待返回的错误对象。
 */
func errorReturner(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

/**
 * @brief 推断当前请求使用的协议。
 * @param c Gin 请求上下文。
 * @return string 协议前缀，格式为 `http://` 或 `https://`。
 */
func getScheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return `https://`
	}
	if scheme := c.GetHeader("X-Forwarded-Proto"); scheme != "" {
		return scheme + `://`
	}
	return `http://`
}
