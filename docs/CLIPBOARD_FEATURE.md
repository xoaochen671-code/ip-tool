# 剪贴板读取功能实现

## 功能说明

支持从剪贴板自动读取 IP 地址或域名进行查询，无需手动输入。

## 使用方式

```bash
# 方式 1: 完整参数
ipq --from-clipboard

# 方式 2: 短参数
ipq -c

# 方式 3: 结合详细模式
ipq -c -d
ipq --from-clipboard --detail
```

## 使用场景

1. **从日志中复制 IP**
   - 复制日志中的 IP 地址
   - 运行 `ipq -c` 快速查询

2. **从网页复制域名**
   - 浏览器中复制域名
   - 终端中 `ipq -c -d` 查看详细信息

3. **快速工作流**
   - 无需切换窗口输入
   - 提高工作效率

---

## 实现代码

### 步骤 1: 修改 cmd/root.go

```go
/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github/shawn/ip-tool/internal/tui"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	showDetail     bool
	fromClipboard  bool
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
		// 如果使用 -c/--from-clipboard 参数，不允许同时传入 target
		if fromClipboard && len(args) > 0 {
			return fmt.Errorf("cannot specify target when using --from-clipboard/-c flag")
		}
		// 否则最多只能有一个参数
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

		// 验证剪贴板内容是否是有效的 IP 或域名
		if !isValidTarget(content) {
			return fmt.Errorf("clipboard content is not a valid IP address or domain name: %s", content)
		}

		target = content
		fmt.Printf("📋 Reading from clipboard: %s\n\n", target)
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

// isValidTarget 验证目标是否是有效的 IP 地址或域名
func isValidTarget(target string) bool {
	target = strings.TrimSpace(target)
	
	// 检查是否是有效的 IP 地址
	if net.ParseIP(target) != nil {
		return true
	}

	// 检查是否是有效的域名
	// 简单验证：长度合理，包含字母/数字/点/连字符
	if len(target) > 253 {
		return false
	}
	
	// 域名基本格式验证
	if strings.Contains(target, " ") {
		return false
	}
	
	// 必须包含至少一个点（除非是 localhost 等特殊情况）
	if !strings.Contains(target, ".") && target != "localhost" {
		return false
	}

	// 检查是否包含非法字符
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// 定义标志
	rootCmd.Flags().BoolVarP(&showDetail, "detail", "d", false, "Show detailed geolocation and ISP info")
	rootCmd.Flags().BoolVarP(&fromClipboard, "from-clipboard", "c", false, "Read IP/domain from clipboard")
}
```

---

## 使用示例

### 场景 1: 从日志复制 IP 查询

1. 在日志文件或网页中复制 IP 地址：`8.8.8.8`
2. 运行命令：
```bash
$ ipq -c
📋 Reading from clipboard: 8.8.8.8

 [DONE]  Target: 8.8.8.8
 ─────────────────────────────────────────
  IPv4      : 8.8.8.8
  IPv6      : Not Applicable

 (Press 4/6 to copy, q to quit)
```

### 场景 2: 复制域名查询详细信息

1. 在浏览器中复制域名：`google.com`
2. 运行命令：
```bash
$ ipq -c -d
📋 Reading from clipboard: google.com

 [DONE]  Target: google.com
 ─────────────────────────────────────────
  IPv4      : 142.250.185.46
  IPv6      : 2404:6800:4008:c06::8a

  [ GEOLOCATION ]
  ISP       : Google LLC
  Location  : Mountain View, California, US

  [ ATTRIBUTES ]
  Mobile Net   : No
  Proxy/VPN    : No
  Data Center  : Yes

 (Press 4/6 to copy, q to quit)
```

### 场景 3: 剪贴板内容无效

```bash
$ ipq -c
Error: clipboard content is not a valid IP address or domain name: some random text
```

### 场景 4: 剪贴板为空

```bash
$ ipq -c
Error: clipboard is empty
```

### 场景 5: 不能同时使用 -c 和传入参数

```bash
$ ipq -c 8.8.8.8
Error: cannot specify target when using --from-clipboard/-c flag
```

---

## 功能特点

