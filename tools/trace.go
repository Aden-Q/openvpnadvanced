package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"openvpnadvanced/dnsmasq"
	"openvpnadvanced/vpn"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run tools/trace.go <domain>")
		os.Exit(1)
	}

	domain := os.Args[1]
	fmt.Printf("🔍 Tracing domain: %s\n", domain)

	// 1. Load routing rules
	rules, err := dnsmasq.LoadDomainRules("assets/merged_rule.list")
	if err != nil {
		log.Fatalf("Failed to load rules: %v", err)
	}

	// 2. Load and prepare cache
	rawCache, _ := dnsmasq.LoadCacheFromFile()
	cache := dnsmasq.NewCacheWithTTL(10 * time.Minute)
	for domain, record := range rawCache {
		cache.Set(domain, record.IP)
	}

	// 3. Resolve domain (recursively handles CNAME)
	shouldRoute, ip, cname := dnsmasq.ResolveWithCNAME(domain, rules, cache)
	if ip == "" {
		fmt.Println("❌ Failed to resolve domain.")
		return
	}

	// 4. Determine current routing interface
	iface, err := vpn.GetRouteInterface(ip)
	if err != nil {
		fmt.Printf("⚠️ Could not determine route interface for %s: %v\n", ip, err)
	}

	fmt.Println("——————————————")
	fmt.Printf("Domain:        %s\n", domain)
	fmt.Printf("Resolved IP:   %s\n", ip)
	fmt.Printf("Matched Rule:  %v\n", shouldRoute)
	if cname != "" && cname != domain {
		fmt.Printf("CNAME Chain:    %s ➜ %s\n", domain, cname)
	}
	if err == nil {
		isVPN := strings.HasPrefix(iface, "utun")
		if isVPN && shouldRoute {
			fmt.Println("✅ This domain is routed through VPN")
		} else {
			fmt.Println("☁️ This domain is routed directly (not via VPN)")
		}
	}
	fmt.Println("——————————————")
}
