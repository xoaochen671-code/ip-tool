/*
批量处理模块

CLI Guidelines 原则 - Unix 哲学:
- 做好一件事，并能与其他工具组合使用
- 支持从文件或 stdin 批量处理

使用示例:

	# 从文件读取
	ipq -f ips.txt

	# 从 stdin 批量处理
	cat ips.txt | ipq --batch

	# 与其他工具组合
	grep "8.8" ips.txt | ipq --batch -o json
*/
package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github/shawn/ip-tool/internal/ip"
	"github/shawn/ip-tool/internal/output"

	"gopkg.in/yaml.v3"
)

// ProcessBatchFile 从文件批量处理
func ProcessBatchFile(filename string, detail bool, format output.Format, quiet bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	return processBatch(bufio.NewScanner(file), detail, format, quiet)
}

// ProcessBatchStdin 从 stdin 批量处理
func ProcessBatchStdin(detail bool, format output.Format, quiet bool) error {
	return processBatch(bufio.NewScanner(os.Stdin), detail, format, quiet)
}

// processBatch 批量处理核心逻辑
//
// 设计决策:
// 1. 跳过空行和注释 (# 开头)
// 2. 无效输入输出到 stderr，不中断处理
// 3. JSON 输出时收集所有结果后一次性输出
// 4. Text 输出时逐条输出
func processBatch(scanner *bufio.Scanner, detail bool, format output.Format, quiet bool) error {
	var results []*output.Result
	count := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 智能提取目标
		target := ip.ExtractFromURL(line)
		if !ip.IsValidTarget(target) {
			// CLI Guidelines: 警告输出到 stderr
			if !quiet {
				fmt.Fprintf(os.Stderr, "Skipping invalid: %s\n", line)
			}
			continue
		}

		// 根据格式决定处理方式
		if format == output.FormatJSON || format == output.FormatYAML {
			results = append(results, output.FetchResult(target, detail))
		} else {
			if count > 0 {
				fmt.Println()
			}
			output.Print(target, detail, format)
		}
		count++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read error: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("no valid targets found")
	}

	// JSON 批量输出: 数组格式
	if format == output.FormatJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	// YAML 批量输出: 数组格式
	if format == output.FormatYAML {
		enc := yaml.NewEncoder(os.Stdout)
		enc.SetIndent(2)
		return enc.Encode(results)
	}

	return nil
}
