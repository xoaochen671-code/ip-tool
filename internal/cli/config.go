/*
配置管理模块

CLI Guidelines 原则 - Configuration:
- 支持配置文件，允许用户自定义默认行为
- 遵循 XDG 基础目录规范
- 配置优先级: 命令行参数 > 环境变量 > 配置文件 > 默认值

配置文件位置 (按优先级):
1. $IPQ_CONFIG 环境变量指定的路径
2. ~/.config/ipq/config.yaml (XDG 规范)
3. ~/.ipq.yaml (简便路径)

配置示例:

	show_detail: true
	timeout: 10s
	api_source: ip-api
*/
package cli

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	ShowDetail bool   `yaml:"show_detail"` // 默认显示详情
	Timeout    string `yaml:"timeout"`     // 请求超时
	APISource  string `yaml:"api_source"`  // API 数据源
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		ShowDetail: false,
		Timeout:    "5s",
		APISource:  "ip-api",
	}
}

// LoadConfig 加载配置
//
// CLI Guidelines: 静默处理缺失的配置文件
func LoadConfig() *Config {
	config := DefaultConfig()

	// 1. 检查环境变量
	configPath := os.Getenv("IPQ_CONFIG")

	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return config
		}

		// 2. XDG 规范路径
		configPath = filepath.Join(home, ".config", "ipq", "config.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// 3. 简便路径
			configPath = filepath.Join(home, ".ipq.yaml")
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				return config
			}
		}
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config
	}

	// 解析 YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return DefaultConfig()
	}

	return config
}
