/*
目标验证和提取模块

CLI Guidelines 原则 - 用户友好:
- 接受多种输入格式，减少用户负担
- 支持直接粘贴 URL，自动提取有效部分
- 严格验证，给出清晰的错误信息

使用场景:
- 用户从浏览器地址栏复制 URL
- 命令行参数验证
- 批量处理文件中的目标
*/
package ip

import (
	"net"
	"strings"
)

// ExtractFromURL 从 URL 中提取域名或 IP
//
// 支持格式:
//   - "https://github.com/user/repo" -> "github.com"
//   - "http://192.168.1.1:8080/path" -> "192.168.1.1"
//   - "http://[::1]:8080/path"       -> "::1"
//   - "google.com"                   -> "google.com" (原样返回)
func ExtractFromURL(input string) string {
	// 0. 移除首尾空白和引号 (echo 会包含引号)
	input = strings.TrimSpace(input)
	input = strings.Trim(input, `"'`)

	// 1. 移除协议前缀 (http://, https://, ftp:// 等)
	if idx := strings.Index(input, "://"); idx != -1 {
		input = input[idx+3:]
	}

	// 2. 移除路径部分 (/path/to/page)
	if idx := strings.Index(input, "/"); idx != -1 {
		input = input[:idx]
	}

	// 3. 分离端口号
	// net.SplitHostPort 统一处理 IPv4 和 IPv6:
	//   - "127.0.0.1:8080" -> "127.0.0.1"
	//   - "[::1]:8080"     -> "::1" (自动去除方括号)
	if host, _, err := net.SplitHostPort(input); err == nil {
		return host
	}

	// 4. 无端口时，处理可能残留的 IPv6 方括号
	// 例如 "http://[::1]/path" 处理后变成 "[::1]"
	input = strings.TrimPrefix(input, "[")
	input = strings.TrimSuffix(input, "]")

	return input
}

// IsValidTarget 验证是否为有效的 IP 地址或域名
//
// 验证规则:
// 1. IP 地址: 使用 net.ParseIP 验证 (支持 IPv4/IPv6)
// 2. 域名:
//   - 长度不超过 253 字符 (DNS 规范)
//   - 不包含空格
//   - 包含至少一个点 (localhost 除外)
//   - 只包含合法字符
func IsValidTarget(target string) bool {
	target = strings.TrimSpace(target)

	// 首先检查是否为 IP 地址
	if net.ParseIP(target) != nil {
		return true
	}

	// 以下验证域名格式

	// DNS 规范: 域名最大长度 253 字符
	if len(target) > 253 {
		return false
	}

	// 不允许空格
	if strings.Contains(target, " ") {
		return false
	}

	// 必须包含点 (localhost 例外)
	// 排除单词输入如 "google"
	if !strings.Contains(target, ".") && target != "localhost" {
		return false
	}

	// 字符白名单
	for _, ch := range target {
		if !isValidDomainChar(ch) {
			return false
		}
	}

	return true
}

// isValidDomainChar 检查是否为有效的域名字符
func isValidDomainChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '.' || ch == '-'
}
