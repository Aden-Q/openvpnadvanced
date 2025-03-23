package vpn

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
)

// AddRoute adds a static route to force <ip> to go through VPN interface
func AddRoute(ip, vpnInterface string) error {
	cmd := exec.Command("sudo", "route", "-n", "add", ip, "-interface", vpnInterface)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// DeleteDefaultVPNRoutes removes OpenVPN's default redirect routes
func DeleteDefaultVPNRoutes() error {
	log.Println("🧹 Removing default VPN catch-all routes (0.0.0.0/1 and 128.0.0.0/1)...")

	routes := [][]string{
		{"route", "-n", "delete", "0.0.0.0/1"},
		{"route", "-n", "delete", "128.0.0.0/1"},
	}

	for _, args := range routes {
		cmd := exec.Command("sudo", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Printf("⚠️ Failed to delete route: %v\n", args)
		}
	}
	return nil
}

// ResolveIPv6 resolves domain to its IPv6 addresses
func ResolveIPv6(domain string) ([]string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}

	var ipv6s []string
	for _, ip := range ips {
		if ip.To16() != nil && ip.To4() == nil {
			ipv6s = append(ipv6s, ip.String())
		}
	}
	return ipv6s, nil
}

// AddIPv6Route adds IPv6 route via specified interface
func AddIPv6Route(ip, iface string) error {
	cmd := exec.Command("sudo", "route", "-n", "add", "-inet6", ip, "-interface", iface)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// HijackIPv6 resolves domain and adds route for each IPv6 address
func HijackIPv6(domain, iface string) error {
	ips, err := ResolveIPv6(domain)
	if err != nil {
		return err
	}

	for _, ip := range ips {
		fmt.Printf("Adding IPv6 route for %s via %s\n", ip, iface)
		if err := AddIPv6Route(ip, iface); err != nil {
			fmt.Printf("[!] Failed to add IPv6 route: %v\n", err)
		}
	}
	return nil
}

// GetRouteInterface checks which interface is used to reach a given IP
func GetRouteInterface(ip string) (string, error) {
	cmd := exec.Command("route", "get", ip)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "interface:") {
			fields := strings.Fields(line)
			if len(fields) == 2 {
				return fields[1], nil
			}
		}
	}
	return "", fmt.Errorf("interface not found in route output")
}

// AddRouteForDomain resolves domain and adds both IPv4 and IPv6 routes
func AddRouteForDomain(domain, vpnInterface string) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		fmt.Printf("Failed to resolve domain %s: %v\n", domain, err)
		return
	}

	for _, ip := range ips {
		if ip.To4() != nil {
			ipStr := ip.String()
			fmt.Printf("add host %s: gateway %s\n", ipStr, vpnInterface)
			if err := AddRoute(ipStr, vpnInterface); err != nil {
				fmt.Printf("[!] Failed to add route: %v\n", err)
			} else {
				fmt.Printf("✅ Route added: %s ➜ %s\n", ipStr, vpnInterface)
			}
		} else {
			ipStr := ip.String()
			fmt.Printf("add IPv6 host %s: gateway %s\n", ipStr, vpnInterface)
			if err := AddIPv6Route(ipStr, vpnInterface); err != nil {
				fmt.Printf("[!] Failed to add IPv6 route: %v\n", err)
			} else {
				fmt.Printf("✅ IPv6 Route added: %s ➜ %s\n", ipStr, vpnInterface)
			}
		}
	}
}
