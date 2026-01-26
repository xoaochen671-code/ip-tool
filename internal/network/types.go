/*
Package network 提供网络请求功能

依赖: internal/ip (仅此一个)

本包负责所有网络 I/O 操作:
- DNS 解析
- HTTP 请求获取公网 IP
- 调用地理位置 API

CLI Guidelines 原则 - 超时控制:
- 所有网络请求必须有超时
- 避免程序在网络问题时无限挂起
- 给用户明确的反馈
*/
package network

// GeoInfo 地理位置信息
//
// 对应 ip-api.com 的 JSON 响应
// fields=18600473 返回以下字段
type GeoInfo struct {
	Status     string `json:"status"`     // "success" 或 "fail"
	Message    string `json:"message"`    // 失败时的错误信息
	Country    string `json:"country"`    // 国家名称
	RegionName string `json:"regionName"` // 地区/省份
	City       string `json:"city"`       // 城市
	ISP        string `json:"isp"`        // 互联网服务提供商
	Mobile     bool   `json:"mobile"`     // 是否为移动网络
	Proxy      bool   `json:"proxy"`      // 是否为代理/VPN
	Hosting    bool   `json:"hosting"`    // 是否为数据中心
}

// IsSuccess 检查查询是否成功
func (g *GeoInfo) IsSuccess() bool {
	return g.Status == "success"
}

// IsFailed 检查查询是否失败
func (g *GeoInfo) IsFailed() bool {
	return g.Status == "fail"
}
