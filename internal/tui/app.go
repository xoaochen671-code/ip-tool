/*
Package tui 提供交互式终端界面

依赖: internal/network, internal/output

基于 Bubble Tea 框架实现，遵循 Elm 架构:
- Model: 应用状态 (App 结构体)
- Update: 处理消息，更新状态
- View: 根据状态渲染界面

CLI Guidelines 原则 - 交互式设计:
- 清晰的状态指示 (加载动画)
- 响应式更新
- 直观的键位绑定

键位设计 (遵循 Unix 惯例):
- q/Ctrl+C: 退出
- r: 刷新
- d: 详情
- 4/6: 复制 IPv4/IPv6
*/
package tui

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github/shawn/ip-tool/internal/network"
	"github/shawn/ip-tool/internal/output"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// App TUI 应用状态
type App struct {
	target         string           // 查询目标
	ipv4           string           // IPv4 结果
	ipv6           string           // IPv6 结果
	geoInfo        *network.GeoInfo // 地理位置信息
	message        string           // 临时消息 (如 "Copied!")
	loading        bool             // 是否加载中
	showDetail     bool             // 是否显示详情
	fetchingDetail bool             // 是否正在获取详情
	spinner        spinner.Model    // 加载动画
}

// NewApp 创建新应用实例
func NewApp(target string, showDetail bool) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return &App{
		target:     target,
		loading:    true,
		showDetail: showDetail,
		spinner:    s,
	}
}

// 消息类型 (Bubble Tea 消息传递模式)
type (
	ipv4Msg   string           // IPv4 查询结果
	ipv6Msg   string           // IPv6 查询结果
	geoMsg    *network.GeoInfo // 地理位置结果
	geoErrMsg string           // 地理位置错误
	clearMsg  struct{}         // 清除临时消息
)

// Init 初始化应用
func (a *App) Init() tea.Cmd {
	cmds := []tea.Cmd{a.spinner.Tick}

	// 如果目标是 IP 地址，直接使用
	if ip := net.ParseIP(a.target); ip != nil {
		if ip.To4() != nil {
			a.ipv4 = a.target
			a.ipv6 = "Not Applicable"
		} else {
			a.ipv6 = a.target
			a.ipv4 = "Not Applicable"
		}
		if a.showDetail {
			cmds = append(cmds, a.fetchGeo(a.target))
		}
		a.updateLoading()
		return tea.Batch(cmds...)
	}

	// 域名或空，需要解析
	cmds = append(cmds,
		func() tea.Msg { return ipv4Msg(network.ResolveIPv4(a.target)) },
		func() tea.Msg { return ipv6Msg(network.ResolveIPv6(a.target)) },
	)
	return tea.Batch(cmds...)
}

// fetchGeo 创建获取地理位置的命令
func (a *App) fetchGeo(ip string) tea.Cmd {
	return func() tea.Msg {
		info, err := network.FetchGeoInfo(ip)
		if err != nil {
			return geoErrMsg(err.Error())
		}
		return geoMsg(info)
	}
}

// Update 处理消息，更新状态
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit

		case "r":
			return a, a.refresh()

		case "d":
			if !a.showDetail {
				a.showDetail = true
				targetIP := a.getValidIP()
				if targetIP != "" {
					a.loading = true
					return a, a.fetchGeo(targetIP)
				}
				a.geoInfo = &network.GeoInfo{Status: "fail", Message: "No valid IP"}
			}

		case "4":
			if a.ipv4 != "" && a.ipv4 != "Not Detected" && a.ipv4 != "Not Applicable" {
				clipboard.WriteAll(a.ipv4)
				a.message = "Copied IPv4!"
				return a, tea.Tick(2*time.Second, func(time.Time) tea.Msg { return clearMsg{} })
			}

		case "6":
			if a.ipv6 != "" && a.ipv6 != "Not Detected" && a.ipv6 != "Not Applicable" {
				clipboard.WriteAll(a.ipv6)
				a.message = "Copied IPv6!"
				return a, tea.Tick(2*time.Second, func(time.Time) tea.Msg { return clearMsg{} })
			}
		}

	case clearMsg:
		a.message = ""

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd

	case ipv4Msg:
		a.ipv4 = string(msg)
		a.updateLoading()
		if a.showDetail && !a.fetchingDetail && a.geoInfo == nil && a.ipv4 != "Not Detected" {
			a.fetchingDetail = true
			return a, a.fetchGeo(a.ipv4)
		}

	case ipv6Msg:
		a.ipv6 = string(msg)
		a.updateLoading()
		if a.showDetail && !a.fetchingDetail && a.geoInfo == nil && a.ipv6 != "Not Detected" {
			a.fetchingDetail = true
			return a, a.fetchGeo(a.ipv6)
		}

	case geoMsg:
		a.geoInfo = msg
		a.fetchingDetail = false
		a.updateLoading()

	case geoErrMsg:
		a.geoInfo = &network.GeoInfo{Status: "fail", Message: string(msg)}
		a.fetchingDetail = false
		a.updateLoading()
	}

	return a, nil
}

