/*
输入处理模块

CLI Guidelines 原则 - 环境感知:
- 自动检测是否在交互式终端运行
- 在管道/CI 环境中切换输出模式

CLI Guidelines 原则 - 可组合性:
- 支持从 stdin 读取输入
- 便于与其他工具配合: echo "8.8.8.8" | ipq
*/
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github/shawn/ip-tool/internal/ip"
)

// HasStdin 检测是否有管道输入
func HasStdin() bool {
	stat, _ := os.Stdin.Stat()
	// 原理: 终端是字符设备，管道不是
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// IsInteractive 检测是否在交互式终端运行
func IsInteractive() bool {
	// stdin 是管道
	if HasStdin() {
		return false
	}

	// stdout 是管道
	stat, _ := os.Stdout.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	// CI 环境
	if os.Getenv("CI") != "" {
		return false
	}

	return true
}

// ReadStdin 从 stdin 读取单个目标
func ReadStdin() (string, error) {
	if !HasStdin() {
		return "", nil
	}

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			return "", fmt.Errorf("empty input")
		}

		// 智能提取目标
		target := ip.ExtractFromURL(line)
		if !ip.IsValidTarget(target) {
			return "", fmt.Errorf("invalid target: %s", line)
		}
		return target, nil
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("no input")
}
