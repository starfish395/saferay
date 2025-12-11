package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	defaultDNS      = "8.8.8.8"
	defaultDNS2     = "8.8.4.4"
	lightConfigPath = "/etc/saferay/light.conf"
)

func cmdLight(action string) {
	switch action {
	case "setup":
		setupLightMode()
	case "reset":
		resetLightMode()
	case "status":
		statusLightMode()
	default:
		fmt.Printf("Unknown light action: %s\n", action)
		fmt.Println("Usage: saferay light [setup|reset|status]")
		os.Exit(1)
	}
}

func setupLightMode() {
	fmt.Println("Setting up light mode...")

	// 1. Setup DNS flush daemon
	setupDNSDaemon()

	// 2. Get active network service
	service := getActiveNetworkService()
	if service == "" {
		fmt.Println("Warning: Could not detect active network service")
		fmt.Println("Please set DNS manually: networksetup -setdnsservers \"Wi-Fi\" 8.8.8.8 8.8.4.4")
		return
	}

	// 3. Save original DNS for reset
	saveOriginalDNS(service)

	// 4. Set DNS to 8.8.8.8
	setDNS(service, defaultDNS, defaultDNS2)

	fmt.Println()
	fmt.Println("✓ Light mode enabled:")
	fmt.Println("  - DNS cache will flush on every reboot")
	fmt.Printf("  - DNS set to %s, %s on %s\n", defaultDNS, defaultDNS2, service)
}

func resetLightMode() {
	fmt.Println("Resetting light mode...")

	// 1. Remove DNS flush daemon
	removeDNSDaemon()

	// 2. Reset DNS to automatic
	service := getActiveNetworkService()
	if service != "" {
		resetDNS(service)
	}

	// 3. Remove config
	_ = exec.Command("sudo", "rm", "-rf", "/etc/saferay").Run()

	fmt.Println("✓ Light mode disabled")
}

func statusLightMode() {
	fmt.Println("=== Light Mode Status ===")
	fmt.Println()

	// Check DNS flush daemon
	statusDNSDaemon()

	// Check current DNS
	service := getActiveNetworkService()
	if service != "" {
		fmt.Printf("\nNetwork service: %s\n", service)
		out, _ := exec.Command("networksetup", "-getdnsservers", service).CombinedOutput()
		outStr := strings.TrimSpace(string(out))
		if strings.Contains(outStr, "There aren't any DNS Servers") {
			fmt.Println("DNS servers:     automatic (DHCP)")
		} else {
			fmt.Printf("DNS servers:     %s\n", strings.ReplaceAll(outStr, "\n", ", "))
		}
	}
}

func getActiveNetworkService() string {
	// Try to find active network service
	// Priority: Wi-Fi > Ethernet > any other

	out, err := exec.Command("networksetup", "-listallnetworkservices").CombinedOutput()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(out), "\n")
	var services []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip header and disabled services (marked with *)
		if line == "" || strings.HasPrefix(line, "An asterisk") || strings.HasPrefix(line, "*") {
			continue
		}
		services = append(services, line)
	}

	// Check which service is active (has an IP)
	priorities := []string{"Wi-Fi", "Ethernet", "USB 10/100/1000 LAN", "Thunderbolt Ethernet"}

	for _, priority := range priorities {
		for _, svc := range services {
			if svc == priority {
				// Check if it has an IP
				out, _ := exec.Command("networksetup", "-getinfo", svc).CombinedOutput()
				if strings.Contains(string(out), "IP address:") && !strings.Contains(string(out), "IP address: none") {
					return svc
				}
			}
		}
	}

	// Fallback: return first service with IP
	for _, svc := range services {
		out, _ := exec.Command("networksetup", "-getinfo", svc).CombinedOutput()
		if strings.Contains(string(out), "IP address:") && !strings.Contains(string(out), "IP address: none") {
			return svc
		}
	}

	// Last fallback: just return Wi-Fi
	for _, svc := range services {
		if svc == "Wi-Fi" {
			return svc
		}
	}

	return ""
}

func saveOriginalDNS(service string) {
	// Create config directory
	_ = exec.Command("sudo", "mkdir", "-p", "/etc/saferay").Run()

	// Get current DNS
	out, _ := exec.Command("networksetup", "-getdnsservers", service).CombinedOutput()
	outStr := strings.TrimSpace(string(out))

	var content string
	if strings.Contains(outStr, "There aren't any DNS Servers") {
		content = fmt.Sprintf("service=%s\ndns=auto\n", service)
	} else {
		dns := strings.ReplaceAll(outStr, "\n", " ")
		content = fmt.Sprintf("service=%s\ndns=%s\n", service, dns)
	}

	// Write to temp and move
	tmpFile := "/tmp/saferay_light.conf"
	_ = os.WriteFile(tmpFile, []byte(content), 0644)
	_ = exec.Command("sudo", "mv", tmpFile, lightConfigPath).Run()
}

func setDNS(service string, dns ...string) {
	args := append([]string{"networksetup", "-setdnsservers", service}, dns...)
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error setting DNS: %v\n", err)
		return
	}

	fmt.Printf("✓ DNS set to %s on %s\n", strings.Join(dns, ", "), service)
}

func resetDNS(service string) {
	// Try to read original config
	content, err := os.ReadFile(lightConfigPath)
	if err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "dns=") {
				dns := strings.TrimPrefix(line, "dns=")
				if dns == "auto" {
					// Set to empty (automatic)
					cmd := exec.Command("sudo", "networksetup", "-setdnsservers", service, "Empty")
					cmd.Stdin = os.Stdin
					_ = cmd.Run()
					fmt.Printf("✓ DNS reset to automatic on %s\n", service)
					return
				}
				// Restore original DNS
				dnsServers := strings.Fields(dns)
				args := append([]string{"networksetup", "-setdnsservers", service}, dnsServers...)
				cmd := exec.Command("sudo", args...)
				cmd.Stdin = os.Stdin
				_ = cmd.Run()
				fmt.Printf("✓ DNS restored to %s on %s\n", dns, service)
				return
			}
		}
	}

	// Fallback: set to empty (automatic)
	cmd := exec.Command("sudo", "networksetup", "-setdnsservers", service, "Empty")
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
	fmt.Printf("✓ DNS reset to automatic on %s\n", service)
}

// SetupLightMode is exported for use in install.go
func SetupLightMode() {
	setupLightMode()
}