// refresh 刷新查询
func (a *App) refresh() tea.Cmd {
	a.ipv4 = ""
	a.ipv6 = ""
	a.geoInfo = nil
	a.loading = true
	a.fetchingDetail = false
	a.message = "Refreshing..."

	clearCmd := tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg { return clearMsg{} })

	if ip := net.ParseIP(a.target); ip != nil {
		if ip.To4() != nil {
			a.ipv4 = a.target
			a.ipv6 = "Not Applicable"
		} else {
			a.ipv6 = a.target
			a.ipv4 = "Not Applicable"
		}
		if a.showDetail {
			return tea.Batch(clearCmd, a.fetchGeo(a.target))
		}
		a.updateLoading()
		return clearCmd
	}

	return tea.Batch(
		clearCmd,
		func() tea.Msg { return ipv4Msg(network.ResolveIPv4(a.target)) },
		func() tea.Msg { return ipv6Msg(network.ResolveIPv6(a.target)) },
	)
}

// View 渲染界面
func (a *App) View() string {
	var b strings.Builder

	// 状态栏
	status := "[DONE]"
	if a.loading {
		status = a.spinner.View() + " Fetching..."
	}

	displayTarget := a.target
	if displayTarget == "" {
		displayTarget = "Localhost"
	}

	b.WriteString(fmt.Sprintf("\n %s  Target: %s\n", status, displayTarget))
	b.WriteString(" ─────────────────────────────────────────\n")

	// IP 地址
	b.WriteString(fmt.Sprintf("  %-10s: %s\n", "IPv4", output.FormatIPDisplay(a.ipv4)))
	b.WriteString(fmt.Sprintf("  %-10s: %s\n\n", "IPv6", output.FormatIPDisplay(a.ipv6)))

	// 详情
	if a.showDetail {
		if a.geoInfo != nil && a.geoInfo.IsSuccess() {
			b.WriteString("  [ GEOLOCATION ]\n")
			b.WriteString(fmt.Sprintf("  %-10s: %s\n", "ISP", a.geoInfo.ISP))
			b.WriteString(fmt.Sprintf("  %-10s: %s\n", "Location", buildLocation(a.geoInfo)))

			b.WriteString("\n  [ ATTRIBUTES ]\n")
			b.WriteString(fmt.Sprintf("  %-12s : %s\n", "Mobile Net", output.FormatBool(a.geoInfo.Mobile)))
			b.WriteString(fmt.Sprintf("  %-12s : %s\n", "Proxy/VPN", output.FormatBool(a.geoInfo.Proxy)))
			b.WriteString(fmt.Sprintf("  %-12s : %s\n", "Data Center", output.FormatBool(a.geoInfo.Hosting)))

		} else if a.loading && a.geoInfo == nil {
			b.WriteString(output.StyleHint.Render("  Fetching geolocation..."))
			b.WriteString("\n")

		} else if a.geoInfo != nil && a.geoInfo.IsFailed() {
			b.WriteString("\n")
			b.WriteString(output.StyleError.Render("  ✗ Geolocation failed"))
			b.WriteString("\n")
			if a.geoInfo.Message != "" {
				b.WriteString(output.StyleHint.Render("  " + a.geoInfo.Message))
				b.WriteString("\n")
			}
			b.WriteString(output.StyleWarning.Render("  → Press 'r' to retry"))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// 底部帮助
	if a.message != "" {
		b.WriteString(fmt.Sprintf("\n  %s\n", a.message))
	} else {
		var keys []string
		keys = append(keys, "r to refresh")
		if !a.showDetail {
			keys = append(keys, "d for detail")
		}
		keys = append(keys, "4/6 to copy", "q to quit")
		b.WriteString(fmt.Sprintf("\n (%s)\n", strings.Join(keys, ", ")))
	}

	return b.String()
}

// updateLoading 更新加载状态
func (a *App) updateLoading() {
	ipReady := a.ipv4 != "" && a.ipv6 != ""
	if !a.showDetail {
		a.loading = !ipReady
		return
	}
	a.loading = !(ipReady && a.geoInfo != nil)
}

// getValidIP 获取有效的 IP 地址
func (a *App) getValidIP() string {
	if a.ipv4 != "" && a.ipv4 != "Not Detected" && a.ipv4 != "Not Applicable" {
		return a.ipv4
	}
	if a.ipv6 != "" && a.ipv6 != "Not Detected" && a.ipv6 != "Not Applicable" {
		return a.ipv6
	}
	return ""
}

// buildLocation 构建位置字符串
func buildLocation(g *network.GeoInfo) string {
	var parts []string
	if g.City != "" {
		parts = append(parts, g.City)
	}
	if g.RegionName != "" && g.RegionName != g.City && g.RegionName != g.Country {
		parts = append(parts, g.RegionName)
	}
	if g.Country != "" && g.Country != g.City {
		parts = append(parts, g.Country)
	}
	if len(parts) == 0 {
		return "(unknown)"
	}
	return strings.Join(parts, ", ")
}
