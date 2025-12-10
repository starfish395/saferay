package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func Execute() {
	// Check macOS
	if runtime.GOOS != "darwin" {
		fmt.Println("Error: saferay only works on macOS")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		cmdInstall()
	case "uninstall":
		cmdUninstall()
	case "dns":
		if len(os.Args) < 3 {
			fmt.Println("Usage: saferay dns [setup|remove|status|flush]")
			os.Exit(1)
		}
		cmdDNS(os.Args[2])
	case "xray":
		if len(os.Args) < 3 {
			fmt.Println("Usage: saferay xray [install|enable|disable|reset|status]")
			os.Exit(1)
		}
		cmdXray(os.Args[2])
	case "check":
		cmdCheck()
	case "help", "-h", "--help":
		printUsage()
	case "version", "-v", "--version":
		cmdVersion()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func cmdCheck() {
	fmt.Println("=== System Check ===\n")
	allOk := true

	// Check macOS
	fmt.Printf("macOS:           ")
	if runtime.GOOS == "darwin" {
		fmt.Println("✓ Yes")
	} else {
		fmt.Println("✗ No (required)")
		allOk = false
	}

	// Check pfctl
	fmt.Printf("pfctl:           ")
	if _, err := exec.LookPath("pfctl"); err == nil {
		fmt.Println("✓ Available")
	} else {
		fmt.Println("✗ Not found")
		allOk = false
	}

	// Check launchctl
	fmt.Printf("launchctl:       ")
	if _, err := exec.LookPath("launchctl"); err == nil {
		fmt.Println("✓ Available")
	} else {
		fmt.Println("✗ Not found")
		allOk = false
	}

	// Check sudo access
	fmt.Printf("sudo:            ")
	if _, err := exec.LookPath("sudo"); err == nil {
		fmt.Println("✓ Available")
	} else {
		fmt.Println("✗ Not found")
		allOk = false
	}

	// Check for VPN tunnel interface
	fmt.Printf("VPN tunnel:      ")
	out, _ := exec.Command("ifconfig").CombinedOutput()
	if strings.Contains(string(out), "utun") {
		// Count utun interfaces
		lines := strings.Split(string(out), "\n")
		count := 0
		for _, line := range lines {
			if strings.HasPrefix(line, "utun") {
				count++
			}
		}
		fmt.Printf("✓ Found %d utun interface(s)\n", count)
	} else {
		fmt.Println("⚠ No utun interfaces (start VPN first)")
	}

	// Check pf.conf exists
	fmt.Printf("pf.conf:         ")
	if _, err := os.Stat("/etc/pf.conf"); err == nil {
		fmt.Println("✓ Exists")
	} else {
		fmt.Println("✗ Not found")
		allOk = false
	}

	fmt.Println()
	if allOk {
		fmt.Println("All checks passed. Ready to use.")
	} else {
		fmt.Println("Some checks failed. saferay may not work correctly.")
	}
}

func printUsage() {
	fmt.Println(`saferay - DNS leak protection for macOS with Xray/Hiddify

Usage:
  saferay install              Install saferay to /usr/local/bin
  saferay uninstall            Remove saferay from system
  saferay check                Check system requirements
  saferay version              Show version

  saferay dns setup            Setup DNS cache flush on reboot
  saferay dns remove           Remove DNS flush daemon
  saferay dns status           Check DNS flush daemon status
  saferay dns flush            Flush DNS cache now

  saferay xray install         Install pf rules for Xray DNS protection
  saferay xray enable          Enable pf firewall with Xray rules
  saferay xray disable         Disable pf firewall
  saferay xray reset           Remove all Xray pf rules
  saferay xray status          Show current pf/Xray status`)
}
