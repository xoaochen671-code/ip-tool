# IP 类型识别功能 - 完整实现方案

## 目标效果

```text
 [DONE]  Target: Localhost
 ─────────────────────────────────────────
  IPv4      : 192.168.1.1 [Private]
  IPv6      : 2001:4860:4860::8888 [Public]

 (Press 4/6 to copy, q to quit)
```

或带详细信息：

```text
 [DONE]  Target: google.com
 ─────────────────────────────────────────
  IPv4      : 8.8.8.8 [Public] [Google DNS]
  IPv6      : 2001:4860:4860::8888 [Public]

  [ GEOLOCATION ]
  ISP       : Google LLC
  ASN       : AS15169
  Location  : Mountain View, California, US

  [ ATTRIBUTES ]
  Mobile Net   : No
  Proxy/VPN    : No
  Data Center  : Yes

 (Press 4/6 to copy, q to quit)
```

---

## 实现步骤

### 步骤 1: 创建 IP 类型识别模块

创建文件：`internal/scanner/iptype.go`

```go
package scanner

import (
	"net"
	"strings"
)

// IPType 表示 IP 地址的类型
type IPType string

const (
	TypePublic     IPType = "Public"
	TypePrivate    IPType = "Private"
	TypeLoopback   IPType = "Loopback"
	TypeLinkLocal  IPType = "Link-Local"
	TypeMulticast  IPType = "Multicast"
	TypeUnspecified IPType = "Unspecified"
	TypeInvalid    IPType = "Invalid"
)

// ClassifyIP 分类 IP 地址类型
func ClassifyIP(ipStr string) IPType {
	ipStr = strings.TrimSpace(ipStr)
	
	// 处理特殊情况
	if ipStr == "" || ipStr == "Not Detected" || ipStr == "Not Applicable" {
		return TypeInvalid
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return TypeInvalid
	}

	// 按优先级检查
	if ip.IsUnspecified() {
		return TypeUnspecified // 0.0.0.0 或 ::
	}
	
	if ip.IsLoopback() {
		return TypeLoopback // 127.0.0.0/8 或 ::1
	}
	
	if ip.IsPrivate() {
		return TypePrivate // 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, fc00::/7
	}
	
	if ip.IsLinkLocalUnicast() {
		return TypeLinkLocal // 169.254.0.0/16 或 fe80::/10
	}
	
	if ip.IsMulticast() {
		return TypeMulticast // 224.0.0.0/4 或 ff00::/8
	}

	return TypePublic
}

// GetIPTypeColor 获取 IP 类型对应的颜色（用于 TUI 高亮）
func GetIPTypeColor(ipType IPType) string {
	switch ipType {
	case TypePublic:
		return "green"   // 公网 IP - 绿色
	case TypePrivate:
		return "yellow"  // 私网 IP - 黄色
	case TypeLoopback:
		return "blue"    // 回环地址 - 蓝色
	case TypeLinkLocal:
		return "cyan"    // 链路本地 - 青色
	default:
		return "white"   // 其他 - 白色
	}
}

// GetIPDescription 获取 IP 类型的描述信息
func GetIPDescription(ipStr string) string {
	ipType := ClassifyIP(ipStr)
	
	descriptions := map[IPType]string{
		TypePublic:      "Internet-routable",
		TypePrivate:     "RFC1918 private address",
		TypeLoopback:    "localhost",
		TypeLinkLocal:   "Non-routable link-local",
		TypeMulticast:   "Multicast address",
		TypeUnspecified: "Wildcard address",
	}
	
	if desc, ok := descriptions[ipType]; ok {
		return desc
	}
	return ""
}

// GetWellKnownIP 识别常见的公共 IP 地址
func GetWellKnownIP(ipStr string) string {
	wellKnown := map[string]string{
		// Google DNS
		"8.8.8.8":                 "Google DNS",
		"8.8.4.4":                 "Google DNS",
		"2001:4860:4860::8888":    "Google DNS",
		"2001:4860:4860::8844":    "Google DNS",
		
		// Cloudflare DNS
		"1.1.1.1":                 "Cloudflare DNS",
		"1.0.0.1":                 "Cloudflare DNS",
		"2606:4700:4700::1111":    "Cloudflare DNS",
		"2606:4700:4700::1001":    "Cloudflare DNS",
		
		// OpenDNS
		"208.67.222.222":          "OpenDNS",
		"208.67.220.220":          "OpenDNS",
		
		// Quad9
		"9.9.9.9":                 "Quad9 DNS",
		"149.112.112.112":         "Quad9 DNS",
		
		// 常见网关
		"192.168.1.1":             "Common Gateway",
		"192.168.0.1":             "Common Gateway",
		"10.0.0.1":                "Common Gateway",
	}
	
	ipStr = strings.TrimSpace(ipStr)
	if name, ok := wellKnown[ipStr]; ok {
		return name
	}
	return ""
}

// FormatIPWithType 格式化 IP 地址带类型标签
// 例如: "8.8.8.8 [Public]" 或 "192.168.1.1 [Private]"
func FormatIPWithType(ipStr string) string {
	if ipStr == "" || ipStr == "Not Detected" || ipStr == "Not Applicable" {
		return ipStr
	}

	ipType := ClassifyIP(ipStr)
	if ipType == TypeInvalid {
		return ipStr
	}

	// 基础格式：IP [类型]
	formatted := ipStr + " [" + string(ipType) + "]"
	
	// 如果是知名 IP，添加名称
	if wellKnown := GetWellKnownIP(ipStr); wellKnown != "" {
		formatted += " [" + wellKnown + "]"
	}
	
	return formatted
}
```

