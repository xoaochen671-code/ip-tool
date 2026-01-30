/*
Package ip 提供 IP 地址分类和验证功能

这是架构的最底层包，不依赖任何其他 internal 包。

CLI Guidelines 原则 - 模块化设计:
- 保持包的单一职责
- 避免循环依赖
- 底层包不应该知道上层的存在

IP 类型分类的意义:
- 帮助用户理解 IP 地址的性质
- 私网 IP 无法查询地理位置，需要提前告知
- 特殊 IP (回环、链路本地) 有特定用途
*/
package ip

import (
	"net"
	"strings"
)

// Type 表示 IP 地址的类型
type Type string

// IP 类型常量
// 分类依据 IANA 特殊用途地址注册表
const (
	TypePublic      Type = "Public"      // 公网 IP - 可在互联网路由
	TypePrivate     Type = "Private"     // 私网 IP - RFC 1918
	TypeLoopback    Type = "Loopback"    // 回环地址 - 127.0.0.0/8, ::1
	TypeLinkLocal   Type = "Link-Local"  // 链路本地 - 169.254.0.0/16, fe80::/10
	TypeMulticast   Type = "Multicast"   // 组播地址 - 224.0.0.0/4, ff00::/8
	TypeUnspecified Type = "Unspecified" // 未指定 - 0.0.0.0, ::
	TypeInvalid     Type = "Invalid"     // 无效地址
)

var reservedBlocks []*net.IPNet

func init() {
	networks := []string{
		"100.64.0.0/10",      // Shared Address Space (CGNAT)
		"192.0.0.0/24",       // IETF Protocol Assignments
		"192.0.2.0/24",       // Documentation (TEST-NET-1)
		"198.18.0.0/15",      // Benchmarking (你提到的那个)
		"198.51.100.0/24",    // Documentation (TEST-NET-2)
		"203.0.113.0/24",     // Documentation (TEST-NET-3)
		"240.0.0.0/4",        // Reserved for Future Use
		"255.255.255.255/32", // Limited Broadcast
	}

	for _, n := range networks {
		_, block, _ := net.ParseCIDR(n)
		reservedBlocks = append(reservedBlocks, block)
	}
}

// Classify 分析 IP 地址并返回其类型
func Classify(ipStr string) Type {
	ipStr = strings.TrimSpace(ipStr)

	// 处理特殊占位符值
	if ipStr == "" || ipStr == "Not Detected" || ipStr == "Not Applicable" {
		return TypeInvalid
	}

	// 使用标准库解析，确保正确性
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return TypeInvalid
	}

	// 按优先级检查类型
	// 注意: switch 顺序很重要，某些地址可能同时满足多个条件
	switch {
	case ip.IsUnspecified():
		return TypeUnspecified
	case ip.IsLoopback():
		return TypeLoopback
	case ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast():
		return TypeLinkLocal
	case ip.IsMulticast():
		return TypeMulticast
	case ip.IsPrivate():
		return TypePrivate
	default:
		// 公网检查是否是特殊保留网段
	}
	// 检查你定义的特殊保留网段
	for _, block := range reservedBlocks {
		if block.Contains(ip) {
			return TypePrivate
		}
	}
	return TypeInvalid
}

// IsValid 检查字符串是否为有效 IP 地址
func IsValid(ipStr string) bool {
	return net.ParseIP(strings.TrimSpace(ipStr)) != nil
}

// IsIPv4 检查是否为 IPv4 地址
func IsIPv4(ipStr string) bool {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	return ip != nil && ip.To4() != nil
}

// IsIPv6 检查是否为 IPv6 地址
func IsIPv6(ipStr string) bool {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	return ip != nil && ip.To4() == nil
}
