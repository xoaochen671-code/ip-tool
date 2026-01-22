package scanner

import (
	"net"
	"strings"
)

func ExtractTargetFromURL(input string) string {
	input = strings.TrimSpace(input)

	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		input = strings.TrimPrefix(input, "http://")
		input = strings.TrimPrefix(input, "https://")
		if idx := strings.Index(input, "/"); idx != -1 {
			input = input[:idx]
		}

		colonCount := strings.Count(input, ":")
		if colonCount == 1 {
			if idx := strings.Index(input, ":"); idx != -1 {
				input = input[:idx]
			}
		}
	}

	return input
}

// IsValidTarget 验证目标是否是有效的 IP 地址或域名
func IsValidTarget(target string) bool {
	target = strings.TrimSpace(target)

	if net.ParseIP(target) != nil {
		return true
	}

	if len(target) > 253 {
		return false
	}

	if strings.Contains(target, " ") {
		return false
	}

	if !strings.Contains(target, ".") && target != "localhost" {
		return false
	}

	for _, ch := range target {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '.' || ch == '-' || ch == ':') {
			return false
		}
	}

	return true
}
