package handles

import (
	"time"
	"urlAPI/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"urlAPI/internal/op"
	"urlAPI/internal/server/middleware"
	"urlAPI/util"
)

/**
 * @brief 处理图像生成请求。
 * @param c Gin 请求上下文。
 */
func ImgHandler(c *gin.Context) {
	referer := c.Request.Referer()
	ip := c.ClientIP()
	host := getScheme(c) + c.Request.Host
	modelName := c.Query("model")
	size := c.Query("size")
	task := model.Task{
		UUID:     uuid.New().String(),
		Time:     time.Now(),
		IP:       ip,
		Type:     util.TypeMap["img.gen"],
		Target:   c.Query("prompt"),
		Region:   util.GetRegion(ip),
		Referer:  referer,
		Device:   util.GetDeviceType(c.GetHeader("User-Agent")),
		API:      c.Query("api"),
		Model:    modelName,
		Size:     size,
		MoreInfo: c.Query("more"),
	}
	_, result, err := op.GenerateImage(task, host, middleware.GetSkipDB(c))
	util.ErrorPrinter(err)
	if err != nil {
		errorReturner(c, err)
		return
	}
	returner(c, result)
}
