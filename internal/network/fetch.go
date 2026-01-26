/*
HTTP 请求模块

提供获取公网 IP 和地理位置信息的功能。

CLI Guidelines 原则 - 健壮性:
- 所有请求有超时控制
- 区分不同类型的错误 (超时、网络错误、API 错误)
- 错误信息用户友好

数据源:
- 公网 IP: icanhazip.com (简单可靠)
- 地理位置: ip-api.com (免费，无需 API 密钥)
*/
package network

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// FetchPublicIPv4 获取本机公网 IPv4
func FetchPublicIPv4() (string, error) {
	return fetchIP("https://ipv4.icanhazip.com")
}

// FetchPublicIPv6 获取本机公网 IPv6
func FetchPublicIPv6() (string, error) {
	return fetchIP("https://ipv6.icanhazip.com")
}

// fetchIP 从指定 URL 获取 IP 地址
//
// 这些服务返回纯文本格式的 IP 地址
func fetchIP(url string) (string, error) {
	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	// 执行请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// 区分超时和其他网络错误
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout")
		}
		return "", fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

// FetchGeoInfo 查询 IP 地理位置
//
// 使用 ip-api.com 服务
// 限制: 每分钟 45 次请求 (非商业用途足够)
func FetchGeoInfo(ip string) (*GeoInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// 构造 API URL
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=18600473", strings.TrimSpace(ip))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout (>5s)")
		}
		return nil, fmt.Errorf("API unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d)", resp.StatusCode)
	}

	// 解析 JSON 响应
	var info GeoInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	// 检查 API 级别的错误
	if info.IsFailed() {
		info.Message = friendlyError(info.Message)
		return &info, fmt.Errorf("lookup failed: %s", info.Message)
	}

	return &info, nil
}

// friendlyError 将 API 错误转换为用户友好的描述
func friendlyError(msg string) string {
	switch msg {
	case "private range":
		return "Private IP (no geolocation)"
	case "reserved range":
		return "Reserved IP (no geolocation)"
	case "invalid query":
		return "Invalid IP format"
	default:
		if msg == "" {
			return "Unknown error"
		}
		return msg
	}
}
