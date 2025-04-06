package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"openvpnadvanced/dnsmasq"
	"openvpnadvanced/vpn"

	"github.com/olekukonko/tablewriter"
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

	// 4. Get VPN interface
	vpnIface, err := vpn.FindVPNInterface()
	if err != nil {
		fmt.Printf("❌ Failed to get VPN interface: %v\n", err)
		return
	}

	// 5. Get current routing interface
	currentIface, err := vpn.GetRouteInterface(ip)
	if err != nil {
		fmt.Printf("❌ Failed to get routing interface: %v\n", err)
		return
	}

	// 6. Get default gateway
	defaultGateway, defaultIface, err := vpn.GetDefaultGateway()
	if err != nil {
		fmt.Printf("❌ Failed to get default gateway: %v\n", err)
		return
	}

	// Display network information
	fmt.Println("\n📊 Network Information")
	networkTable := tablewriter.NewWriter(os.Stdout)
	networkTable.SetHeader([]string{"Item", "Value"})
	networkTable.SetAutoWrapText(false)
	networkTable.SetAlignment(tablewriter.ALIGN_LEFT)
	networkTable.SetBorder(false)
	networkTable.Append([]string{"Domain", domain})
	networkTable.Append([]string{"Resolved IP", ip})
	networkTable.Append([]string{"Matched Rule", map[bool]string{true: "VPN", false: "DIRECT"}[shouldRoute]})
	if cname != "" && cname != domain {
		networkTable.Append([]string{"CNAME Chain", domain + " -> " + cname})
	}
	networkTable.Render()

	// Display routing information
	fmt.Println("\n🛣️ Routing Information")
	routeTable := tablewriter.NewWriter(os.Stdout)
	routeTable.SetHeader([]string{"Item", "Value"})
	routeTable.SetAutoWrapText(false)
	routeTable.SetAlignment(tablewriter.ALIGN_LEFT)
	routeTable.SetBorder(false)
	routeTable.Append([]string{"Current Interface", currentIface})
	routeTable.Append([]string{"VPN Interface", vpnIface})
	routeTable.Append([]string{"Default Gateway", defaultGateway})
	routeTable.Append([]string{"Default Gateway Interface", defaultIface})
	routeTable.Render()

	// Check if routing is optimal
	if shouldRoute && currentIface != vpnIface {
		fmt.Println("\n⚠️ Warning: Domain should be routed through VPN but is using direct connection")
		fmt.Printf("Attempting to fix routing...\n")
		if err := vpn.AddRoute(ip, vpnIface); err != nil {
			fmt.Printf("❌ Failed to fix routing: %v\n", err)
		} else {
			fmt.Println("✅ Routing fixed successfully")
			// Get updated routing interface
			currentIface, _ = vpn.GetRouteInterface(ip)
			fmt.Printf("Updated routing interface: %s\n", currentIface)
		}
	} else if !shouldRoute && currentIface == vpnIface {
		fmt.Println("\n⚠️ Warning: Domain should be routed directly but is using VPN")
		fmt.Println("Suggestion: Check routing rule configuration")
	} else {
		fmt.Println("\n✅ Routing status is normal")
	}

	fmt.Println("\n—————————————— End of Trace ——————————————")
}
