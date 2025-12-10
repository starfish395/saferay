# saferay

DNS leak protection utility for macOS with Xray/Hiddify VPN.

## What it does

When using Xray-based VPN clients (Hiddify, v2rayN, etc.), DNS requests can bypass the tunnel and leak to your ISP. This utility:

1. **Blocks direct DNS requests** — all DNS traffic is forced through the VPN tunnel
2. **Flushes DNS cache on reboot** — prevents stale DNS entries from leaking

## How it works

Uses macOS `pf` (Packet Filter) firewall to:
- Allow DNS (port 53) only through `utun4` (Xray tunnel interface)
- Allow DNS to localhost (127.0.0.1) for local resolvers
- Block all other DNS traffic

## Requirements

- macOS 10.15+
- Go 1.21+ (for building)
- Xray-based VPN client (Hiddify, v2rayN, etc.)
- Administrator privileges (sudo)

## Installation

### Build from source

```bash
git clone https://github.com/yourusername/saferay.git
cd saferay
go build -o saferay .
```

### Install globally

```bash
./saferay install
```

This copies the binary to `/usr/local/bin/saferay`.

## Usage

### Quick Start

```bash
# 1. Install DNS protection rules
saferay xray install

# 2. Start your VPN (Hiddify, etc.)

# 3. Enable protection
saferay xray enable

# 4. Check status
saferay xray status
```

### Commands

#### Global

| Command | Description |
|---------|-------------|
| `saferay install` | Install saferay to `/usr/local/bin` |
| `saferay uninstall` | Remove saferay and all its configurations |
| `saferay help` | Show help message |

#### DNS Cache Management

| Command | Description |
|---------|-------------|
| `saferay dns setup` | Setup automatic DNS cache flush on system reboot |
| `saferay dns remove` | Remove the DNS flush daemon |
| `saferay dns status` | Check if DNS flush daemon is installed |
| `saferay dns flush` | Flush DNS cache immediately |

#### Xray DNS Protection

| Command | Description |
|---------|-------------|
| `saferay xray install` | Install pf firewall rules for DNS protection |
| `saferay xray enable` | Enable the firewall (activate protection) |
| `saferay xray disable` | Disable the firewall (deactivate protection) |
| `saferay xray reset` | Remove all Xray-related firewall rules |
| `saferay xray status` | Show current protection status |

## Typical Workflow

### First-time setup

```bash
# Install saferay globally
./saferay install

# Setup DNS flush on reboot (optional but recommended)
saferay dns setup

# Install firewall rules
saferay xray install
```

### Daily usage

```bash
# 1. Start your VPN client (Hiddify, etc.)

# 2. Enable DNS protection
saferay xray enable

# 3. Work...

# 4. When done, disable protection
saferay xray disable

# 5. Stop your VPN client
```

### Verify protection is working

```bash
# Check status
saferay xray status

# Expected output:
# === Xray DNS Protection Status ===
#
# Rules installed: ✓ Yes
# pf firewall:     ✓ Enabled
# Anchor loaded:   ✓ Yes
#
# Active rules:
#   pass out quick on utun4 proto udp from any to any port = 53
#   pass out quick on utun4 proto tcp from any to any port = 53
#   pass out quick on lo0 inet proto udp from any to 127.0.0.0/8 port = 53
#   pass out quick on lo0 inet proto tcp from any to 127.0.0.0/8 port = 53
#   block drop out quick proto udp from any to any port = 53
#   block drop out quick proto tcp from any to any port = 53
```

### Test DNS leak protection

```bash
# With VPN running and protection enabled:

# This should work (goes through tunnel)
nslookup google.com

# This should also work (Xray intercepts it)
nslookup google.com 8.8.8.8

# Disable VPN, keep protection enabled:
# Both commands above should timeout (blocked)
```

## Troubleshooting

### "Resource busy" error when enabling

```bash
# Flush all rules and retry
sudo pfctl -F all
saferay xray enable
```

### DNS not working after enabling

1. Make sure your VPN is running
2. Check the tunnel interface name:
   ```bash
   ifconfig | grep "^utun"
   ```
3. If your tunnel is not `utun4`, you need to modify the rules (see Configuration section)

### Check what's in pf.conf

```bash
cat /etc/pf.conf
```

Should contain at the end:
```
anchor "xray-dns"
load anchor "xray-dns" from "/etc/pf.anchors/xray-dns"
```

### View active firewall rules

```bash
sudo pfctl -s rules
sudo pfctl -a "xray-dns" -s rules
```

### Complete reset

```bash
# Remove everything and start fresh
saferay xray reset
saferay dns remove

# Or uninstall completely
saferay uninstall
```

## Configuration

### Changing tunnel interface

By default, saferay uses `utun4` as the tunnel interface. If your VPN uses a different interface:

1. Find your tunnel interface:
   ```bash
   ifconfig | grep "^utun"
   ```

2. Edit `/etc/pf.anchors/xray-dns` and replace `utun4` with your interface name

3. Reload rules:
   ```bash
   saferay xray enable
   ```

### Firewall rules explained

```
pass out quick on utun4 proto { udp tcp } to any port 53
```
Allow DNS through VPN tunnel (utun4)

```
pass out quick on lo0 proto { udp tcp } to 127.0.0.0/8 port 53
```
Allow DNS to localhost (for local DNS resolvers)

```
block out quick proto { udp tcp } to any port 53
```
Block all other DNS traffic (prevents leaks)

## Files

| Path | Description |
|------|-------------|
| `/usr/local/bin/saferay` | Main binary |
| `/etc/pf.conf` | macOS packet filter config (modified) |
| `/etc/pf.anchors/xray-dns` | DNS protection rules |
| `/Library/LaunchDaemons/com.saferay.dnsflush.plist` | DNS flush daemon |

## Uninstallation

```bash
# Remove everything
saferay uninstall
```

This will:
- Remove the binary from `/usr/local/bin`
- Remove DNS flush daemon
- Remove all firewall rules
- Restore original pf.conf

## Security Notes

- All commands require `sudo` for system modifications
- The firewall rules only affect DNS traffic (port 53)
- Other traffic is not affected
- When protection is disabled, DNS works normally