### 1. 智能验证
- ✅ 自动验证剪贴板内容是否为有效的 IP 或域名
- ✅ 支持 IPv4 和 IPv6 地址
- ✅ 支持域名格式验证
- ✅ 友好的错误提示

### 2. 用户体验
- ✅ 显示从剪贴板读取的内容
- ✅ 短参数 `-c` 快速使用
- ✅ 可以与 `-d` 参数结合使用
- ✅ 清晰的错误信息

### 3. 安全性
- ✅ 验证输入格式，防止注入
- ✅ 长度限制（域名最大 253 字符）
- ✅ 字符白名单验证

---

## 测试检查清单

### 测试用例

1. **有效 IPv4 地址**
   ```bash
   # 复制: 8.8.8.8
   ipq -c
   # 预期：成功查询
   ```

2. **有效 IPv6 地址**
   ```bash
   # 复制: 2001:4860:4860::8888
   ipq -c
   # 预期：成功查询
   ```

3. **有效域名**
   ```bash
   # 复制: google.com
   ipq -c
   # 预期：成功查询
   ```

4. **带子域名**
   ```bash
   # 复制: www.google.com
   ipq -c
   # 预期：成功查询
   ```

5. **localhost**
   ```bash
   # 复制: localhost
   ipq -c
   # 预期：成功查询
   ```

6. **空剪贴板**
   ```bash
   # 清空剪贴板
   ipq -c
   # 预期：错误提示 "clipboard is empty"
   ```

7. **无效内容**
   ```bash
   # 复制: "这是一段中文文本"
   ipq -c
   # 预期：错误提示 "not a valid IP address or domain name"
   ```

8. **带空格的内容**
   ```bash
   # 复制: "8.8.8.8 is a DNS server"
   ipq -c
   # 预期：错误提示（内容包含空格）
   ```

9. **结合详细模式**
   ```bash
   # 复制: 8.8.8.8
   ipq -c -d
   # 预期：显示详细信息
   ```

10. **冲突参数**
    ```bash
    # 复制: google.com
    ipq -c google.com
    # 预期：错误提示 "cannot specify target when using --from-clipboard"
    ```

---

## 常见问题

### Q1: 剪贴板为什么读取失败？

A: 可能的原因：
- 没有复制任何内容
- 剪贴板被其他程序锁定
- 权限问题（少见）

解决方法：
- 确保先复制内容
- 尝试重新复制
- 检查终端权限

### Q2: 为什么我的域名被识别为无效？

A: 域名验证规则：
- 长度不超过 253 字符
- 只能包含字母、数字、点、连字符
- 不能包含空格
- 通常需要包含至少一个点

### Q3: 支持从剪贴板读取多行内容吗？

A: 目前不支持。剪贴板内容会被 `TrimSpace()` 处理，如果包含换行符会被视为无效。
未来可以考虑支持批量查询。

### Q4: 可以配置剪贴板读取的快捷键吗？

A: 目前是通过 `-c` 参数触发。如果需要在 TUI 运行时从剪贴板读取，
可以在未来版本中添加快捷键（如 `p` for paste）。

---

## 进阶功能建议

### 1. 自动检测剪贴板
如果不传入任何参数且剪贴板有有效内容，自动使用：
```go
// 在 printIP 函数中
if target == "" {
    content, err := clipboard.ReadAll()
    if err == nil && isValidTarget(content) {
        target = strings.TrimSpace(content)
        fmt.Printf("💡 Auto-detected from clipboard: %s\n\n", target)
    }
}
```

### 2. 剪贴板历史
记录最近从剪贴板查询的内容：
```bash
ipq clipboard-history
```

### 3. 批量查询
支持剪贴板中的多个 IP/域名（用换行分隔）：
```bash
ipq -c --batch
```

---

## 实现时间估计

- **编码时间**: 30-45 分钟
- **测试时间**: 15-20 分钟
- **总计**: 约 1 小时

---

## 相关文件

- `cmd/root.go` - 主要修改文件
- `internal/tui/ipquery.go` - 无需修改（已支持传入 target）

---

需要我直接帮你修改代码吗？
