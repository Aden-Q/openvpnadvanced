package cli

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"openvpnadvanced/cmd/config"
	"openvpnadvanced/cmd/core"
	"openvpnadvanced/dnsmasq"
	"openvpnadvanced/fetcher"
	"openvpnadvanced/vpn"
)

func printHelp() {
	fmt.Println(`Commands:
  help - Show this help message
  auto-subscribe true/false - Enable/Disable auto-subscribe
  update-period <duration> - Set auto-update period (e.g. 30m, 1h)
  update-now - Force a manual subscription update
  show-config - Show current configuration
  show-iface - Show current VPN interface info
  reload-config - Reload config.ini without restarting
  exit - Exit the program
  check-openvpn-on - Enable OpenVPN check
  check-openvpn-off - Disable OpenVPN check
  start - Start the main DNS/VPN logic
  startv - Start the main logic and print to console
  view-log err - Show only log lines with [ERROR]
  view-log info - Show all logs
  view-log direct - Show lines with [DIRECT]
  view-log vpn - Show lines with [VPN]
  set-log-level info/err/vpn - Set logging level
  clear-logs - Clear all log files
  compress-logs - Archive all log files into a zip
  clear - Clear console output
  test <domain> - Check if a domain will be routed via VPN or direct
  rtest <domain> - Check routing and interface info for a domain
  status - Show current running status of the core and VPN client`)
}

func printStatus() {
	if core.IsCoreStarted() {
		fmt.Println("✅ Core logic is running.")
	} else {
		fmt.Println("🛑 Core logic is not running.")
	}
	cfg := config.GetConfig()
	if cfg.CheckOpenVPN {
		if vpn.IsTunnelblickRunning() {
			fmt.Println("✅ OpenVPN (Tunnelblick) is running.")
		} else {
			fmt.Println("❌ OpenVPN (Tunnelblick) is NOT running.")
		}
	} else {
		fmt.Println("⚠️ OpenVPN check is disabled.")
	}
}

func handleAutoSubscribe(parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("missing value: true or false")
	}
	cfg := config.GetConfig()
	cfg.AutoSubscribe = (parts[1] == "true")
	config.SetConfig(cfg)
	return config.SaveINIConfig("config.ini")
}

func handleUpdatePeriod(parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("missing duration string")
	}
	dur, err := time.ParseDuration(parts[1])
	if err != nil {
		return fmt.Errorf("invalid duration: %v", err)
	}
	cfg := config.GetConfig()
	cfg.UpdatePeriod = dur
	config.SetConfig(cfg)
	return config.SaveINIConfig("config.ini")
}

func handleUpdateNow() error {
	fmt.Println("⏳ Manually updating subscriptions...")
	cfg := config.GetConfig()
	if cfg.AutoSubscribe {
		err := fetcher.FetchAndMergeRules("assets/subscriptions.txt", "assets/merged_rule.list")
		if err != nil {
			fmt.Println("❌ Subscription update failed:", err)
			return err
		}
		fmt.Println("✅ Subscription updated.")
	} else {
		fmt.Println("Auto Subscribe is disabled. Skipping update.")
	}
	return nil
}

func showConfig() {
	cfg := config.GetConfig()
	settings := map[string]string{
		"Auto Subscribe": fmt.Sprintf("%v", cfg.AutoSubscribe),
		"Update Period":  cfg.UpdatePeriod.String(),
		"Check OpenVPN":  fmt.Sprintf("%v", cfg.CheckOpenVPN),
		"Log Level":      cfg.LogLevel,
	}

	// Calculate max widths
	maxKeyLen := 0
	maxValLen := 0
	for k, v := range settings {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
		if len(v) > maxValLen {
			maxValLen = len(v)
		}
	}
	border := fmt.Sprintf("+-%s-+-%s-+", strings.Repeat("-", maxKeyLen), strings.Repeat("-", maxValLen))
	fmt.Println(border)
	fmt.Printf("| %-*s | %-*s |\n", maxKeyLen, "Setting", maxValLen, "Value")
	fmt.Println(border)
	for k, v := range settings {
		fmt.Printf("| %-*s | %-*s |\n", maxKeyLen, k, maxValLen, v)
	}
	fmt.Println(border)
}

func handleCheckOpenVPN(enable bool) error {
	cfg := config.GetConfig()
	cfg.CheckOpenVPN = enable
	config.SetConfig(cfg)
	return config.SaveINIConfig("config.ini")
}

