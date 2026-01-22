/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github/shawn/ip-tool/internal/scanner"
	"github/shawn/ip-tool/internal/tui"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	showDetail    bool
	fromClipboard bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ipq [target]",
	Short: "A powerful IP lookup tool",
	Long: `IPQ is a modern TUI tool for querying IP addresses and domain information.

Examples:
  ipq                      # Show your public IP
  ipq 8.8.8.8              # Query specific IP
  ipq google.com -d        # Query domain with details
  ipq -c                   # Query from clipboard
  ipq --from-clipboard -d  # Query from clipboard with details`,
	Args: func(cmd *cobra.Command, args []string) error {

		if fromClipboard && len(args) > 0 {
			return fmt.Errorf("cannot specify target when using --from-clipboard/-c flag")
		}

		if len(args) > 1 {
			return fmt.Errorf("too many arguments, expected at most 1")
		}
		return nil
	},
	RunE: printIP,
}

func printIP(cmd *cobra.Command, args []string) error {
	var target string

	// 从剪贴板读取
	if fromClipboard {
		content, err := clipboard.ReadAll()
		if err != nil {
			return fmt.Errorf("failed to read clipboard: %w", err)
		}

		content = strings.TrimSpace(content)
		if content == "" {
			return fmt.Errorf("clipboard is empty")
		}

		// 尝试从 URL 中提取域名
		extracted := scanner.ExtractTargetFromURL(content)

		// 验证提取后的内容是否是有效的 IP 或域名
		if !scanner.IsValidTarget(extracted) {
			return fmt.Errorf("clipboard content is not a valid IP address or domain name: %s", content)
		}

		target = extracted
	} else if len(args) > 0 {
		target = args[0]
	}

	p := tea.NewProgram(tui.InitialModel(target, showDetail))

	// 运行程序
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&showDetail, "detail", "d", false, "Show detailed geolocation and ISP info")
	rootCmd.Flags().BoolVarP(&fromClipboard, "from-clipboard", "c", false, "Read IP/domain from clipboard")
}
