package tui

import (
	"fmt"
	"github/shawn/ip-tool/internal/scanner"
	"net"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type ipQueryApp struct {
	target   string
	ipv4     string
	ipv6     string
	ipInfo   scanner.IPInfo
	message  string
	loading  bool
	isIP     bool
	isDetail bool
	spinner  spinner.Model
}

func InitialModel(target string, isDetail bool, isIP bool) *ipQueryApp {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return &ipQueryApp{
		target:   target,
		loading:  true,
		isIP:     isIP,
		isDetail: isDetail,
		spinner:  s,
	}
}

func (q *ipQueryApp) Init() tea.Cmd {
	cmds := []tea.Cmd{q.spinner.Tick}
	if q.isIP {
		if ipObj := net.ParseIP(q.target); ipObj.To4() != nil {
			q.ipv4 = q.target
			q.ipv6 = "Not Applicable"
		} else {
			q.ipv6 = q.target
			q.ipv4 = "Not Applicable"
		}

		if q.isDetail {
			cmds = append(cmds, q.fetchDetailCmd(q.target))
		}

		q.checkLoading()
		return tea.Batch(cmds...)
	}
	cmds = append(cmds,
		func() tea.Msg { return v4Msg(scanner.FetchV4(q.target)) },
		func() tea.Msg { return v6Msg(scanner.FetchV6(q.target)) },
	)

	return tea.Batch(cmds...)
}

func (q *ipQueryApp) fetchDetailCmd(ip string) tea.Cmd {
	return func() tea.Msg {
		info, _ := scanner.ResolveIP(ip)
		return detailMsg(info)
	}
}

type v4Msg string
type v6Msg string
type detailMsg scanner.IPInfo

func (q *ipQueryApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return q, tea.Quit

		case "4":
			if q.ipv4 != "" && q.ipv4 != "Not Detected" {
				clipboard.WriteAll(q.ipv4)
				q.message = "Copied IPv4 to clipboard!"
			}

		case "6":
			if q.ipv6 != "" && q.ipv6 != "Not Detected" {
				clipboard.WriteAll(q.ipv6)
				q.message = "Copied IPv6 to clipboard!"
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		q.spinner, cmd = q.spinner.Update(msg)
		return q, cmd
	case v4Msg:
		q.ipv4 = string(msg)
		q.checkLoading()
		if !q.isIP && q.isDetail && q.ipv4 != "Not Detected" && q.ipInfo.Status == "" {
			return q, q.fetchDetailCmd(q.ipv4)
		}
		return q, nil
	case v6Msg:
		q.ipv6 = string(msg)
		q.checkLoading()
		if !q.isIP && q.isDetail && q.ipv4 == "Not Detected" && q.ipv6 != "Not Detected" && q.ipInfo.Status == "" {
			return q, q.fetchDetailCmd(q.ipv6)
		}
		return q, nil

	case detailMsg:
		q.ipInfo = scanner.IPInfo(msg)
		q.checkLoading()
		return q, nil
	}

	return q, nil
}

func (q *ipQueryApp) View() string {
	var b strings.Builder

	// 顶部状态：只有还没全部取完时才显示小菊花
	status := "[DONE]"
	if q.loading {
		status = q.spinner.View() + " Fetching..."
	}

	b.WriteString(fmt.Sprintf("\n %s  Target: %s\n", status, q.getDisplayTarget()))
	b.WriteString(" ─────────────────────────────────────────\n")

	// 2. IP 地址显示
	b.WriteString(fmt.Sprintf("  %-10s: %s\n", "IPv4", formatVal(q.ipv4)))
	b.WriteString(fmt.Sprintf("  %-10s: %s\n\n", "IPv6", formatVal(q.ipv6)))

	// 3. 详细信息显示 (仅在 isDetail 模式开启时)
	if q.isDetail {
		if q.ipInfo.Status == "success" {
			b.WriteString("  [ GEOLOCATION ]\n")
			b.WriteString(fmt.Sprintf("  %-10s: %s\n", "ISP", q.ipInfo.ISP))

			var locParts []string
			if q.ipInfo.City != "" {
				locParts = append(locParts, q.ipInfo.City)
			}
			// 只有当省份与城市不同时才添加，避免 "Shanghai, Shanghai"
			if q.ipInfo.RegionName != "" && q.ipInfo.RegionName != q.ipInfo.City {
				locParts = append(locParts, q.ipInfo.RegionName)
			}
			if q.ipInfo.Country != "" {
				locParts = append(locParts, q.ipInfo.Country)
			}

			locationStr := strings.Join(locParts, ", ")
			if locationStr == "" {
				locationStr = "(unknown)"
			}

			b.WriteString(fmt.Sprintf("  %-10s: %s\n", "Location", locationStr))

			b.WriteString("\n  [ ATTRIBUTES ]\n")

			b.WriteString(fmt.Sprintf("  %-12s : %s\n", "Mobile Net", renderAttr(q.ipInfo.Mobile)))
			b.WriteString(fmt.Sprintf("  %-12s : %s\n", "Proxy/VPN", renderAttr(q.ipInfo.Proxy)))
			b.WriteString(fmt.Sprintf("  %-12s : %s\n", "Data Center", renderAttr(q.ipInfo.Hosting)))

		} else if q.loading && q.ipInfo.Status == "" {
			b.WriteString("[DOING] Locating your target...")
		} else if q.ipInfo.Status == "fail" {
			b.WriteString("  ❌ Could not resolve geolocation info.")
		}
		b.WriteString("\n")
	}
	// 4. 底部帮助
	if q.message != "" {
		b.WriteString(fmt.Sprintf("\n  %s\n", q.message))
	} else {
		b.WriteString("\n (Press 4/6 to copy, q to quit)\n")
	}

	return b.String()
}

func (q *ipQueryApp) checkLoading() {
	ipReady := q.ipv4 != "" && q.ipv6 != ""
	if !q.isDetail {
		// 普通模式：IP 好了就结束
		q.loading = !ipReady
		return
	}
	detailReady := q.ipInfo.Status != ""

	// 如果是域名模式，Detail 需要等待 IP 出来后触发，所以要综合判断
	q.loading = !(ipReady && detailReady)
}

func (q *ipQueryApp) getDisplayTarget() string {
	if q.target == "" {
		return "Localhost"
	}
	return q.target
}

func formatVal(v string) string {
	if v == "" {
		return "..."
	}
	return v
}
func renderAttr(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}
