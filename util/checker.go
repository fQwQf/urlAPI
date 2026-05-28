package util

import (
	"github.com/dlclark/regexp2"
	"image/png"
	"os"
	"regexp"
	"strings"
)

/**
 * @brief 检查字符串是否命中规则列表。
 *
 * 支持普通字符串、带 `*` 的通配符，以及 `re:` 前缀的正则表达式。
 *
 * @param strs 规则列表。
 * @param str 待匹配字符串。
 * @return bool 是否命中任一规则。
 */
func WildcardChecker(strs *[]string, str *string) bool {
	for _, r := range *strs {
		if strings.HasPrefix(r, "re:") {
			pattern := r[3:]
			re := regexp2.MustCompile(pattern, 0)
			match, err := re.MatchString(*str)
			if err == nil && match {
				return true
			}
			continue
		}
		if strings.Contains(r, "*") {
			pattern := "^" + strings.ReplaceAll(regexp.QuoteMeta(r), `\*`, ".*") + "$"
			re := regexp2.MustCompile(pattern, 0)
			match, err := re.MatchString(*str)
			if err == nil && match {
				return true
			}
			continue
		}
		if r == *str {
			return true
		}
	}
	return false
}

/**
 * @brief 检查字符串是否在允许列表中。
 * @param apis 允许值列表。
 * @param api 待检查值。
 * @return bool 为空或命中允许列表时返回 true。
 */
func ListChecker(apis *[]string, api *string) bool {
	if *api == "" {
		return true
	}
	for _, a := range *apis {
		if a == *api {
			return true
		}
	}
	return false
}

/**
 * @brief 检查指定路径上的 PNG 文件是否可正常解码。
 * @param path 文件路径。
 * @return bool 文件存在且可解析为 PNG 时返回 true。
 */
func PngChecker(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	if _, err = png.Decode(file); err != nil {
		return false
	}
	return true
}
