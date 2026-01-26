/*
Package output 提供输出格式化功能

依赖: internal/ip, internal/network

CLI Guidelines 原则 - NO_COLOR:
- 尊重 NO_COLOR 环境变量 (https://no-color.org/)
- 某些用户因视觉障碍需要禁用颜色
- 某些终端不支持 ANSI 颜色码
- 颜色输出可能干扰日志处理和管道操作

颜色方案:
- 红色 (9): 错误信息
- 黄色 (11): 警告信息
- 绿色 (10): 成功/正面信息
- 灰色 (8): 提示/次要信息
- 亮蓝 (12): 建议/操作指引
*/
package output

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// 全局样式变量
// 在 init() 中根据 NO_COLOR 环境变量初始化
var (
	StyleError      lipgloss.Style // 错误样式 (红色加粗)
	StyleWarning    lipgloss.Style // 警告样式 (黄色)
	StyleSuccess    lipgloss.Style // 成功样式 (绿色)
	StyleHint       lipgloss.Style // 提示样式 (灰色斜体)
	StyleSuggestion lipgloss.Style // 建议样式 (亮蓝)
)

func init() {
	// CLI Guidelines: 检查 NO_COLOR 环境变量
	if os.Getenv("NO_COLOR") != "" {
		// 禁用所有颜色，使用空样式
		StyleError = lipgloss.NewStyle()
		StyleWarning = lipgloss.NewStyle()
		StyleSuccess = lipgloss.NewStyle()
		StyleHint = lipgloss.NewStyle()
		StyleSuggestion = lipgloss.NewStyle()
		return
	}

	// 正常彩色样式
	// 使用 ANSI 256 色彩码，兼容大多数终端
	StyleError = lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Bold(true)

	StyleWarning = lipgloss.NewStyle().
		Foreground(lipgloss.Color("11"))

	StyleSuccess = lipgloss.NewStyle().
		Foreground(lipgloss.Color("10"))

	StyleHint = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)

	StyleSuggestion = lipgloss.NewStyle().
		Foreground(lipgloss.Color("12"))
}
