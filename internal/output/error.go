/*
错误格式化模块

CLI Guidelines 原则 - Human-first Design:
- 错误信息应该清晰、可操作
- 包含三个部分:
 1. 标题: 简洁描述问题
 2. 原因: 解释为什么出错
 3. 建议: 如何解决问题

输出格式示例:

	✗ Invalid clipboard content
	  Content: some random text

	  → Copy a valid IP (8.8.8.8) or domain (google.com)
*/
package output

import (
	"fmt"
	"strings"
)

// NewError 创建格式化的错误信息
//
// 参数:
//   - title: 错误标题 (必填)
//   - reason: 错误原因/详情 (可选，空则不显示)
//   - suggestion: 解决建议 (可选，空则不显示)
func NewError(title, reason, suggestion string) error {
	var b strings.Builder

	// 换行开头，与上文分隔
	b.WriteString("\n")

	// 错误标题 (红色加粗，带 ✗ 前缀)
	b.WriteString(StyleError.Render("✗ " + title))
	b.WriteString("\n")

	// 错误原因 (灰色斜体，缩进显示)
	if reason != "" {
		b.WriteString(StyleHint.Render("  " + reason))
		b.WriteString("\n")
	}

	// 解决建议 (亮蓝色，带 → 前缀)
	if suggestion != "" {
		b.WriteString("\n")
		b.WriteString(StyleSuggestion.Render("  → " + suggestion))
		b.WriteString("\n")
	}

	return fmt.Errorf("%s", b.String())
}
