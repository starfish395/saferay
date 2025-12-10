package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	daemonLabel = "com.saferay.dnsflush"
	daemonPath  = "/Library/LaunchDaemons/com.saferay.dnsflush.plist"
	daemonPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.saferay.dnsflush</string>
    <key>ProgramArguments</key>
    <array>
        <string>/bin/bash</string>
        <string>-c</string>
        <string>dscacheutil -flushcache; killall -HUP mDNSResponder</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>`
)

func cmdDNS(action string) {
	switch action {
	case "setup":
		setupDNSDaemon()
	case "remove":
		removeDNSDaemon()
	case "status":
		statusDNSDaemon()
	case "flush":
		flushDNS()
	default:
		fmt.Printf("Unknown dns action: %s\n", action)
		os.Exit(1)
	}
}

func setupDNSDaemon() {
	// Write plist to temp and move with sudo
	tmpPath := "/tmp/saferay_dns.plist"
	if err := os.WriteFile(tmpPath, []byte(daemonPlist), 0644); err != nil {
		fmt.Printf("Error writing plist: %v\n", err)
		os.Exit(1)
	}

	cmds := [][]string{
		{"sudo", "mv", tmpPath, daemonPath},
		{"sudo", "chown", "root:wheel", daemonPath},
		{"sudo", "chmod", "644", daemonPath},
		{"sudo", "launchctl", "load", "-w", daemonPath},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error running %v: %v\n", args, err)
			os.Exit(1)
		}
	}

	fmt.Println("✓ DNS flush daemon installed (will flush DNS on every reboot)")
}

func removeDNSDaemon() {
	exec.Command("sudo", "launchctl", "unload", "-w", daemonPath).Run()
	exec.Command("sudo", "rm", "-f", daemonPath).Run()
	fmt.Println("✓ DNS flush daemon removed")
}

func statusDNSDaemon() {
	if _, err := os.Stat(daemonPath); os.IsNotExist(err) {
		fmt.Println("DNS flush daemon: not installed")
		return
	}

	out, _ := exec.Command("sudo", "launchctl", "list", daemonLabel).CombinedOutput()
	if strings.Contains(string(out), daemonLabel) {
		fmt.Println("DNS flush daemon: ✓ installed and loaded")
	} else {
		fmt.Println("DNS flush daemon: installed but not loaded")
	}
}

func flushDNS() {
	cmd := exec.Command("sudo", "dscacheutil", "-flushcache")
	cmd.Stdin = os.Stdin
	cmd.Run()

	cmd = exec.Command("sudo", "killall", "-HUP", "mDNSResponder")
	cmd.Stdin = os.Stdin
	cmd.Run()

	fmt.Println("✓ DNS cache flushed")
}