---

### 步骤 2: 更新 IPInfo 结构体添加 ASN 信息

修改文件：`internal/scanner/client.go`

在 `IPInfo` 结构体中添加 ASN 字段：

```go
type IPInfo struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	Country    string `json:"country"`
	RegionName string `json:"regionName"`
	City       string `json:"city"`
	ISP        string `json:"isp"`
	AS         string `json:"as"`      // 新增：ASN 编号，如 "AS15169 Google LLC"
	Mobile     bool   `json:"mobile"`
	Proxy      bool   `json:"proxy"`
	Hosting    bool   `json:"hosting"`
}
```

注意：ip-api.com 默认就会返回 AS 字段，不需要修改 API 调用。

---

### 步骤 3: 更新 TUI 显示逻辑

修改文件：`internal/tui/model.go`

找到 `View()` 方法中 IP 地址显示的部分，修改为：

```go
// 导入 scanner 包中的新函数
import (
	"fmt"
	"github/shawn/ip-tool/internal/scanner"
	"net"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// 在 View() 方法中修改 IP 显示部分：

func (m *model) View() string {
	var b strings.Builder

	// 顶部状态
	status := "[DONE]"
	if m.loading {
		status = m.spinner.View() + " Fetching..."
	}

	b.WriteString(fmt.Sprintf("\n %s  Target: %s\n", status, m.getDisplayTarget()))
	b.WriteString(" ─────────────────────────────────────────\n")

	// 2. IP 地址显示 - 使用新的格式化函数
	b.WriteString(fmt.Sprintf("  %-10s: %s\n", "IPv4", m.formatIPWithType(m.ipv4)))
	b.WriteString(fmt.Sprintf("  %-10s: %s\n\n", "IPv6", m.formatIPWithType(m.ipv6)))

	// 3. 详细信息显示
	if m.isDetail {
		if m.ipInfo.Status == "success" {
			b.WriteString("  [ GEOLOCATION ]\n")
			b.WriteString(fmt.Sprintf("  %-10s: %s\n", "ISP", m.ipInfo.ISP))
			
			// 显示 ASN 信息（如果有）
			if m.ipInfo.AS != "" {
				b.WriteString(fmt.Sprintf("  %-10s: %s\n", "ASN", m.ipInfo.AS))
			}

			var locParts []string
			if m.ipInfo.City != "" {
				locParts = append(locParts, m.ipInfo.City)
			}
			if m.ipInfo.RegionName != "" && m.ipInfo.RegionName != m.ipInfo.City {
				locParts = append(locParts, m.ipInfo.RegionName)
			}
			if m.ipInfo.Country != "" {
				locParts = append(locParts, m.ipInfo.Country)
			}

			locationStr := strings.Join(locParts, ", ")
			if locationStr == "" {
				locationStr = "(unknown)"
			}

			b.WriteString(fmt.Sprintf("  %-10s: %s\n", "Location", locationStr))

			b.WriteString("\n  [ ATTRIBUTES ]\n")
			b.WriteString(fmt.Sprintf("  %-12s : %s\n", "Mobile Net", renderAttr(m.ipInfo.Mobile)))
			b.WriteString(fmt.Sprintf("  %-12s : %s\n", "Proxy/VPN", renderAttr(m.ipInfo.Proxy)))
			b.WriteString(fmt.Sprintf("  %-12s : %s\n", "Data Center", renderAttr(m.ipInfo.Hosting)))

		} else if m.loading && m.ipInfo.Status == "" {
			b.WriteString("[DOING] Locating your target...")
		} else if m.ipInfo.Status == "fail" {
			b.WriteString("  ❌ Could not resolve geolocation info.")
		}
		b.WriteString("\n")
	}
	
	// 4. 底部帮助
	if m.message != "" {
		b.WriteString(fmt.Sprintf("\n  %s\n", m.message))
	} else {
		b.WriteString("\n (Press 4/6 to copy, q to quit)\n")
	}

	return b.String()
}

// 添加新的辅助方法
func (m *model) formatIPWithType(ip string) string {
	if ip == "" {
		return "..."
	}
	if ip == "Not Detected" || ip == "Not Applicable" {
		return ip
	}
	
	// 使用 scanner 包的格式化函数
	return scanner.FormatIPWithType(ip)
}
```

---

### 步骤 4: 测试用例

创建测试文件：`internal/scanner/iptype_test.go`

