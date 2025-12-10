package cmd

import (
	"fmt"
	"os"
)

func Execute() {
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
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`saferay - DNS leak protection for macOS with Xray/Hiddify

Usage:
  saferay install              Install saferay to /usr/local/bin
  saferay uninstall            Remove saferay from system

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
