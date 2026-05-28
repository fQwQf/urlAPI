package middleware

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
	"urlAPI/internal/database"
	"urlAPI/util"

	"github.com/gin-gonic/gin"
)

const skipDBKey = "security.skipDB"
const generalKey = "security.general"

/**
 * @brief 请求安全上下文。
 *
 * 统一保存一次请求的来源信息、任务类型以及安全检查结果。
 */
type General struct {
	Referer string    `json:"referer"` //Complete Referer
	IP      string    `json:"ip"`
	Type    string    `json:"type"` // 任务类型
	Target  string    `json:"target"`
	Time    time.Time `json:"time"`
	Unsafe  bool      `json:"unsafe"`
	SkipDB  bool      `json:"skip_db"`
	Info    string    `json:"info"`
}

/**
 * @brief 构造通用安全中间件。
 * @param kind 请求所属的业务类型。
 * @return gin.HandlerFunc Gin 中间件函数。
 */
func GeneralSecurityMiddleware(kind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		general := General{
			Referer: c.Request.Referer(),
			IP:      c.ClientIP(),
			Time:    time.Now(),
		}
		switch kind {
		case "txt":
			general.Type = util.TypeMap["txt.gen"]
			general.Target = c.Query("prompt")
			general.Info = c.Query("more")
		case "img":
			general.Type = util.TypeMap["img.gen"]
			general.Target = c.Query("prompt")
			general.Info = c.Query("more")
		case "rand":
			general.Type = util.TypeMap["rand"]
			general.Target = c.Query("user") + "/" + c.Query("repo")
			general.Info = c.Query("more")
		case "web":
			general.Type = util.TypeMap["web.img"]
			general.Target = c.Query("img")
			general.Info = c.Query("more")
		case "download":
			general.Type = util.TypeMap["download"]
			general.Target = c.Query("img")
			general.SkipDB = true
		}
		checkGeneralSecurity(&general)
		if general.Unsafe {
			log.Printf("%s from %s\n", general.Info, c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{"error": general.Info})
			c.Abort()
			return
		}
		c.Set(skipDBKey, general.SkipDB)
		c.Set(generalKey, general)
		c.Next()
	}
}

/**
 * @brief 文本生成请求安全校验中间件。
 * @return gin.HandlerFunc Gin 中间件函数。
 */
func TextSecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		general, ok := GetGeneral(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "security context missing"})
			return
		}
		if !general.Unsafe {
			settings := database.SettingsStore.Get()
			txtgenenabled := settings.Text.EnabledPromptKeys
			txtacceptprompt := settings.Text.AcceptedPromptGlob
			textAPIs := []string{"openai", "alibaba", "deepseek", "otherapi"}
			api := c.Query("api")
			var prompt string
			reason := ""
			if _, ok := database.PromptMap[general.Target]; ok {
				prompt = general.Target
			} else {
				prompt = "other"
			}
			switch {
			case !settings.Features.TextEnabled:
				reason = "Txt is not enabled"
			case !util.ListChecker(&txtgenenabled, &prompt):
				reason = fmt.Sprintf("Target %s is not enabled for Txt Gen", general.Target)
			case (general.Target == "" || !util.WildcardChecker(&txtacceptprompt, &general.Target)) && prompt == "other":
				reason = fmt.Sprintf("Prompt %s is not enabled for Txt Gen", general.Target)
			case !util.ListChecker(&textAPIs, &api):
				reason = "Invalid API"
			}
			if reason != "" {
				general.Info = reason
				general.Unsafe = true
			}
		}
		abortIfUnsafe(c, general)
	}
}

/**
 * @brief 图像生成请求安全校验中间件。
 * @return gin.HandlerFunc Gin 中间件函数。
 */
func ImageSecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		general, ok := GetGeneral(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "security context missing"})
			return
		}
		if !general.Unsafe {
			settings := database.SettingsStore.Get()
			imgacceptprompt := settings.Image.AcceptedPromptGlob
			imgAPIs := []string{"openai", "alibaba"}
			api := c.Query("api")
			reason := ""
			if !settings.Features.ImageEnabled {
				reason = "Img is not enabled"
			} else if general.Target == "" || !util.WildcardChecker(&imgacceptprompt, &general.Target) {
				reason = fmt.Sprintf("Prompt %s is not allowed for ImgGen", general.Target)
			} else if !util.ListChecker(&imgAPIs, &api) {
				reason = "Invalid API"
			}
			if reason != "" {
				general.Info = reason
				general.Unsafe = true
			}
		}
		abortIfUnsafe(c, general)
	}
}

/**
 * @brief 随机图片请求安全校验中间件。
 * @return gin.HandlerFunc Gin 中间件函数。
 */
func RandomSecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		general, ok := GetGeneral(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "security context missing"})
			return
		}
		if !general.Unsafe {
			settings := database.SettingsStore.Get()
			randAPIs := []string{"github", "gitee"}
			api := c.Query("api")
			reason := ""
			if !settings.Features.RandomEnabled {
				reason = "Random is not enabled"
			} else if !util.ListChecker(&randAPIs, &api) {
				reason = "Invalid API"
			}
			if reason != "" {
				general.Info = reason
				general.Unsafe = true
			}
		}
		abortIfUnsafe(c, general)
	}
}

/**
 * @brief 网页截图请求安全校验中间件。
 * @return gin.HandlerFunc Gin 中间件函数。
 */
func WebSecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		general, ok := GetGeneral(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "security context missing"})
			return
		}
		parsedURL, _ := url.Parse(general.Target)
		if !general.Unsafe {
			settings := database.SettingsStore.Get()
			webimgallowed := settings.Web.AllowedHosts
			api := parsedURL.Host
			reason := ""
			switch {
			case !settings.Features.WebImgEnabled:
				reason = "WebImg is not enabled"
			case !util.ListChecker(&webimgallowed, &api):
				reason = fmt.Sprintf("API %s is not enabled", api)
			case api == "www.ithome.com" && !settings.Features.TextEnabled:
				reason = "For ITHome, TxtSum is not enabled"
			}
			if reason != "" {
				general.Info = reason
				general.Unsafe = true
			}
		}
		abortIfUnsafe(c, general)
	}
}

/**
 * @brief 下载请求安全中间件。
 * @return gin.HandlerFunc Gin 中间件函数。
 */
func DownloadSecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

/**
 * @brief 从上下文中提取通用安全信息。
 * @param c Gin 请求上下文。
 * @return General 安全上下文对象。
 * @return bool 是否成功读取。
 */
func GetGeneral(c *gin.Context) (General, bool) {
	value, exists := c.Get(generalKey)
	if !exists {
		return General{}, false
	}
	general, ok := value.(General)
	return general, ok
}

/**
 * @brief 从上下文中读取是否跳过数据库记录。
 * @param c Gin 请求上下文。
 * @return bool 是否跳过数据库写入。
 */
func GetSkipDB(c *gin.Context) bool {
	value, exists := c.Get(skipDBKey)
	if !exists {
		return false
	}
	skipDB, ok := value.(bool)
	return ok && skipDB
}

/**
 * @brief 根据安全检查结果决定是否中止请求。
 * @param c Gin 请求上下文。
 * @param general 通用安全信息。
 */
func abortIfUnsafe(c *gin.Context, general General) {
	if general.Unsafe {
		log.Printf("%s from %s\n", general.Info, c.ClientIP())
		c.JSON(http.StatusForbidden, gin.H{"error": general.Info})
		c.Abort()
		return
	}
	c.Set(skipDBKey, general.SkipDB)
	c.Set(generalKey, general)
	c.Next()
}
