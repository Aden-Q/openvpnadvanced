# OpenVPNAdvanced (English Documentation)

> A rule-based OpenVPN traffic splitter supporting DoH DNS proxy, rule subscriptions, dynamic route injection, DNS caching, and more.

---
[中文文档](https://github.com/iaaaannn0/openvpnadvanced/blob/main/README_CN.md)

## 📚 Table of Contents

- [Project Overview](#project-overview)
- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Build & Installation](#build--installation)
  - [Start the Service](#start-the-service)
  - [Configure Local DNS](#configure-local-dns)
- [Configuration Guide](#configuration-guide)
- [How It Works](#how-it-works)
- [Architecture](#architecture)
- [Module Description](#module-description)
- [FAQ](#faq)
- [Performance Optimization](#performance-optimization)
- [Security & Privacy](#security--privacy)
- [How to Verify VPN Routing](#how-to-verify-vpn-routing)
- [Developer Guide](#developer-guide)
- [License](#license)

---

## Project Overview

This project is designed to provide OpenVPN users with a high-performance and flexible rule-based traffic splitter. It prevents all traffic from going through VPN and supports subscriptions, DNS caching, CNAME resolution, and DNS pollution protection.

### Key Benefits
- **Smart Traffic Routing**: Automatically routes traffic based on rules
- **Enhanced Privacy**: Supports DoH (DNS over HTTPS) for secure DNS queries
- **Improved Performance**: DNS caching and optimized routing
- **Easy Management**: Simple configuration and rule management
- **Real-time Monitoring**: Comprehensive logging and status tracking

---

## Features

### Core Features
- ✅ Local DNS proxy (supports DoH / TCP / UDP)
- ✅ Custom rules and remote subscriptions (auto deduplication & merge)
- ✅ Accurate routing (adds static route via utunX)
- ✅ Automatic VPN interface detection (e.g. utun0 / utun8)
- ✅ Fixes default macOS gateway to direct network interface
- ✅ Supports recursive CNAME resolution
- ✅ Ultra-fast response via cache
- ✅ One-command startup, no complex setup

### Advanced Features
- 🔍 Domain Tracing Tool (`trace.go`)
  - Detailed network information display
  - Routing path analysis
  - Automatic route fixing
  - CNAME chain visualization
- 📊 Interactive Console (`ovpnctl`)
  - Real-time log viewing
  - Route testing
  - Interface management
  - Configuration reloading

---

## Getting Started

### Prerequisites

- Go 1.18+
- macOS (supports `route`, `scutil`, etc.)
- Connected OpenVPN client (e.g. Tunnelblick)

### Build & Installation

```bash
# Clone the repository
git clone https://github.com/iaaaannn0/openvpnadvanced.git
cd openvpnadvanced

# Build the project
go build -o openvpnadvanced ./cmd
```

### Start the Service

```bash
# Start the service
sudo ./openvpnadvanced
```

### Interactive Console

The tool provides an interactive command console (ovpnctl) for runtime control.

#### Start the Console

```bash
sudo ./openvpnadvanced --start
```

#### Available Commands

| Command | Description | Example |
|---------|-------------|---------|
| `start` | Start core logic in background | `start` |
| `startv` | Start with real-time logs | `startv` |
| `status` | Check service status | `status` |
| `view-log` | View logs with filters | `view-log info` |
| `test` | Test domain rule match | `test example.com` |
| `rtest` | Test domain resolution | `rtest example.com` |
| `show-iface` | Show interface info | `show-iface` |
| `reload-config` | Reload configuration | `reload-config` |
| `clear` | Clear console | `clear` |

### Domain Tracing Tool

The `trace.go` tool provides detailed information about domain resolution and routing:

```bash
# Run the tracing tool
go run tools/trace.go example.com
```

#### Output Information
- Network Information
  - Domain resolution
  - IP address
  - Matched rules
  - CNAME chain
- Routing Information
  - Current interface
  - VPN interface
  - Default gateway
  - Route status

---

## Configuration Guide

### DNS Configuration
1. Set local DNS to 127.0.0.1
2. Configure DNS proxy settings in `config.ini`
3. Add custom rules or subscribe to rule lists

### Rule Management
- Local rules: `assets/rule.list`
- Remote subscriptions: Add URLs in `config.ini`
- Automatic updates: Configure in `config.ini`

---

## How It Works

1. **DNS Resolution**
   - Local DNS proxy handles queries
   - Supports DoH for secure queries
   - Caches responses for performance

2. **Traffic Routing**
   - Analyzes domain rules
   - Routes traffic through VPN or direct
   - Maintains optimal routing paths

3. **Interface Management**
   - Detects VPN interfaces
   - Manages network routes
   - Handles interface changes

---

## Architecture

```
├── cmd/                 # Command-line interface
├── dnsmasq/            # DNS proxy implementation
├── vpn/                # VPN routing management
├── tools/              # Utility tools
│   └── trace.go        # Domain tracing tool
├── assets/             # Configuration and rules
└── config.ini          # Main configuration file
```

---

## Module Description

### DNS Proxy (`dnsmasq/`)
- Handles DNS queries
- Implements caching
- Supports DoH
- Manages rules

### VPN Routing (`vpn/`)
- Manages network interfaces
- Handles route injection
- Detects VPN status
- Fixes routing issues

### Tools (`tools/`)
- Domain tracing
- Route testing
- Interface inspection
- Log management

---

## FAQ

### Common Issues
1. **DNS not working**
   - Check local DNS settings
   - Verify DNS proxy is running
   - Check rule configuration

2. **VPN routing issues**
   - Verify VPN connection
   - Check interface detection
   - Review route rules

3. **Performance problems**
   - Clear DNS cache
   - Optimize rules
   - Check network conditions

---

## Performance Optimization

### DNS Optimization
- Implement caching
- Optimize rule matching
- Use efficient algorithms

### Routing Optimization
- Minimize route changes
- Optimize interface detection
- Cache route decisions

---

## Security & Privacy

### DNS Security
- Support for DoH
- DNS cache protection
- Rule validation

### Routing Security
- Secure route injection
- Interface validation
- Access control

---

## How to Verify VPN Routing

1. Use the tracing tool:
```bash
go run tools/trace.go example.com
```

2. Check routing information:
```bash
sudo ./openvpnadvanced --start
ovpnctl> rtest example.com
```

---

## Developer Guide

### Building
```bash
go build -o openvpnadvanced ./cmd
```

### Testing
```bash
go test ./...
```

### Contributing
1. Fork the repository
2. Create a feature branch
3. Submit a pull request

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
