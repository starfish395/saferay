package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	pfConf      = "/etc/pf.conf"
	anchorPath  = "/etc/pf.anchors/xray-dns"
	anchorName  = "xray-dns"
	anchorRules = `pass out quick on utun4 proto { udp tcp } to any port 53
pass out quick on lo0 proto { udp tcp } to 127.0.0.0/8 port 53
block out quick proto { udp tcp } to any port 53
`
)

func cmdXray(action string) {
	switch action {
	case "install":
		installXrayRules()
	case "enable":
		enableXray()
	case "disable":
		disableXray()
	case "reset":
		resetXrayRules()
	case "status":
		statusXray()
	default:
		fmt.Printf("Unknown xray action: %s\n", action)
		os.Exit(1)
	}
}

func installXrayRules() {
	// Write anchor file
	tmpAnchor := "/tmp/xray-dns-anchor"
	if err := os.WriteFile(tmpAnchor, []byte(anchorRules), 0644); err != nil {
		fmt.Printf("Error writing anchor: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("sudo", "mv", tmpAnchor, anchorPath)
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error moving anchor: %v\n", err)
		os.Exit(1)
	}

	// Read current pf.conf
	pfContent, err := os.ReadFile(pfConf)
	if err != nil {
		fmt.Printf("Error reading pf.conf: %v\n", err)
		os.Exit(1)
	}

	// Check if anchor already exists
	if strings.Contains(string(pfContent), anchorName) {
		// Remove existing xray-dns lines first
		lines := strings.Split(string(pfContent), "\n")
		var filtered []string
		for _, line := range lines {
			if !strings.Contains(line, "xray-dns") {
				filtered = append(filtered, line)
			}
		}
		pfContent = []byte(strings.Join(filtered, "\n"))
	}

	// Add anchor lines
	newContent := string(pfContent)
	// Remove trailing whitespace
	newContent = strings.TrimRight(newContent, "\n\t ")
	newContent += fmt.Sprintf("\nanchor \"%s\"\nload anchor \"%s\" from \"%s\"\n", anchorName, anchorName, anchorPath)

	tmpPf := "/tmp/pf.conf.new"
	if err := os.WriteFile(tmpPf, []byte(newContent), 0644); err != nil {
		fmt.Printf("Error writing pf.conf: %v\n", err)
		os.Exit(1)
	}

	cmd = exec.Command("sudo", "mv", tmpPf, pfConf)
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error updating pf.conf: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Xray DNS protection rules installed")
	fmt.Println("  Run 'saferay xray enable' to activate")
}

func enableXray() {
	// Check if rules installed
	if _, err := os.Stat(anchorPath); os.IsNotExist(err) {
		fmt.Println("Xray rules not installed. Run 'saferay xray install' first")
		os.Exit(1)
	}

	cmd := exec.Command("sudo", "pfctl", "-ef", pfConf)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error enabling pf: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Xray DNS protection enabled")
}

func disableXray() {
	cmd := exec.Command("sudo", "pfctl", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()

	fmt.Println("✓ pf firewall disabled")
}

func resetXrayRules() {
	// Disable pf first
	exec.Command("sudo", "pfctl", "-d").Run()

	// Read and clean pf.conf
	pfContent, err := os.ReadFile(pfConf)
	if err == nil {
		lines := strings.Split(string(pfContent), "\n")
		var filtered []string
		for _, line := range lines {
			if !strings.Contains(line, "xray-dns") {
				filtered = append(filtered, line)
			}
		}
		newContent := strings.Join(filtered, "\n")

		tmpPf := "/tmp/pf.conf.clean"
		if os.WriteFile(tmpPf, []byte(newContent), 0644) == nil {
			exec.Command("sudo", "mv", tmpPf, pfConf).Run()
		}
	}

	// Remove anchor file
	exec.Command("sudo", "rm", "-f", anchorPath).Run()

	fmt.Println("✓ Xray DNS rules removed")
}

func statusXray() {
	fmt.Println("=== Xray DNS Protection Status ===")
	fmt.Println()

	// Check anchor file
	if _, err := os.Stat(anchorPath); os.IsNotExist(err) {
		fmt.Println("Rules installed: ✗ No")
	} else {
		fmt.Println("Rules installed: ✓ Yes")
	}

	// Check pf status
	out, _ := exec.Command("sudo", "pfctl", "-s", "info").CombinedOutput()
	if strings.Contains(string(out), "Status: Enabled") {
		fmt.Println("pf firewall:     ✓ Enabled")
	} else {
		fmt.Println("pf firewall:     ✗ Disabled")
	}

	// Check anchor loaded
	out, _ = exec.Command("sudo", "pfctl", "-s", "Anchors").CombinedOutput()
	if strings.Contains(string(out), anchorName) {
		fmt.Println("Anchor loaded:   ✓ Yes")
	} else {
		fmt.Println("Anchor loaded:   ✗ No")
	}

	// Show rules if loaded
	out, _ = exec.Command("sudo", "pfctl", "-a", anchorName, "-s", "rules").CombinedOutput()
	outStr := string(out)
	// Filter out ALTQ warnings
	lines := strings.Split(outStr, "\n")
	var rules []string
	for _, line := range lines {
		if !strings.Contains(line, "ALTQ") && strings.TrimSpace(line) != "" {
			rules = append(rules, line)
		}
	}
	if len(rules) > 0 {
		fmt.Println("\nActive rules:")
		for _, rule := range rules {
			fmt.Println("  " + rule)
		}
	}
}
