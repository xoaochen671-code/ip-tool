/*
Package cmd 提供 CLI 命令定义

CLI Guidelines 原则:

1. Human-first Design (人类优先设计)
  - 清晰的错误信息，包含问题、原因、建议
  - 无参数时有合理的默认行为

2. Composability (可组合性)
  - 支持从 stdin 读取
  - 支持多种输出格式

3. Robustness (健壮性)
  - 验证所有用户输入
  - 优雅处理错误

输入源优先级:

	剪贴板 (-c) > 命令行参数 > 标准输入 > 默认（本机）
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github/shawn/ip-tool/internal/cli"
	"github/shawn/ip-tool/internal/ip"
	"github/shawn/ip-tool/internal/output"
	"github/shawn/ip-tool/internal/tui"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// 命令行标志
var (
	showDetail    bool   // -d: 显示详情
	fromClipboard bool   // -c: 从剪贴板读取
	quiet         bool   // -q: 静默模式
	batch         bool   // --batch: 批量处理
	inputFile     string // -f: 输入文件
	outputFormat  string // -o: 输出格式
)

var rootCmd = &cobra.Command{
	Use:   "ipq [target]",
	Short: "Query IP addresses and domains",
	Long: `IPQ - A modern IP lookup tool with TUI interface.

EXAMPLES:
  ipq                    Query your public IP
  ipq 8.8.8.8 -d         Query with details
  ipq google.com         Query domain
  ipq -c                 Read from clipboard
  echo "8.8.8.8" | ipq   Read from stdin
  ipq -f ips.txt         Batch from file
  ipq 8.8.8.8 -o json    JSON output

ENVIRONMENT:
  NO_COLOR               Disable colors
  CI                     Force non-interactive mode
  IPQ_CONFIG             Config file path`,

	SilenceUsage:  true, // 错误时不打印用法
	SilenceErrors: true, // 错误由我们处理

	// 参数验证
	Args: func(cmd *cobra.Command, args []string) error {
		// -c 和参数冲突
		if fromClipboard && len(args) > 0 {
			return output.NewError(
				"Cannot use both clipboard and target argument",
				"",
				"ipq -c",
			)
		}
		// 最多一个参数
		if len(args) > 1 {
			return output.NewError(
				"Too many arguments",
				"",
				"ipq 8.8.8.8",
			)
		}
		return nil
	},

	RunE: run,
}

// run 主逻辑
func run(cmd *cobra.Command, args []string) error {
	format := getFormat()

	// 批量处理
	if inputFile != "" {
		return cli.ProcessBatchFile(inputFile, showDetail, format, quiet)
	}
	if batch && cli.HasStdin() {
		return cli.ProcessBatchStdin(showDetail, format, quiet)
	}

	// 获取目标
	target, err := getTarget(args)
	if err != nil {
		return err
	}

	// 输出
	if format == output.FormatTUI && cli.IsInteractive() {
		p := tea.NewProgram(tui.NewApp(target, showDetail))
		if _, err := p.Run(); err != nil {
			return output.NewError("Application error", err.Error(), "")
		}
	} else {
		return output.Print(target, showDetail, format)
	}

	return nil
}

// getTarget 获取查询目标
//
// 优先级: 剪贴板 > 参数 > stdin > 空 (本机)
func getTarget(args []string) (string, error) {
	// 从剪贴板
	if fromClipboard {
		content, err := clipboard.ReadAll()
		if err != nil {
			return "", output.NewError(
				"Failed to read clipboard",
				"",
				"Copy an IP or domain, then run: ipq -c",
			)
		}

		content = strings.TrimSpace(content)
		if content == "" {
			return "", output.NewError(
				"Clipboard is empty",
				"",
				"Copy an IP address or domain name first",
			)
		}

		target := ip.ExtractFromURL(content)
		if !ip.IsValidTarget(target) {
			return "", output.NewError(
				"Invalid clipboard content",
				fmt.Sprintf("Content: %s", content),
				"Copy a valid IP or domain",
			)
		}
		return target, nil
	}

	// 从参数
	if len(args) > 0 {
		return args[0], nil
	}

	// 从 stdin
	if cli.HasStdin() {
		target, err := cli.ReadStdin()
		if err != nil {
			return "", output.NewError(
				"Failed to read stdin",
				err.Error(),
				"echo '8.8.8.8' | ipq",
			)
		}
		return target, nil
	}

	// 空目标 = 查询本机
	return "", nil
}

// getFormat 确定输出格式
func getFormat() output.Format {
	switch outputFormat {
	case "json":
		return output.FormatJSON
	case "yaml":
		return output.FormatYAML
	case "text":
		return output.FormatText
	case "quiet":
		return output.FormatQuiet
	}

	if quiet {
		return output.FormatQuiet
	}

	// 非交互式环境自动使用文本
	if !cli.IsInteractive() {
		return output.FormatText
	}

	return output.FormatTUI
}

// Execute CLI 入口点
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// CLI Guidelines: 错误输出到 stderr
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cli.ExitGeneralError)
	}
}

// init 注册标志
func init() {
	// 查询选项
	rootCmd.Flags().BoolVarP(&showDetail, "detail", "d", false, "Show detailed info")
	rootCmd.Flags().BoolVarP(&fromClipboard, "from-clipboard", "c", false, "Read from clipboard")

	// 输入选项
	rootCmd.Flags().StringVarP(&inputFile, "file", "f", "", "Read targets from file")
	rootCmd.Flags().BoolVar(&batch, "batch", false, "Batch process from stdin")

	// 输出选项
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format: json, yaml, text, quiet")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Only output IP addresses")
}
