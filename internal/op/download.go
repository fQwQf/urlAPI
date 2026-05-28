package op

import (
	"os"
	"urlAPI/file"
	"urlAPI/internal/database"
)

/**
 * @brief 读取已生成图片或内置占位图。
 * @param target 图片标识或特殊值 `empty`。
 * @return []byte 图片字节内容。
 * @return string 失败时可回退的图片地址。
 * @return error 读取失败时返回错误。
 */
func downloadImage(target string) ([]byte, string, error) {
	var img []byte
	var err error
	switch target {
	case "empty":
		img, err = file.EmptyPNG.ReadFile("empty.png")
	default:
		img, err = os.ReadFile(ImgPath + target + ".png")
	}
	if err != nil {
		return nil, database.SettingsStore.Get().Web.FallbackImageURL, err
	}
	return img, "", nil
}
