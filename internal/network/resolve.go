/*
统一解析接口

提供简化的 IP 解析功能，自动处理不同类型的输入。

设计决策:
- 返回 "Not Detected" 而非空字符串，便于 UI 显示
- 返回 "Not Applicable" 表示不适用 (如 IPv4 地址查询 IPv6)
- 隐藏错误细节，简化调用方代码
*/
package network

import "net"

// ResolveIPv4 获取目标的 IPv4 地址
//
// 智能处理:
// - target 为空: 获取本机公网 IPv4
// - target 是 IPv4: 直接返回
// - target 是 IPv6: 返回 "Not Applicable"
// - target 是域名: DNS 解析
func ResolveIPv4(target string) string {
	// 空目标 = 查询本机
	if target == "" {
		ip, err := FetchPublicIPv4()
		if err != nil {
			return "Not Detected"
		}
		return ip
	}

	// 检查是否为 IP 地址
	if ip := net.ParseIP(target); ip != nil {
		if ip.To4() != nil {
			return target // 是 IPv4，直接返回
		}
		return "Not Applicable" // 是 IPv6，不适用
	}

	// 是域名，DNS 解析
	ip, err := LookupIPv4(target)
	if err != nil {
		return "Not Detected"
	}
	return ip
}

// ResolveIPv6 获取目标的 IPv6 地址
//
// 智能处理:
// - target 为空: 获取本机公网 IPv6
// - target 是 IPv6: 直接返回
// - target 是 IPv4: 返回 "Not Applicable"
// - target 是域名: DNS 解析
func ResolveIPv6(target string) string {
	if target == "" {
		ip, err := FetchPublicIPv6()
		if err != nil {
			return "Not Detected"
		}
		return ip
	}

	if ip := net.ParseIP(target); ip != nil {
		if ip.To4() == nil {
			return target // 是 IPv6，直接返回
		}
		return "Not Applicable" // 是 IPv4，不适用
	}

	ip, err := LookupIPv6(target)
	if err != nil {
		return "Not Detected"
	}
	return ip
}
