# IPQ - IP Query Tool

A modern CLI/TUI tool for querying IP addresses and domains.

## Features

- Query public IPv4/IPv6 addresses
- Look up any IP or domain
- IP type identification (Public/Private/Loopback)
- Geolocation and ISP information
- Multiple input sources: args, clipboard, stdin, file
- Multiple output formats: TUI, JSON, YAML, text
- Respects `NO_COLOR` and auto-detects non-interactive environments

## Installation

```bash
# Quick build
go build -o ipq

# Build with version info (recommended)
./build.sh        # Linux/macOS
build.cmd         # Windows
```

## Usage

```bash
ipq [target] [flags]
```

| Flag | Description |
|------|-------------|
| `-c` | Read from clipboard |
| `-d` | Show detailed info |
| `-f FILE` | Read targets from file |
| `--batch` | Batch process from stdin |
| `-o FORMAT` | Output: json, yaml, text, quiet |
| `-q` | Quiet mode (only IPs) |
| `-V` | Print version |

## Examples

```bash
ipq                    # Query your public IP
ipq 8.8.8.8 -d         # Query with details
ipq google.com         # Query domain
ipq -c                 # From clipboard
echo "8.8.8.8" | ipq   # From stdin
ipq -f ips.txt         # Batch from file
ipq 8.8.8.8 -o json    # JSON output
```

## Demo

```
$ ipq 8.8.8.8 -d

 [DONE]  Target: 8.8.8.8
 ─────────────────────────────────────────
  IPv4      : 8.8.8.8 [Public]
  IPv6      : N/A

  [ GEOLOCATION ]
  ISP       : Google LLC
  Location  : Mountain View, California, US

  [ ATTRIBUTES ]
  Mobile Net   : No
  Proxy/VPN    : No
  Data Center  : ✓ Yes

 (r to refresh, 4/6 to copy, q to quit)
```

## Interactive Keys

| Key | Action |
|-----|--------|
| `r` | Refresh |
| `d` | Toggle detail |
| `4/6` | Copy IPv4/IPv6 |
| `q` | Quit |

## Architecture

```mermaid
flowchart TB
  %% 目标：清晰展示分层架构 + 单向依赖（无循环）

  subgraph CMD[cmd]
    cmd_root[root.go]
    cmd_version[version.go]
    cmd_completion[completion.go]
  end

  subgraph CLI[internal/cli]
    cli_input[input.go]
    cli_batch[batch.go]
    cli_config[config.go]
    cli_exit[exit.go]
  end

  subgraph TUI[internal/tui]
    tui_app[app.go]
  end

  subgraph OUT[internal/output]
    out_format[format.go]
    out_style[style.go]
    out_error[error.go]
  end

  subgraph NET[internal/network]
    net_dns[dns.go]
    net_fetch[fetch.go]
    net_resolve[resolve.go]
    net_types[types.go]
  end

  subgraph IP[internal/ip]
    ip_classify[classify.go]
    ip_validate[validate.go]
  end

  %% 依赖关系（单向）
  CMD --> CLI
  CMD --> TUI
  CMD --> OUT
  CMD --> IP

  CLI --> IP
  CLI --> OUT

  TUI --> NET
  TUI --> OUT

  OUT --> IP
  OUT --> NET

  NET --> IP
```

**依赖方向 (严格单向，无循环):**

```
ip        ← 底层，0 依赖
 ↑
network   ← 依赖 ip
 ↑
output    ← 依赖 ip, network
 ↑
tui       ← 依赖 network, output
 ↑
cli       ← 依赖 ip, output
 ↑
cmd       ← 依赖 cli, tui, output, ip
```

## Project Structure

```
.
├── cmd/                    # CLI 命令
│   ├── root.go             # 主命令
│   ├── version.go          # 版本命令
│   └── completion.go       # Shell 补全
│
├── internal/
│   ├── ip/                 # IP 地址处理 (底层)
│   │   ├── classify.go     # 类型分类
│   │   └── validate.go     # 验证、URL 提取
│   │
│   ├── network/            # 网络请求
│   │   ├── types.go        # 数据结构
│   │   ├── dns.go          # DNS 解析
│   │   ├── fetch.go        # HTTP 请求
│   │   └── resolve.go      # 统一解析接口
│   │
│   ├── output/             # 输出格式化
│   │   ├── style.go        # 终端样式
│   │   ├── error.go        # 错误格式化
│   │   └── format.go       # JSON/YAML/Text
│   │
│   ├── tui/                # 交互式界面
│   │   └── app.go          # Bubble Tea 应用
│   │
│   └── cli/                # CLI 辅助
│       ├── exit.go         # 退出码
│       ├── config.go       # 配置加载
│       ├── input.go        # stdin/环境检测
│       └── batch.go        # 批量处理
│
├── main.go
├── go.mod
└── README.md
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `NO_COLOR` | Disable colors |
| `CI` | Force non-interactive mode |
| `IPQ_CONFIG` | Custom config path |

## Configuration

`~/.config/ipq/config.yaml` or `~/.ipq.yaml`:

```yaml
show_detail: false
timeout: 5s
```

## Shell Completion

```bash
ipq completion bash > /etc/bash_completion.d/ipq
ipq completion zsh > "${fpath[1]}/_ipq"
ipq completion fish > ~/.config/fish/completions/ipq.fish
ipq completion powershell >> $PROFILE
```

## License

MIT
