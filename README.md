# saferay

DNS leak protection utility for macOS with Xray/Hiddify VPN.

## What it does

When using Xray-based VPN clients (Hiddify, v2rayN, etc.), DNS requests can bypass the tunnel and leak to your ISP. This utility provides two protection modes:

1. **Xray mode** — Forces all DNS through VPN tunnel using pf firewall
2. **Light mode** — Uses Google DNS (8.8.8.8) + DNS cache flush (no VPN required)

## Requirements

- macOS 10.15+
- Go 1.21+ (for building from source)
- Administrator privileges (sudo)
- For Xray mode: Xray-based VPN client (Hiddify, v2rayN, etc.)

## Installation

### Build from source

```bash
git clone https://github.com/alexfomin/saferay.git
cd saferay
make build
```

### Install globally

```bash
# Basic install (binary only)
./saferay install

# Install with light mode (DNS flush + 8.8.8.8)
./saferay install --light
```

### Check system requirements

```bash
saferay check
```

## Modes

### Light Mode (no VPN required)

Simple DNS protection without VPN:
- Sets DNS to Google (8.8.8.8, 8.8.4.4)
- Flushes DNS cache on every reboot
- Good for basic protection from ISP DNS hijacking

```bash
# Setup light mode
saferay light setup

# Check status
saferay light status

# Remove light mode
saferay light reset
```

### Xray Mode (requires VPN)

Full DNS leak protection with VPN:
- Blocks all DNS except through VPN tunnel
- Prevents any DNS leaks to ISP
- Requires Xray/Hiddify running

**Recommended: Auto mode** (enables/disables automatically with VPN):

```bash
# One-time setup
saferay xray install
saferay xray auto start

# That's it! Protection auto-enables when VPN connects
# and auto-disables when VPN disconnects
```

**Manual mode** (if you prefer manual control):

```bash
# Install firewall rules
saferay xray install

# Start your VPN (Hiddify, etc.)

# Enable protection
saferay xray enable

# Check status
saferay xray status

# Disable when done
saferay xray disable
```

## Commands Reference

### Global

| Command | Description |
|---------|-------------|
| `saferay install` | Install saferay to `/usr/local/bin` |
| `saferay install --light` | Install + setup light mode |
| `saferay uninstall` | Remove saferay and all configurations |
| `saferay check` | Check system requirements |
| `saferay version` | Show version |
| `saferay help` | Show help message |

### Light Mode

| Command | Description |
|---------|-------------|
| `saferay light setup` | Setup light mode (DNS flush + 8.8.8.8) |
| `saferay light reset` | Remove light mode settings |
| `saferay light status` | Show light mode status |

### DNS Cache

| Command | Description |
|---------|-------------|
| `saferay dns setup` | Setup DNS cache flush on reboot |
| `saferay dns remove` | Remove DNS flush daemon |
| `saferay dns status` | Check DNS flush daemon status |
| `saferay dns flush` | Flush DNS cache now |

### Xray Mode

| Command | Description |
|---------|-------------|
| `saferay xray install` | Install pf firewall rules |
| `saferay xray enable` | Enable firewall (activate protection) |
| `saferay xray disable` | Disable firewall |
| `saferay xray reset` | Remove all Xray firewall rules |
| `saferay xray status` | Show protection status |
| `saferay xray auto start` | Start auto mode (recommended) |
| `saferay xray auto stop` | Stop auto mode |
| `saferay xray auto status` | Show auto mode status |

## Switching Modes

### Light → Xray

```bash
# Just run xray install - it will automatically:
# - Reset light mode DNS settings
# - Keep DNS flush daemon (useful for both)
saferay xray install
saferay xray enable
```

### Xray → Light

```bash
saferay xray reset
saferay light setup
```

## Typical Workflows

### Light mode (no VPN)

```bash
# One-time setup
./saferay install --light

# That's it! DNS is now set to 8.8.8.8 and will flush on reboot
```

### Xray mode (with VPN)

```bash
# One-time setup
./saferay install
saferay xray install

# Daily usage
# 1. Start VPN (Hiddify)
# 2. Enable protection
saferay xray enable

# 3. Work...

# 4. Disable protection
saferay xray disable
# 5. Stop VPN
```

## Troubleshooting

### "Resource busy" error

```bash
sudo pfctl -F all
saferay xray enable
```

### DNS not working with Xray mode

1. Make sure VPN is running
2. Check tunnel interface:
   ```bash
   ifconfig | grep "^utun"
   ```
3. If not `utun4`, edit `/etc/pf.anchors/xray-dns`

### View firewall rules

```bash
sudo pfctl -s rules
sudo pfctl -a "xray-dns" -s rules
```

### Complete reset

```bash
saferay uninstall
```

## Files

| Path | Description |
|------|-------------|
| `/usr/local/bin/saferay` | Main binary |
| `/etc/pf.conf` | macOS packet filter config |
| `/etc/pf.anchors/xray-dns` | Xray DNS protection rules |
| `/etc/saferay/light.conf` | Light mode config |
| `/Library/LaunchDaemons/com.saferay.dnsflush.plist` | DNS flush daemon |
| `/Library/LaunchDaemons/com.saferay.xray-auto.plist` | Auto mode daemon |
| `/var/log/saferay-xray.log` | Auto mode log |

## How Xray Mode Works

Uses macOS `pf` (Packet Filter) firewall:

```
pass out quick on utun4 proto { udp tcp } to any port 53
```
Allow DNS through VPN tunnel

```
pass out quick on lo0 proto { udp tcp } to 127.0.0.0/8 port 53
```
Allow DNS to localhost

```
block out quick proto { udp tcp } to any port 53
```
Block all other DNS (prevents leaks)

## Development

```bash
# Setup pre-commit hooks
pre-commit install

# Format code
make fmt

# Run linter
make lint

# Build (runs fmt + lint first)
make build

# Build without checks (fast)
make build-fast

# Test locally with goreleaser
make snapshot

# Release (requires git tag)
git tag v1.0.0
git push origin --tags
# GitHub Actions will build and release
```

## Security Notes

- All commands require `sudo` for system modifications
- Firewall rules only affect DNS traffic (port 53)
- Other traffic is not affected
- When protection is disabled, DNS works normally
