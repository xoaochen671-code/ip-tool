# ip_tool

A minimal CLI tool for querying IP and DNS information.

---

## Install

```bash
go build -o ip_tool
```

---

## Usage

```bash
ip_tool [options]
```

### Options

```text
--ipv6            Use IPv6 instead of IPv4
--ip <ip>         Query specified IP
--domain <domain> Resolve domain to IP
--info            Show detailed IP information
--dns             Show DNS server information
-h, --help        Show help
```

---

## Examples

```bash
# Show local public IP
ip_tool

# Show detailed IP info
ip_tool --info

# Query specified IP
ip_tool --ip 1.1.1.1 --info

# Query domain
ip_tool --domain google.com --info

# Show DNS info
ip_tool --dns

# IP + DNS
ip_tool --info --dns

# IPv6
ip_tool --ipv6 --info
```

---

## Output (example)

```text
IP Information
-----------------------
IP         : 8.8.8.8
Scope      : Public
Country    : United States
Region     : California
City       : Mountain View
ISP        : Google LLC
Mobile     : No
Proxy      : No
Hosting    : Yes
```

---

## Data Sources

* ip-api.com
* edns.ip-api.com
* icanhazip.com

---

## License

MIT
