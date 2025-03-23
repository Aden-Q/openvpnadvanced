package main

import (
	"fmt"
	"log"
	"time"

	"openvpnadvanced/dnsmasq"
	"openvpnadvanced/dnsproxy"
	"openvpnadvanced/fetcher"
	"openvpnadvanced/vpn"
)

func main() {
	fmt.Println("🚀 openvpnadvanced starting...")

	// Step 1: Fetch and merge rules from GitHub
	err := fetcher.FetchAndMergeRules("assets/subscriptions.txt", "assets/merged_rule.list")
	if err != nil {
		log.Fatalf("Failed to fetch subscriptions: %v", err)
	}
	fmt.Println("✅ Merged rules into assets/merged_rule.list")

	// Step 2: Load DNS cache from file
	rawCache, err := dnsmasq.LoadCacheFromFile()
	if err != nil {
		log.Fatalf("Failed to load DNS cache: %v", err)
	}
	cache := dnsmasq.NewCacheWithTTL(10 * time.Minute)
	for domain, record := range rawCache {
		cache.Set(domain, record.IP)
	}

	// Step 3: Load routing rules
	rules, err := dnsmasq.LoadDomainRules("assets/merged_rule.list")
	if err != nil {
		log.Fatalf("Failed to load rule list: %v", err)
	}

	// Step 4: Check if VPN is up and get interface
	if !vpn.IsTunnelblickRunning() {
		log.Fatalf("Tunnelblick is not running. Please start your OpenVPN profile.")
	}

	iface, err := vpn.FindVPNInterface()
	if err != nil {
		log.Fatalf("No VPN interface found: %v", err)
	}
	fmt.Printf("✅ VPN interface detected: %s\n", iface)

	// Step 5: Remove catch-all VPN routes
	err = vpn.DeleteDefaultVPNRoutes()
	if err != nil {
		log.Printf("Warning: failed to delete default VPN routes: %v", err)
	}

	// Step 6: Start DNS server
	dnsServer := dnsproxy.NewServer(rules, cache, "8.8.8.8:53", iface)
	dnsServer.Start()

	// Step 7: Periodically save cache to disk
	go func() {
		for {
			time.Sleep(30 * time.Second)
			err := dnsmasq.SaveCacheToFile(cache)
			if err != nil {
				log.Printf("Failed to save cache: %v", err)
			} else {
				log.Println("✅ Cache saved to cache.json")
			}
		}
	}()

	// Step 8: Keep program running forever
	select {}

	// Step 9: Optional manual test (uncomment to use)
	// fmt.Println("=== Manual Route Check ===")
	// err = vpn.HijackIPv6("www.facebook.com", iface)
	// if err != nil {
	//     log.Printf("Failed IPv6 hijack: %v", err)
	// }
	//
	// err = utils.CurlVerify("www.facebook.com", "2a03:2880:f138:83:face:b00c:0:25de")
	// if err != nil {
	//     log.Printf("Curl test failed: %v", err)
	// }
	//
	// routeIface, err := vpn.GetRouteInterface("31.13.72.36")
	// if err != nil {
	//     log.Printf("Failed to get route interface: %v", err)
	// } else {
	//     fmt.Printf("IP 31.13.72.36 is routed via interface: %s\n", routeIface)
	// }
}
