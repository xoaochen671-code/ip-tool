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

type errMsg error

type model struct {
	target   string
	ipv4     string
	ipv6     string
	ipInfo   scanner.IPInfo
	message  string
	loading  bool
	isIP     bool
	isDetail bool
	spinner  spinner.Model
	err      error
}

func InitialModel(target string, isDetail bool, isIP bool) *model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return &model{
		target:   target,
		loading:  true,
		isIP:     isIP,
		isDetail: isDetail,
		spinner:  s,
	}
}

func (m *model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.spinner.Tick}
	if m.isIP {
		if ipObj := net.ParseIP(m.target); ipObj.To4() != nil {
			m.ipv4 = m.target
			m.ipv6 = "Not Applicable"
		} else {
			m.ipv6 = m.target
			m.ipv4 = "Not Applicable"
		}

		if m.isDetail {
			cmds = append(cmds, m.fetchDetailCmd(m.target))
		}

		m.checkLoading()
		return tea.Batch(cmds...)
	}
	cmds = append(cmds,
		func() tea.Msg { return v4Msg(scanner.FetchV4(m.target)) },
		func() tea.Msg { return v6Msg(scanner.FetchV6(m.target)) },
	)

	return tea.Batch(cmds...)
}

func (m *model) fetchDetailCmd(ip string) tea.Cmd {
	return func() tea.Msg {
		info, _ := scanner.ResolveIP(ip)
		return detailMsg(info)
	}
}

type v4Msg string
type v6Msg string
type detailMsg scanner.IPInfo

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "4":
			if m.ipv4 != "" && m.ipv4 != "Not Detected" {
				clipboard.WriteAll(m.ipv4)
				m.message = "Copied IPv4 to clipboard!"
			}

		case "6":
			if m.ipv6 != "" && m.ipv6 != "Not Detected" {
				clipboard.WriteAll(m.ipv6)
				m.message = "Copied IPv6 to clipboard!"
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case v4Msg:
		m.ipv4 = string(msg)
		m.checkLoading()
		if !m.isIP && m.isDetail && m.ipv4 != "Not Detected" && m.ipInfo.Status == "" {
			return m, m.fetchDetailCmd(m.ipv4)
		}
		return m, nil
	case v6Msg:
		m.ipv6 = string(msg)
		m.checkLoading()
		if !m.isIP && m.isDetail && m.ipv4 == "Not Detected" && m.ipv6 != "Not Detected" && m.ipInfo.Status == "" {
			return m, m.fetchDetailCmd(m.ipv6)
		}
		return m, nil

	case detailMsg:
		m.ipInfo = scanner.IPInfo(msg)
		m.checkLoading()
		return m, nil
	}

	return m, nil
}

func (m *model) View() string {
	var b strings.Builder

	// 顶部状态：只有还没全部取完时才显示小菊花
	status := "[DONE]"
	if m.loading {
		status = m.spinner.View() + " Fetching..."
	}

	b.WriteString(fmt.Sprintf("\n %s  Target: %s\n", status, m.getDisplayTarget()))
	b.WriteString(" ─────────────────────────────────────────\n")

	// 2. IP 地址显示
	b.WriteString(fmt.Sprintf("  %-10s: %s\n", "IPv4", formatVal(m.ipv4)))
	b.WriteString(fmt.Sprintf("  %-10s: %s\n\n", "IPv6", formatVal(m.ipv6)))

	// 3. 详细信息显示 (仅在 isDetail 模式开启时)
	if m.isDetail {
		if m.ipInfo.Status == "success" {
			b.WriteString("  [ GEOLOCATION ]\n")
			b.WriteString(fmt.Sprintf("  %-10s: %s\n", "ISP", m.ipInfo.ISP))

			var locParts []string
			if m.ipInfo.City != "" {
				locParts = append(locParts, m.ipInfo.City)
			}
			// 只有当省份与城市不同时才添加，避免 "Shanghai, Shanghai"
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

func (m *model) checkLoading() {
	ipReady := m.ipv4 != "" && m.ipv6 != ""
	if !m.isDetail {
		// 普通模式：IP 好了就结束
		m.loading = !ipReady
		return
	}
	detailReady := m.ipInfo.Status != ""

	// 如果是域名模式，Detail 需要等待 IP 出来后触发，所以要综合判断
	m.loading = !(ipReady && detailReady)
}

func (m *model) getDisplayTarget() string {
	if m.target == "" {
		return "Localhost"
	}
	return m.target
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
