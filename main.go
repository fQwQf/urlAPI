package main

import "urlAPI/cmd"

/**
 * @brief 程序主入口。
 *
 * 启动命令行参数解析并分派具体子命令。
 */
func main() {
	cmd.Execute()
}