```go
package scanner

import "testing"

func TestClassifyIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected IPType
	}{
		// Public IPv4
		{"8.8.8.8", TypePublic},
		{"1.1.1.1", TypePublic},
		
		// Private IPv4
		{"192.168.1.1", TypePrivate},
		{"10.0.0.1", TypePrivate},
		{"172.16.0.1", TypePrivate},
		
		// Loopback
		{"127.0.0.1", TypeLoopback},
		{"::1", TypeLoopback},
		
		// Link-Local
		{"169.254.1.1", TypeLinkLocal},
		{"fe80::1", TypeLinkLocal},
		
		// Public IPv6
		{"2001:4860:4860::8888", TypePublic},
		
		// Invalid
		{"", TypeInvalid},
		{"Not Detected", TypeInvalid},
		{"invalid", TypeInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := ClassifyIP(tt.ip)
			if result != tt.expected {
				t.Errorf("ClassifyIP(%s) = %s, want %s", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestGetWellKnownIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected string
	}{
		{"8.8.8.8", "Google DNS"},
		{"1.1.1.1", "Cloudflare DNS"},
		{"9.9.9.9", "Quad9 DNS"},
		{"192.168.1.1", "Common Gateway"},
		{"1.2.3.4", ""}, // 未知 IP
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := GetWellKnownIP(tt.ip)
			if result != tt.expected {
				t.Errorf("GetWellKnownIP(%s) = %s, want %s", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestFormatIPWithType(t *testing.T) {
	tests := []struct {
		ip       string
		expected string
	}{
		{"8.8.8.8", "8.8.8.8 [Public] [Google DNS]"},
		{"192.168.1.1", "192.168.1.1 [Private] [Common Gateway]"},
		{"127.0.0.1", "127.0.0.1 [Loopback]"},
		{"1.2.3.4", "1.2.3.4 [Public]"},
		{"Not Detected", "Not Detected"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := FormatIPWithType(tt.ip)
			if result != tt.expected {
				t.Errorf("FormatIPWithType(%s) = %s, want %s", tt.ip, result, tt.expected)
			}
		})
	}
}
```

---

## 使用效果演示

### 场景 1: 查询本机 IP（公网）

```bash
$ ipq
```

输出：
```text
 [DONE]  Target: Localhost
 ─────────────────────────────────────────
  IPv4      : 203.0.113.42 [Public]
  IPv6      : 2001:db8::1 [Public]

 (Press 4/6 to copy, q to quit)
```

### 场景 2: 查询 Google DNS

```bash
$ ipq 8.8.8.8 -d
```

输出：
```text
 [DONE]  Target: 8.8.8.8
 ─────────────────────────────────────────
  IPv4      : 8.8.8.8 [Public] [Google DNS]
  IPv6      : Not Applicable

  [ GEOLOCATION ]
  ISP       : Google LLC
  ASN       : AS15169
  Location  : Mountain View, California, US

  [ ATTRIBUTES ]
  Mobile Net   : No
  Proxy/VPN    : No
  Data Center  : Yes

 (Press 4/6 to copy, q to quit)
```

### 场景 3: 查询私网地址

```bash
$ ipq 192.168.1.1
```

输出：
```text
 [DONE]  Target: 192.168.1.1
 ─────────────────────────────────────────
  IPv4      : 192.168.1.1 [Private] [Common Gateway]
  IPv6      : Not Applicable

 (Press 4/6 to copy, q to quit)
```

### 场景 4: 查询域名

```bash
$ ipq google.com -d
```

输出：
```text
 [DONE]  Target: google.com
 ─────────────────────────────────────────
  IPv4      : 142.250.185.46 [Public]
  IPv6      : 2404:6800:4008:c06::8a [Public]

  [ GEOLOCATION ]
  ISP       : Google LLC
  ASN       : AS15169
  Location  : Mountain View, California, US

  [ ATTRIBUTES ]
  Mobile Net   : No
  Proxy/VPN    : No
  Data Center  : Yes

 (Press 4/6 to copy, q to quit)
```

---

## 实现检查清单

- [ ] 创建 `internal/scanner/iptype.go` 文件
- [ ] 在 `internal/scanner/client.go` 中添加 `AS` 字段
- [ ] 在 `internal/tui/model.go` 中添加 `formatIPWithType()` 方法
- [ ] 修改 `View()` 方法中的 IP 显示逻辑
- [ ] 在 `View()` 方法中添加 ASN 显示
- [ ] 创建测试文件 `internal/scanner/iptype_test.go`
- [ ] 运行测试：`go test ./internal/scanner/...`
- [ ] 测试各种场景：公网 IP、私网 IP、域名、知名 DNS
- [ ] 确认复制功能仍然正常工作

---

## 预计工作量

- **编码时间**: 1-2 小时
- **测试时间**: 30 分钟
- **总计**: 约 2-2.5 小时

---

## 下一步可选增强

完成这个功能后，可以考虑：

1. **添加颜色高亮**（使用 lipgloss）
2. **添加反向 DNS 查询**（显示 PTR 记录）
3. **添加延迟测试**（ping 该 IP 的响应时间）
4. **添加更多知名 IP 识别**（CDN 节点等）

---

需要我帮你开始编写代码吗？或者你有什么疑问？
