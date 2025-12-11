package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

const installPath = "/usr/local/bin/saferay"

func cmdInstallWithOptions(lightMode bool) {
	cmdInstall()

	if lightMode {
		fmt.Println()
		SetupLightMode()
	}
}

func cmdInstall() {
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		os.Exit(1)
	}

	if execPath == installPath {
		fmt.Println("saferay is already installed")
		return
	}

	// Copy binary to temp first
	src, err := os.Open(execPath)
	if err != nil {
		fmt.Printf("Error opening source: %v\n", err)
		os.Exit(1)
	}
	defer src.Close()

	tmpPath := "/tmp/saferay_install"
	dst, err := os.OpenFile(tmpPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		os.Exit(1)
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		fmt.Printf("Error copying binary: %v\n", err)
		os.Exit(1)
	}
	dst.Close()

	// Move to /usr/local/bin with sudo
	cmd := exec.Command("sudo", "mv", tmpPath, installPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error installing (need sudo): %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ saferay installed to /usr/local/bin/saferay")
}

func cmdUninstall() {
	// Remove binary
	cmd := exec.Command("sudo", "rm", "-f", installPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error removing binary: %v\n", err)
	}

	// Also cleanup DNS daemon, xray rules, and light mode
	removeDNSDaemon()
	resetXrayRules()

	// Reset light mode DNS if configured
	if _, err := os.Stat(lightConfigPath); err == nil {
		service := getActiveNetworkService()
		if service != "" {
			resetDNS(service)
		}
		_ = exec.Command("sudo", "rm", "-rf", "/etc/saferay").Run()
	}

	fmt.Println("✓ saferay uninstalled")
}
