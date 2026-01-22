# IPQ - IP Query Tool

A powerful and modern TUI (Terminal User Interface) tool for querying IP addresses, domain resolution, and geolocation information.

Built with [Cobra](https://github.com/spf13/cobra) and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

---

## Features

- Query your public IPv4 and IPv6 addresses
- Look up any IP address or domain
- Detailed geolocation and ISP information
- Quick copy to clipboard (press `4` for IPv4, `6` for IPv6)
- Beautiful interactive TUI with loading animations
- Fast and lightweight

---

## Installation

### Build from source

```bash
go build -o ipq
```

### Run directly

```bash
go run main.go
```

---

## Usage

```bash
ipq [target] [flags]
```

### Arguments

- `target` (optional): IP address or domain name to query
  - If omitted, shows your public IP addresses

### Flags

- `-d, --detail`: Show detailed geolocation and ISP information
- `-h, --help`: Show help message

---

## Examples

### Query your public IP

```bash
ipq
```

Output:
```text
 [DONE]  Target: Localhost
 ─────────────────────────────────────────
  IPv4      : 203.0.113.1
  IPv6      : 2001:db8::1

 (Press 4/6 to copy, q to quit)
```

### Query with detailed information

```bash
ipq -d
```

Output:
```text
 [DONE]  Target: Localhost
 ─────────────────────────────────────────
  IPv4      : 203.0.113.1
  IPv6      : 2001:db8::1

  [ GEOLOCATION ]
  ISP       : Example ISP
  Location  : Shanghai, China

  [ ATTRIBUTES ]
  Mobile Net   : No
  Proxy/VPN    : No
  Data Center  : No

 (Press 4/6 to copy, q to quit)
```

### Query a specific IP address

```bash
ipq 8.8.8.8 -d
```

### Query a domain

```bash
ipq google.com -d
```

---

## Interactive Controls

When the TUI is running:

- `4` - Copy IPv4 address to clipboard
- `6` - Copy IPv6 address to clipboard
- `q` or `Ctrl+C` - Quit the application

---

## Project Structure

```
ip_tool/
├── cmd/
│   └── root.go          # Cobra command definitions
├── internal/
│   ├── scanner/
│   │   └── client.go    # IP/domain resolution and API calls
│   └── tui/
│       └── model.go     # Bubble Tea TUI model
├── main.go              # Application entry point
├── go.mod
└── README.md
```

---

## Data Sources

- **ip-api.com** - Geolocation and ISP information
- **icanhazip.com** - Public IP detection (IPv4/IPv6)
- **System DNS** - Domain resolution

---

## Technologies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [atotto/clipboard](https://github.com/atotto/clipboard) - Cross-platform clipboard support

---

## License

MIT