func handleStart(verbose bool) error {
	if core.IsCoreStarted() {
		if verbose {
			fmt.Println("⚠️ Core logic is already running.")
		}
		return nil
	}
	if verbose {
		return core.RunCoreLogic(true)
	}
	go func() {
		if err := core.RunCoreLogic(false); err != nil {
			fmt.Printf("Error starting core: %v\n", err)
		}
	}()
	time.Sleep(1 * time.Second)
	fmt.Println("✅ Core started silently in background. Logs are written to logs/app.log")
	return nil
}

func handleShowIface() error {
	iface, err := vpn.FindVPNInterface()
	if err != nil {
		return fmt.Errorf("VPN interface not found: %v", err)
	}
	addrs, _ := net.InterfaceByName(iface)
	ipList, _ := addrs.Addrs()
	fmt.Printf("✅ VPN Interface: %s\n", iface)
	for _, addr := range ipList {
		fmt.Printf("   ➜ %s\n", addr.String())
	}
	return nil
}

func handleReloadConfig() error {
	if err := config.LoadINIConfig("config.ini"); err != nil {
		return fmt.Errorf("failed to reload config.ini: %v", err)
	}
	fmt.Println("✅ Configuration reloaded.")
	return nil
}

func handleSetLogLevel(parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("missing log level: info, err, or vpn")
	}
	cfg := config.GetConfig()
	cfg.LogLevel = parts[1]
	config.SetConfig(cfg)
	if err := config.SaveINIConfig("config.ini"); err != nil {
		return err
	}
	fmt.Println("Log level set to:", cfg.LogLevel)
	return nil
}

func handleViewLog(parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("missing argument for view-log")
	}
	filter := parts[1]
	logPath := "logs/app.log"

	viewLog := func(filter string) {
		content, err := os.ReadFile(logPath)
		if err != nil {
			fmt.Println("Failed to read log file:", err)
			return
		}
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			switch filter {
			case "info":
				if strings.TrimSpace(line) != "" {
					fmt.Println(line)
				}
			case "err":
				if strings.Contains(line, "[ERROR]") {
					fmt.Println(line)
				}
			case "direct":
				if strings.Contains(line, "[DIRECT]") {
					fmt.Println(line)
				}
			case "vpn":
				if strings.Contains(line, "[VPN]") {
					fmt.Println(line)
				}
			}
		}
	}

	viewLog(filter)
	return nil
}

func handleClearLogs() error {
	fmt.Println("Clearing all logs...")
	_ = os.Truncate("logs/app.log", 0)
	_ = os.Truncate("logs/err.log", 0)
	_ = os.Truncate("logs/vpn.log", 0)
	fmt.Println("✅ Logs cleared.")
	return nil
}

func handleCompressLogs() error {
	timestamp := time.Now().Format("20060102_150405")
	archive := fmt.Sprintf("logs/archive_%s.zip", timestamp)
	cmd := exec.Command("zip", "-j", archive, "logs/app.log", "logs/err.log", "logs/vpn.log")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to compress logs: %v", err)
	}
	fmt.Println("✅ Logs compressed into", archive)
	return nil
}

func handleClear() error {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func handleTest(parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("usage: test <domain>")
	}
	domain := parts[1]

	rules, err := dnsmasq.LoadDomainRules("assets/merged_rule.list")
	if err != nil {
		return fmt.Errorf("failed to load domain rules: %v", err)
	}

	if dnsmasq.MatchesRules(domain, rules) {
		fmt.Printf("🔒 %s ➜ Routed via VPN (matched rule)\n", domain)
	} else {
		fmt.Printf("🌐 %s ➜ Direct connection (no match)\n", domain)
	}
	return nil
}

func handleRTest(parts []string) error {
	if len(parts) < 2 {
		return fmt.Errorf("usage: rtest <domain>")
	}
	domain := parts[1]

	rules, err := dnsmasq.LoadDomainRules("assets/merged_rule.list")
	if err != nil {
		return fmt.Errorf("failed to load domain rules: %v", err)
	}

	matched := dnsmasq.MatchesRules(domain, rules)
	ipList, err := net.LookupIP(domain)
	if err != nil || len(ipList) == 0 {
		return fmt.Errorf("DNS lookup failed: %v", err)
	}

	ip := ipList[0].String()
	routeIface, err := vpn.GetRouteInterface(ip)
	if err != nil {
		return fmt.Errorf("could not determine interface for %s (%s): %v", domain, ip, err)
	}

	if matched {
		fmt.Printf("🔒 %s ➜ Routed via VPN\n", domain)
	} else {
		fmt.Printf("🌐 %s ➜ Direct connection\n", domain)
	}
	fmt.Printf("   ➜ Interface: %s (IP: %s)\n", routeIface, ip)
	return nil
}
