/*
Package cli 提供 CLI 辅助功能

依赖: internal/ip, internal/output

CLI Guidelines 原则 - Exit Codes (退出码):
- 程序应该返回有意义的退出码
- 0 表示成功，非 0 表示失败
- 不同的错误类型应该有不同的退出码
- 便于脚本判断程序执行结果

Unix 退出码惯例:
- 0: 成功
- 1: 一般错误
- 2: 命令行参数错误 (bash 惯例)

使用示例:

	ipq 8.8.8.8 || echo "失败，退出码: $?"

	if ipq invalid 2>/dev/null; then
	  echo "成功"
	else
	  case $? in
	    2) echo "参数错误" ;;
	    3) echo "网络错误" ;;
	  esac
	fi
*/
package cli

import "os"

// 退出码常量
const (
	ExitSuccess      = 0 // 成功
	ExitGeneralError = 1 // 一般错误
	ExitInvalidArgs  = 2 // 参数无效
	ExitNetworkError = 3 // 网络错误
	ExitNotFound     = 4 // 资源未找到
	ExitTimeout      = 5 // 超时
)

// Exit 以指定退出码终止程序
func Exit(code int) {
	os.Exit(code)
}
