package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

func GetIP(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		// catch the error
		return "Not Detected", fmt.Errorf("failed to get IP: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// catch the error
		return "Not Detected", fmt.Errorf("failed to read response body: %w", err)
	}

	return strings.TrimSpace(string(body)), nil
}

func LookupDomain(host string, qType string) (string, error) {
	resolver := net.DefaultResolver

	// 根据 qType 决定调用哪个方法
	switch strings.ToUpper(qType) {
	case "A":
		// 仅查询 IPv4
		ips, err := resolver.LookupIP(context.Background(), "ip4", host)
		if err != nil || len(ips) == 0 {
			return "Not Detected", err
		}
		return ips[0].String(), nil

	case "AAAA":
		// 仅查询 IPv6
		ips, err := resolver.LookupIP(context.Background(), "ip6", host)
		if err != nil || len(ips) == 0 {
			return "Not Detected", err
		}
		return ips[0].String(), nil

	case "CNAME":
		cname, err := resolver.LookupCNAME(context.Background(), host)
		if err != nil {
			return "Not Detected", err
		}
		return cname, nil

	default:
		return "", fmt.Errorf("unsupported query type: %s", qType)
	}
}

// FetchV4 根据是否有 domain 决定从 DNS 还是 HTTP 获取 IPv4
func FetchV4(domain string) string {
	if domain != "" {
		res, _ := LookupDomain(domain, "A")
		return res
	}
	res, _ := GetIP("https://ipv4.icanhazip.com")
	return res
}

// FetchV6 根据是否有 domain 决定获取 IPv6
func FetchV6(domain string) string {
	if domain != "" {
		res, _ := LookupDomain(domain, "AAAA")
		return res
	}
	res, _ := GetIP("https://ipv6.icanhazip.com")
	return res
}

func ResolveIP(ip string) (IPInfo, error) {
	addr := IPInfo{}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	cleanIP := strings.TrimSpace(ip)
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=18600473", cleanIP)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return addr, fmt.Errorf("create request: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return addr, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return addr, fmt.Errorf("http status: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&addr); err != nil {
		return addr, fmt.Errorf("decode json: %w", err)
	}

	if addr.Status == "fail" {
		return addr, fmt.Errorf("api error: %s", addr.Message)
	}

	return addr, nil
}

type IPInfo struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	Country    string `json:"country"`
	RegionName string `json:"regionName"`
	City       string `json:"city"`
	ISP        string `json:"isp"`
	Mobile     bool   `json:"mobile"`
	Proxy      bool   `json:"proxy"`
	Hosting    bool   `json:"hosting"`
}
