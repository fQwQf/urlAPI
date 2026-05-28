package handles

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"urlAPI/internal/op"
	"urlAPI/util"
)

/**
 * @brief 下载生成后的图片资源。
 * @param c Gin 请求上下文。
 */
func DownloadHandler(c *gin.Context) {
	img, fallbackURL, err := op.DownloadImage(c.Query("img"))
	util.ErrorPrinter(err)
	downloadReturn(c, img, fallbackURL)
}

/**
 * @brief 统一返回下载结果或回退地址。
 * @param c Gin 请求上下文。
 * @param img 图片字节内容。
 * @param fallbackURL 回退地址，非空时执行重定向。
 */
func downloadReturn(c *gin.Context, img []byte, fallbackURL string) {
	if fallbackURL != "" {
		c.Redirect(http.StatusFound, fallbackURL)
		return
	}
	c.Header("Content-Type", "image/png")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="download.png"`))
	c.Header("Accept-Length", fmt.Sprintf("%d", len(img)))
	c.Writer.Write(img)
}
