# IPQ - IP Query Tool

A modern TUI tool for querying IP addresses, domains, and geolocation information.

## Features

- Query your public IPv4 and IPv6 addresses
- Look up any IP address or domain (auto-extracts from URLs)
- IP type identification (Public, Private, Loopback, etc.)
- Read from clipboard with `-c` flag
- Interactive TUI with refresh and detail toggle
- Geolocation and ISP information
- One-key copy to clipboard

## Installation

```bash
go build -o ipq
```

## Usage

```bash
ipq [target] [flags]
```

**Flags:**
- `-c, --from-clipboard` - Read IP/domain from clipboard
- `-d, --detail` - Show detailed geolocation info
- `-h, --help` - Show help

**Examples:**

```bash
ipq                    # Show your public IP
ipq 8.8.8.8           # Query specific IP
ipq google.com -d     # Query domain with details
ipq -c                # Query from clipboard
```

## Quick Demo

```bash
$ ipq 8.8.8.8 -d
```

```text
 [DONE]  Target: 8.8.8.8
 ─────────────────────────────────────────
  IPv4      : 8.8.8.8 [Public] [Google DNS]
  IPv6      : Not Applicable

  [ GEOLOCATION ]
  ISP       : Google LLC
  Location  : Mountain View, California, US

  [ ATTRIBUTES ]
  Mobile Net   : No
  Proxy/VPN    : No
  Data Center  : Yes

 (r to refresh, 4/6 to copy, q to quit)
```

## Interactive Keys

| Key | Action |
|-----|--------|
| `r` | Refresh/re-query |
| `d` | Toggle detail mode |
| `4` | Copy IPv4 to clipboard |
| `6` | Copy IPv6 to clipboard |
| `q` | Quit |

## Smart Features

- **URL Support**: Paste `https://github.com`, auto-extracts `github.com`
- **IP Type**: Automatically identifies Public, Private, Loopback IPs
- **Known IPs**: Recognizes Google DNS, Cloudflare DNS, etc.
- **Clipboard**: Copy any URL/IP/domain, use `ipq -c`

## License

MIT
