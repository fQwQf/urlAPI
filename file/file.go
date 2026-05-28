package file

import "embed"

/** @brief 内嵌字体资源文件系统。 */
//go:embed ssfonts.ttf
var Font embed.FS

/** @brief 内嵌图标资源文件系统。 */
//go:embed icon/*
var Icons embed.FS

/** @brief 内嵌 Logo 资源文件系统。 */
//go:embed logo/*
var Logos embed.FS

/** @brief 内嵌配置模板文件系统。 */
//go:embed setting.json settings.json
var Settings embed.FS

/** @brief 内嵌空白占位图片文件系统。 */
//go:embed empty.png
var EmptyPNG embed.FS
