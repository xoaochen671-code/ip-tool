/*
输出格式化模块

CLI Guidelines 原则 - 多种输出格式:
- TUI: 交互式界面，适合人类用户
- Text: 简单文本，适合非交互式环境
- JSON: 机器可读，便于程序解析
- YAML: 机器可读，更易人类阅读
- Quiet: 最小输出，便于管道处理

CLI Guidelines 原则 - 可组合性:
- 支持 JSON/YAML 输出便于与 jq、yq 等工具配合
- Quiet 模式便于管道: ipq google.com -q | xargs ping
*/
package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github/shawn/ip-tool/internal/ip"
	"github/shawn/ip-tool/internal/network"

	"gopkg.in/yaml.v3"
)

// Format 输出格式类型
type Format string

const (
	FormatTUI   Format = "tui"   // 交互式 TUI
	FormatText  Format = "text"  // 人类可读文本
	FormatJSON  Format = "json"  // 机器可读 JSON
	FormatYAML  Format = "yaml"  // 机器可读 YAML
	FormatQuiet Format = "quiet" // 最小输出
)

// Result 查询结果
type Result struct {
	Target  string  `json:"target" yaml:"target"`
	IPv4    string  `json:"ipv4" yaml:"ipv4"`
	IPv6    string  `json:"ipv6" yaml:"ipv6"`
	Type    string  `json:"type,omitempty" yaml:"type,omitempty"`
	Detail  *Detail `json:"detail,omitempty" yaml:"detail,omitempty"`
	Success bool    `json:"success" yaml:"success"`
	Error   string  `json:"error,omitempty" yaml:"error,omitempty"`
}

// Detail 详细信息
type Detail struct {
	ISP     string `json:"isp" yaml:"isp"`
	Country string `json:"country" yaml:"country"`
	Region  string `json:"region" yaml:"region"`
	City    string `json:"city" yaml:"city"`
	Mobile  bool   `json:"mobile" yaml:"mobile"`
	Proxy   bool   `json:"proxy" yaml:"proxy"`
	Hosting bool   `json:"hosting" yaml:"hosting"`
}

// FetchResult 获取查询结果
func FetchResult(target string, withDetail bool) *Result {
	result := &Result{
		Target:  target,
		Success: true,
	}

	// 空目标表示查询本机
	if target == "" {
		result.Target = "(localhost)"
	}

	// 获取 IP
	result.IPv4 = network.ResolveIPv4(target)
	result.IPv6 = network.ResolveIPv6(target)

	// 检测IP类型
	switch {
	case isValidIP(result.IPv4):
		result.Type = string(ip.Classify(result.IPv4))
	case isValidIP(result.IPv6):
		result.Type = string(ip.Classify(result.IPv6))
	default:
		result.Success = false
		result.Error = "Could not detect IP address"
	}

	// 获取详情
	if withDetail && result.Success {
		targetIP := result.IPv4
		if !isValidIP(targetIP) {
			targetIP = result.IPv6
		}
		if isValidIP(targetIP) {
			if info, err := network.FetchGeoInfo(targetIP); err == nil {
				result.Detail = &Detail{
					ISP:     info.ISP,
					Country: info.Country,
					Region:  info.RegionName,
					City:    info.City,
					Mobile:  info.Mobile,
					Proxy:   info.Proxy,
					Hosting: info.Hosting,
				}
			}
		}
	}

	return result
}

// isValidIP 检查是否为有效 IP 值
func isValidIP(s string) bool {
	return s != "" && s != "Not Detected" && s != "Not Applicable"
}

// Print 输出结果
func Print(target string, detail bool, format Format) error {
	result := FetchResult(target, detail)

	switch format {
	case FormatJSON:
		return printJSON(result)
	case FormatYAML:
		return printYAML(result)
	case FormatQuiet:
		return printQuiet(result)
	default:
		return printText(result, detail)
	}
}

// printJSON 输出 JSON 格式
func printJSON(r *Result) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ") // 缩进便于人类阅读
	return enc.Encode(r)
}

// printYAML 输出 YAML 格式
func printYAML(r *Result) error {
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	return enc.Encode(r)
}

// printQuiet 静默输出
//
// 例如: ipq google.com -q | xargs ping
func printQuiet(r *Result) error {
	if isValidIP(r.IPv4) {
		fmt.Println(r.IPv4)
	}
	if isValidIP(r.IPv6) {
		fmt.Println(r.IPv6)
	}
	return nil
}

// printText 输出文本格式
func printText(r *Result, detail bool) error {
	fmt.Printf("Target: %s\n", r.Target)
	fmt.Printf("IPv4: %s\n", formatIP(r.IPv4))
	fmt.Printf("IPv6: %s\n", formatIP(r.IPv6))

	if detail && r.Detail != nil {
		fmt.Println("---")
		fmt.Printf("ISP: %s\n", r.Detail.ISP)
		fmt.Printf("Location: %s, %s, %s\n",
			r.Detail.City, r.Detail.Region, r.Detail.Country)
		fmt.Printf("Mobile: %v | Proxy: %v | Hosting: %v\n",
			r.Detail.Mobile, r.Detail.Proxy, r.Detail.Hosting)
	}

	return nil
}

// formatIP 格式化 IP 地址输出
func formatIP(s string) string {
	if !isValidIP(s) {
		return "-"
	}
	return fmt.Sprintf("%s [%s]", s, ip.Classify(s))
}

// FormatIPDisplay 格式化 IP 显示 (供 TUI 使用)
func FormatIPDisplay(v string) string {
	if v == "" {
		return StyleHint.Render("...") // 加载中
	}
	if v == "Not Detected" {
		return StyleError.Render("Not Detected")
	}
	if v == "Not Applicable" {
		return StyleHint.Render("N/A")
	}
	return fmt.Sprintf("%s [%s]", v, ip.Classify(v))
}

// FormatBool 格式化布尔值显示
func FormatBool(v bool) string {
	if v {
		return StyleSuccess.Render("✓ Yes")
	}
	return StyleHint.Render("No")
}
