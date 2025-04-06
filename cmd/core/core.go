package core

import (
	"fmt"
	"log"
	"time"

	"openvpnadvanced/cmd/config"
	"openvpnadvanced/dnsmasq"
	"openvpnadvanced/dnsproxy"
	"openvpnadvanced/fetcher"
	"openvpnadvanced/vpn"
)

var coreStarted bool

func RunCoreLogic(verbose bool) error {
	if coreStarted {
		if verbose {
			fmt.Println("⚠️ Core logic is already running.")
		}
		log.Println("Core logic already started.")
		return nil
	}
	coreStarted = true

	cfg := config.GetConfig()

	if cfg.AutoSubscribe {
		err := fetcher.FetchAndMergeRules("assets/subscriptions.txt", "assets/merged_rule.list")
		if err != nil {
			return fmt.Errorf("failed to fetch subscriptions: %v", err)
		}
		if verbose {
			fmt.Println("✅ Merged rules into assets/merged_rule.list")
		}
	}

	// Load DNS cache from file
	rawCache, err := dnsmasq.LoadCacheFromFile()
	if err != nil {
		return fmt.Errorf("failed to load DNS cache: %v", err)
	}
	cache := dnsmasq.NewCacheWithTTL(10 * time.Minute)
	for domain, record := range rawCache {
		cache.Set(domain, record.IP)
	}

	// Load routing rules
	rules, err := dnsmasq.LoadDomainRules("assets/merged_rule.list")
	if err != nil {
		return fmt.Errorf("failed to load rule list: %v", err)
	}

	// Check if VPN is up and get interface
	if cfg.CheckOpenVPN && !vpn.IsTunnelblickRunning() {
		return fmt.Errorf("Tunnelblick is not running. Please start your OpenVPN profile")
	}

	iface, err := vpn.FindVPNInterface()
	if err != nil {
		return fmt.Errorf("no VPN interface found: %v", err)
	}
	if verbose {
		fmt.Printf("✅ VPN interface detected: %s\n", iface)
	}
	log.Printf("VPN interface detected: %s\n", iface)

	// Remove catch-all VPN routes
	if err := vpn.DeleteDefaultVPNRoutes(); err != nil {
		log.Printf("Warning: failed to delete default VPN routes: %v", err)
	}

	if err := vpn.CorrectDefaultRoute(); err != nil {
		log.Printf("Warning: failed to correct default route: %v", err)
	}

	// Start DNS server
	if verbose {
		fmt.Printf("🧠 Loaded %d domain rules\n", len(rules))
		fmt.Println("🚦 Starting DNS proxy server...")
	}
	dnsServer := dnsproxy.NewServer(rules, cache, "127.0.0.1:53", iface)
	dnsServer.Start()

	// Periodically save cache to disk
	go func() {
		for {
			time.Sleep(30 * time.Second)
			if err := dnsmasq.SaveCacheToFile(cache); err != nil {
				log.Printf("Failed to save cache: %v", err)
			}
		}
	}()

	return nil
}

func IsCoreStarted() bool {
	return coreStarted
}
