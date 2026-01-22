package scanner

import (
	"net"
	"strings"
)

type IPType string

const (
	TypePublic      IPType = "Public"
	TypePrivate     IPType = "Private"
	TypeLoopback    IPType = "Loopback"
	TypeLinkLocal   IPType = "Link-Local"
	TypeMulticast   IPType = "Multicast"
	TypeUnspecified IPType = "Unspecified"
	TypeInvalid     IPType = "Invalid"
)

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

	return formatted
}
