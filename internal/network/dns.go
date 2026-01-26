/*
DNS 解析模块

提供域名到 IP 地址的解析功能。

CLI Guidelines 原则 - 超时控制:
- 所有 DNS 查询都有 5 秒超时
- 避免慢速 DNS 服务器导致程序挂起
*/
package network

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// defaultTimeout 所有网络操作的默认超时
const defaultTimeout = 5 * time.Second

// LookupIPv4 查询域名的 IPv4 地址 (A 记录)
func LookupIPv4(host string) (string, error) {
	return lookupIP(host, "ip4")
}

// LookupIPv6 查询域名的 IPv6 地址 (AAAA 记录)
func LookupIPv6(host string) (string, error) {
	return lookupIP(host, "ip6")
}

// lookupIP 执行 DNS 查询
//
// 参数:
//   - host: 要查询的域名
//   - network: "ip4" 或 "ip6"
//
// 返回:
//   - 成功: 第一个 IP 地址
//   - 失败: 错误信息
func lookupIP(host, network string) (string, error) {
	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// 使用系统默认 DNS 解析器
	ips, err := net.DefaultResolver.LookupIP(ctx, network, host)
	if err != nil {
		return "", fmt.Errorf("DNS lookup failed: %w", err)
	}

	// 检查是否有结果
	if len(ips) == 0 {
		return "", fmt.Errorf("no %s address found", strings.ToUpper(network))
	}

	// 返回第一个结果
	return ips[0].String(), nil
}

// LookupCNAME 查询 CNAME 记录
func LookupCNAME(host string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cname, err := net.DefaultResolver.LookupCNAME(ctx, host)
	if err != nil {
		return "", fmt.Errorf("CNAME lookup failed: %w", err)
	}
	return cname, nil
}
