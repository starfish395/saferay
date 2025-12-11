package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	autoDaemonLabel = "com.saferay.xray-auto"
	autoDaemonPath  = "/Library/LaunchDaemons/com.saferay.xray-auto.plist"
	autoDaemonPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.saferay.xray-auto</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/saferay</string>
        <string>xray</string>
        <string>watch</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/var/log/saferay-xray.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/saferay-xray.log</string>
</dict>
</plist>`
	watchInterval = 5 * time.Second
)

// cmdXrayAuto handles the auto subcommand
func cmdXrayAuto(action string) {
	switch action {
	case "start":
		startAutoDaemon()
	case "stop":
		stopAutoDaemon()
	case "status":
		statusAutoDaemon()
	default:
		fmt.Printf("Unknown auto action: %s\n", action)
		fmt.Println("Usage: saferay xray auto [start|stop|status]")
		os.Exit(1)
	}
}

// cmdXrayWatch runs the VPN monitoring loop (called by daemon)
func cmdXrayWatch() {
	fmt.Println("saferay: Starting VPN watch daemon...")

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	vpnConnected := false
	pfEnabled := isPfEnabled()

	// Initial state: if VPN is connected, enable pf
	if isVPNConnected() {
		if !pfEnabled {
			fmt.Println("saferay: VPN detected at startup, enabling pf...")
			enablePfQuiet()
		}
		vpnConnected = true
	} else {
		// VPN not connected - make sure pf is disabled
		if pfEnabled {
			fmt.Println("saferay: No VPN at startup, disabling pf...")
			disablePfQuiet()
		}
	}

	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("saferay: Shutting down watch daemon...")
			return
		case <-ticker.C:
			currentlyConnected := isVPNConnected()

			if currentlyConnected && !vpnConnected {
				// VPN just connected
				fmt.Println("saferay: VPN connected, enabling DNS protection...")
				enablePfQuiet()
				vpnConnected = true
			} else if !currentlyConnected && vpnConnected {
				// VPN just disconnected
				fmt.Println("saferay: VPN disconnected, disabling DNS protection...")
				disablePfQuiet()
				vpnConnected = false
			}
		}
	}
}

// isVPNConnected checks if a VPN tunnel interface exists
func isVPNConnected() bool {
	// Check scutil --dns for utun interface with DNS
	out, err := exec.Command("scutil", "--dns").CombinedOutput()
	if err != nil {
		return false
	}

	// Look for utun interface in DNS config
	// This indicates VPN is routing DNS
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "if_index") && strings.Contains(line, "utun") {
			return true
		}
	}

	// Fallback: check if utun4+ exists with IP
	out, err = exec.Command("ifconfig").CombinedOutput()
	if err != nil {
		return false
	}

	// Parse ifconfig output for utun interfaces
	inUtun := false
	utunName := ""
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "utun") {
			parts := strings.Split(line, ":")
			if len(parts) > 0 {
				utunName = parts[0]
				// Check if it's utun4 or higher (skip system utuns 0-3)
				if len(utunName) > 4 {
					numPart := utunName[4:]
					if numPart >= "4" {
						inUtun = true
					}
				}
			}
		} else if inUtun && strings.Contains(line, "inet ") && !strings.Contains(line, "inet6") {
			// Found IPv4 address on utun4+
			return true
		} else if !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, " ") {
			inUtun = false
		}
	}

	return false
}

// isPfEnabled checks if pf firewall is currently enabled
func isPfEnabled() bool {
	out, _ := exec.Command("pfctl", "-s", "info").CombinedOutput()
	return strings.Contains(string(out), "Status: Enabled")
}

// enablePfQuiet enables pf without printing to stdout
func enablePfQuiet() {
	cmd := exec.Command("pfctl", "-ef", pfConf)
	_ = cmd.Run()
}

// disablePfQuiet disables pf without printing to stdout
func disablePfQuiet() {
	cmd := exec.Command("pfctl", "-d")
	_ = cmd.Run()
}

func startAutoDaemon() {
	// Check if saferay is installed
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		fmt.Println("Error: saferay not installed. Run 'saferay install' first.")
		os.Exit(1)
	}

	// Check if xray rules are installed
	if _, err := os.Stat(anchorPath); os.IsNotExist(err) {
		fmt.Println("Error: Xray rules not installed. Run 'saferay xray install' first.")
		os.Exit(1)
	}

	// Write daemon plist
	tmpPath := "/tmp/saferay_xray_auto.plist"
	if err := os.WriteFile(tmpPath, []byte(autoDaemonPlist), 0644); err != nil {
		fmt.Printf("Error writing daemon plist: %v\n", err)
		os.Exit(1)
	}

	// Stop existing daemon if running
	_ = exec.Command("sudo", "launchctl", "unload", "-w", autoDaemonPath).Run()

	cmds := [][]string{
		{"sudo", "mv", tmpPath, autoDaemonPath},
		{"sudo", "chown", "root:wheel", autoDaemonPath},
		{"sudo", "chmod", "644", autoDaemonPath},
		{"sudo", "launchctl", "load", "-w", autoDaemonPath},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error running %v: %v\n", args, err)
			os.Exit(1)
		}
	}

	fmt.Println("✓ Auto mode enabled")
	fmt.Println("  - DNS protection will auto-enable when VPN connects")
	fmt.Println("  - DNS protection will auto-disable when VPN disconnects")
	fmt.Println("  - Log: /var/log/saferay-xray.log")
}

func stopAutoDaemon() {
	_ = exec.Command("sudo", "launchctl", "unload", "-w", autoDaemonPath).Run()
	_ = exec.Command("sudo", "rm", "-f", autoDaemonPath).Run()

	// Also disable pf if it was enabled by daemon
	_ = exec.Command("pfctl", "-d").Run()

	fmt.Println("✓ Auto mode disabled")
}

func statusAutoDaemon() {
	fmt.Println("=== Xray Auto Mode Status ===")
	fmt.Println()

	// Check daemon installed
	if _, err := os.Stat(autoDaemonPath); os.IsNotExist(err) {
		fmt.Println("Auto daemon:     ✗ Not installed")
	} else {
		// Check if running
		out, _ := exec.Command("sudo", "launchctl", "list", autoDaemonLabel).CombinedOutput()
		if strings.Contains(string(out), autoDaemonLabel) {
			fmt.Println("Auto daemon:     ✓ Running")
		} else {
			fmt.Println("Auto daemon:     ⚠ Installed but not running")
		}
	}

	// Check VPN status
	if isVPNConnected() {
		fmt.Println("VPN connected:   ✓ Yes")
	} else {
		fmt.Println("VPN connected:   ✗ No")
	}

	// Check pf status
	if isPfEnabled() {
		fmt.Println("pf firewall:     ✓ Enabled")
	} else {
		fmt.Println("pf firewall:     ✗ Disabled")
	}

	// Show log tail if exists
	if _, err := os.Stat("/var/log/saferay-xray.log"); err == nil {
		fmt.Println("\nRecent log:")
		out, _ := exec.Command("tail", "-5", "/var/log/saferay-xray.log").CombinedOutput()
		if len(out) > 0 {
			fmt.Println(string(out))
		}
	}
}
