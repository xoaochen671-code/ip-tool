package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

type DNSInfo struct {
	DNS struct {
		Geo string `json:"geo"`
		IP  string `json:"ip"`
	} `json:"dns"`
}
type App struct {
	IPv6    bool
	Info    bool
	DNS     bool
	InputIP string
	Domain  string
	IPInfo  IPInfo
	DNSInfo DNSInfo
}

type IPInfo struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	Continent  string `json:"continent"`
	Country    string `json:"country"`
	RegionName string `json:"regionName"`
	City       string `json:"city"`
	District   string `json:"district"`
	ISP        string `json:"isp"`
	Mobile     bool   `json:"mobile"`
	Proxy      bool   `json:"proxy"`
	Hosting    bool   `json:"hosting"`
}

func NewApp() *App {
	app := &App{}

	pflag.BoolVar(&app.IPv6, "ipv6", false, "")
	pflag.BoolVar(&app.Info, "info", false, "")
	pflag.BoolVar(&app.DNS, "dns", false, "")
	pflag.StringVar(&app.InputIP, "ip", "", "")
	pflag.StringVar(&app.Domain, "domain", "", "")

	pflag.Usage = func() {
		fmt.Println("Usage: ip_tool [options]")
		fmt.Println("Options:")
		fmt.Println("  --ipv6       Use IPv6 instead of IPv4")
		fmt.Println("  --ip <IP>    Query specified IP instead of local IP")
		fmt.Println("  --info       Show detailed IP information")
		fmt.Println("  --dns        Show DNS server information")
		fmt.Println("  -h, --help   Show this help message")
	}
	pflag.Parse()

	return app
}

func (a *App) run() {
	a.SetIP()

	if !a.Info && !a.DNS { // 只展示IP
		fmt.Printf("IP     : %s\n", a.InputIP)
		return
	}

	if a.Info { // 显示详细 IP 信息
		addr, err := a.ResolveIP()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving IP: %v\n", err)
			os.Exit(1)
		}
		a.IPInfo = addr
		a.PrintIPInfo()
	}

	if a.DNS { // 显示 DNS 信息
		dns, err := a.ResolveDNS()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving DNS: %v\n", err)
			os.Exit(1)
		}
		a.DNSInfo = dns
		if a.Info {
			fmt.Println()
		}

		a.PrintDNSInfo()
	}

}

func (a *App) SetIP() error {
	if a.InputIP != "" { // 如果用户已输入IP
		return nil
	}

	if a.Domain != "" { // 如果用户已输入域名
		ip, err := net.LookupIP(a.Domain)
		if err != nil {
			return fmt.Errorf("failed to resolve IP from domain %s: %v", a.Domain, err)
		}
		a.InputIP = ip[0].String()
		return nil
	}

	url := "https://ipv4.icanhazip.com"
	if a.IPv6 {
		url = "https://ipv6.icanhazip.com"
	}
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("IPv6 not available, falling back to IPv4...")
		url = "https://ipv4.icanhazip.com"
		resp, err = http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to get IP from both IPv6 and IPv4: %v", err)
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read IP response: %v", err)
	}
	a.InputIP = strings.TrimSpace(string(body))
	return nil
}

func (a *App) ResolveIP() (IPInfo, error) {
	addr := IPInfo{}
	url := fmt.Sprintf(
		"http://ip-api.com/json/%s?fields=18600473",
		a.InputIP,
	)
	resp, err := http.Get(url)
	if err != nil {
		return addr, fmt.Errorf("failed to resolve IP from %s: %v", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return addr, fmt.Errorf("failed to read resolve response: %v", err)
	}

	err = json.Unmarshal(body, &addr)
	if err != nil {
		return addr, fmt.Errorf("failed to parse resolve response: %v", err)
	}

	return addr, nil
}

func (a *App) ResolveDNS() (DNSInfo, error) {
	dns := DNSInfo{}
	url := "http://edns.ip-api.com/json"
	resp, err := http.Get(url)
	if err != nil {
		return dns, fmt.Errorf("failed to resolve DNS from %s: %v", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return dns, fmt.Errorf("failed to read DNS response: %v", err)
	}

	err = json.Unmarshal(body, &dns)
	if err != nil {
		return dns, fmt.Errorf("failed to parse DNS response: %v", err)
	}

	return dns, nil
}

func (a *App) PrintIPInfo() {
	fmt.Println("IP Information")
	fmt.Println("-----------------------")

	fmt.Printf("IP         : %s\n", a.InputIP)
	fmt.Printf("Scope      : %s\n\n", a.IPType())
	fmt.Printf("Continent  : %s\n", a.IPInfo.Continent)
	fmt.Printf("Country    : %s\n", a.IPInfo.Country)
	fmt.Printf("Region     : %s\n", emptyAsDash(a.IPInfo.RegionName))
	fmt.Printf("City       : %s\n", emptyAsDash(a.IPInfo.City))
	fmt.Printf("District   : %s\n\n", emptyAsDash(a.IPInfo.District))

	fmt.Printf("ISP        : %s\n", a.IPInfo.ISP)
	fmt.Printf("Mobile     : %s\n", yesNo(a.IPInfo.Mobile))
	fmt.Printf("Proxy      : %s\n", yesNo(a.IPInfo.Proxy))
	fmt.Printf("Hosting    : %s\n", yesNo(a.IPInfo.Hosting))

}

func (a *App) PrintDNSInfo() {
	fmt.Println("DNS Information")
	fmt.Println("-----------------------")

	fmt.Printf("DNS        : %s\n", a.DNSInfo.DNS.IP)
	fmt.Printf("Geo        : %s\n", a.DNSInfo.DNS.Geo)

}

func (a *App) IPType() string {
	ip := net.ParseIP(a.InputIP)
	if ip == nil {
		return "Invalid"
	}

	switch {
	case ip.IsLoopback():
		return "Loopback"
	case ip.IsLinkLocalUnicast():
		return "Link-Local"
	case ip.IsPrivate():
		return "Private"
	case ip.IsGlobalUnicast():
		return "Public"
	default:
		return "Unknown"
	}
}

func yesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func emptyAsDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func main() {
	app := NewApp()
	app.run()
}
