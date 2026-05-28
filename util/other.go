package util

import (
	"fmt"
	"log"
	"strings"
)

/**
 * @brief 批量替换字符串列表中的指定片段。
 * @param list 待处理的字符串列表指针。
 * @param old 待替换的旧子串。
 * @param new 替换后的新子串。
 */
func ListReplacer(list *[]string, old string, new string) {
	var ret []string
	for _, item := range *list {
		ret = append(ret, strings.Replace(item, old, new, -1))
	}
	*list = ret
}

/**
 * @brief 打印错误及其与本项目相关的堆栈信息。
 * @param err 待输出的错误对象。
 */
func ErrorPrinter(err error) {
	if err != nil {
		log.Println(err)
		stackTrace := fmt.Sprintf("%+v", err)
		lines := strings.Split(stackTrace, "\n")
		for _, line := range lines {
			if strings.Contains(line, "urlAPI") {
				log.Println(line)
			}
		}
	}
}

/** @brief 需要在仓库文件遍历时排除的文件后缀列表。 */
var excludedFiles = []string{
	".gitignore",
	".DS_Store",
	".ini",
	".yml",
	".yaml",
	".md",
	".txt",
	".json",
	".xml",
	".csv",
	".log",
}

/**
 * @brief 过滤掉不适合处理的文件链接。
 * @param links 原始链接列表。
 * @return []string 过滤后的链接列表。
 */
func LinkFilter(links []string) []string {
	var ret []string
	for _, link := range links {
		excluded := false
		for _, exclude := range excludedFiles {
			if strings.HasSuffix(link, exclude) {
				excluded = true
				break
			}
		}
		if !excluded {
			ret = append(ret, link)
		}
	}
	return ret
}
