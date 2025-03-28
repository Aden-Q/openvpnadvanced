package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/peterh/liner"
	"gopkg.in/ini.v1"

	"openvpnadvanced/dnsmasq"
	"openvpnadvanced/dnsproxy"
	"openvpnadvanced/fetcher"
	"openvpnadvanced/vpn"
)

type AppConfig struct {
	AutoSubscribe bool
	UpdatePeriod  time.Duration
	CheckOpenVPN  bool
	LogLevel      string
}

var appConfig AppConfig
var logFile *os.File
var errLogFile *os.File
var vpnLogFile *os.File
var coreStarted bool

func loadINIConfig(path string) error {
	cfg, err := ini.Load(path)
	if err != nil {
		return err
	}
	appConfig.AutoSubscribe = cfg.Section("").Key("auto-subscribe").MustBool(false)
	appConfig.UpdatePeriod = cfg.Section("").Key("update-period").MustDuration(30 * time.Minute)
	appConfig.CheckOpenVPN = cfg.Section("").Key("check-openvpn").MustBool(true)
	appConfig.LogLevel = cfg.Section("").Key("log-level").MustString("info")
	return nil
}

func saveINIConfig(path string) error {
	cfg := ini.Empty()
	cfg.Section("").Key("auto-subscribe").SetValue(fmt.Sprintf("%v", appConfig.AutoSubscribe))
	cfg.Section("").Key("update-period").SetValue(appConfig.UpdatePeriod.String())
	cfg.Section("").Key("check-openvpn").SetValue(fmt.Sprintf("%v", appConfig.CheckOpenVPN))
	cfg.Section("").Key("log-level").SetValue(appConfig.LogLevel)
	return cfg.SaveTo(path)
}

func showConfig() {
	cfg := map[string]string{
		"Auto Subscribe": fmt.Sprintf("%v", appConfig.AutoSubscribe),
		"Update Period":  appConfig.UpdatePeriod.String(),
		"Check OpenVPN":  fmt.Sprintf("%v", appConfig.CheckOpenVPN),
		"Log Level":      appConfig.LogLevel,
	}

	// Calculate max widths
	maxKeyLen := 0
	maxValLen := 0
	for k, v := range cfg {
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
	for k, v := range cfg {
		fmt.Printf("| %-*s | %-*s |\n", maxKeyLen, k, maxValLen, v)
	}
	fmt.Println(border)
}

func manualUpdate() {
	fmt.Println("⏳ Manually updating subscriptions...")
	if appConfig.AutoSubscribe {
		err := fetcher.FetchAndMergeRules("assets/subscriptions.txt", "assets/merged_rule.list")
		if err != nil {
			fmt.Println("❌ Subscription update failed:", err)
		} else {
			fmt.Println("✅ Subscription updated.")
		}
	} else {
		fmt.Println("Auto Subscribe is disabled. Skipping update.")
	}
}

func runCoreLogic(verbose bool) {
	if coreStarted {
		if verbose {
			fmt.Println("⚠️ Core logic is already running.")
		}
		log.Println("Core logic already started.")
		return
	}
	coreStarted = true

	if !verbose {
		log.SetOutput(logFile)
	} else {
		log.SetOutput(os.Stdout)
	}

	if appConfig.AutoSubscribe {
		err := fetcher.FetchAndMergeRules("assets/subscriptions.txt", "assets/merged_rule.list")
		if err != nil {
			log.Fatalf("Failed to fetch subscriptions: %v", err)
		}
		if verbose {
			fmt.Println("✅ Merged rules into assets/merged_rule.list")
		}
	}

	// Load DNS cache from file
	rawCache, err := dnsmasq.LoadCacheFromFile()
	if err != nil {
		log.Fatalf("Failed to load DNS cache: %v", err)
	}
	cache := dnsmasq.NewCacheWithTTL(10 * time.Minute)
	for domain, record := range rawCache {
		cache.Set(domain, record.IP)
	}

	// Load routing rules
	rules, err := dnsmasq.LoadDomainRules("assets/merged_rule.list")
	if err != nil {
		log.Fatalf("Failed to load rule list: %v", err)
	}

	// Check if VPN is up and get interface
	if appConfig.CheckOpenVPN && !vpn.IsTunnelblickRunning() {
		log.Fatalf("Tunnelblick is not running. Please start your OpenVPN profile.")
	}

	iface, err := vpn.FindVPNInterface()
	if err != nil {
		log.Fatalf("No VPN interface found: %v", err)
	}
	if verbose {
		fmt.Printf("✅ VPN interface detected: %s\n", iface)
	}
	log.Printf("VPN interface detected: %s\n", iface)
	if appConfig.LogLevel == "info" || appConfig.LogLevel == "vpn" {
		_, err = vpnLogFile.WriteString(fmt.Sprintf("[VPN] Interface detected: %s\n", iface))
		if err != nil {
			log.Printf("Warning: failed to write to vpn.log: %v", err)
		}
	}

	// Remove catch-all VPN routes
	err = vpn.DeleteDefaultVPNRoutes()
	if err != nil {
		log.Printf("Warning: failed to delete default VPN routes: %v", err)
	}

	err = vpn.CorrectDefaultRoute()
	if err != nil {
		log.Printf("Warning: failed to correct default route: %v", err)
	}

	// Start DNS server
	if verbose {
		fmt.Printf("🧠 Loaded %d domain rules\n", len(rules))
		fmt.Println("🚦 Starting DNS proxy server...")
	}
	dnsServer := dnsproxy.NewServer(rules, cache, "8.8.8.8:53", iface)
	dnsServer.Start()

	// Periodically save cache to disk
	go func() {
		for {
			time.Sleep(30 * time.Second)
			err := dnsmasq.SaveCacheToFile(cache)
			if err != nil {
				msg := fmt.Sprintf("Failed to save cache: %v", err)
				log.Println("[ERROR]", msg)
				if appConfig.LogLevel == "err" || appConfig.LogLevel == "info" {
					_, _ = logFile.WriteString("[ERROR] " + msg + "\n")
					_, _ = errLogFile.WriteString("[ERROR] " + msg + "\n")
				}
			} else {
				successMsg := "✅ Cache saved to cache.json"
				log.Println(successMsg)
				if appConfig.LogLevel == "info" {
					_, _ = logFile.WriteString(successMsg + "\n")
				}
			}
		}
	}()
}

func startConsole() {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	line.SetCompleter(func(line string) (c []string) {
		commands := []string{
			"help", "auto-subscribe true", "auto-subscribe false", "update-period", "update-now",
			"show-config", "show-iface", "reload-config", "exit",
			"check-openvpn-on", "check-openvpn-off", "start", "startv",
			"view-log err", "view-log info", "view-log direct", "view-log vpn",
			"set-log-level info", "set-log-level err", "set-log-level vpn",
			"clear-logs", "compress-logs", "clear", "test", "rtest",
			"status",
		}
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, line) {
				c = append(c, cmd)
			}
		}
		return
	})

	for {
		input, err := line.Prompt("ovpnctl> ")
		if err != nil {
			break
		}
		input = strings.TrimSpace(input)
		line.AppendHistory(input)

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "help":
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
		case "exit":
			return
		case "status":
			if coreStarted {
				fmt.Println("✅ Core logic is running.")
			} else {
				fmt.Println("🛑 Core logic is not running.")
			}
			if appConfig.CheckOpenVPN {
				if vpn.IsTunnelblickRunning() {
					fmt.Println("✅ OpenVPN (Tunnelblick) is running.")
				} else {
					fmt.Println("❌ OpenVPN (Tunnelblick) is NOT running.")
				}
			} else {
				fmt.Println("⚠️ OpenVPN check is disabled.")
			}
		case "auto-subscribe":
			if len(parts) < 2 {
				fmt.Println("Missing value: true or false")
				continue
			}
			appConfig.AutoSubscribe = (parts[1] == "true")
			_ = saveINIConfig("config.ini")
		case "update-period":
			if len(parts) < 2 {
				fmt.Println("Missing duration string")
				continue
			}
			dur, err := time.ParseDuration(parts[1])
			if err != nil {
				fmt.Println("Invalid duration:", err)
			} else {
				appConfig.UpdatePeriod = dur
				_ = saveINIConfig("config.ini")
			}
		case "update-now":
			manualUpdate()
		case "show-config":
			showConfig()
		case "check-openvpn-on":
			appConfig.CheckOpenVPN = true
			_ = saveINIConfig("config.ini")
		case "check-openvpn-off":
			appConfig.CheckOpenVPN = false
			_ = saveINIConfig("config.ini")
		case "set-log-level":
			if len(parts) < 2 {
				fmt.Println("Missing log level: info, err, or vpn")
				continue
			}
			appConfig.LogLevel = parts[1]
			_ = saveINIConfig("config.ini")
			fmt.Println("Log level set to:", appConfig.LogLevel)
		case "start":
			if coreStarted {
				fmt.Println("⚠️ Core logic is already running.")
			} else {
				go runCoreLogic(false)
				time.Sleep(1 * time.Second) // Brief delay to ensure goroutine initiates
				fmt.Println("✅ Core started silently in background. Logs are written to logs/app.log")
			}
		case "startv":
			runCoreLogic(true)
		case "show-iface":
			iface, err := vpn.FindVPNInterface()
			if err != nil {
				fmt.Println("❌ VPN interface not found:", err)
			} else {
				addrs, _ := net.InterfaceByName(iface)
				ipList, _ := addrs.Addrs()
				fmt.Printf("✅ VPN Interface: %s\n", iface)
				for _, addr := range ipList {
					fmt.Printf("   ➜ %s\n", addr.String())
				}
			}
		case "reload-config":
			if err := loadINIConfig("config.ini"); err != nil {
				fmt.Println("❌ Failed to reload config.ini:", err)
			} else {
				fmt.Println("✅ Configuration reloaded.")
			}
		case "rtest":
			if len(parts) < 2 {
				fmt.Println("Usage: rtest <domain>")
				continue
			}
			domain := parts[1]
			rules, err := dnsmasq.LoadDomainRules("assets/merged_rule.list")
			if err != nil {
				fmt.Println("Failed to load domain rules:", err)
				continue
			}
			matched := dnsmasq.MatchesRules(domain, rules)
			ipList, err := net.LookupIP(domain)
			if err != nil || len(ipList) == 0 {
				fmt.Println("❌ DNS lookup failed:", err)
				continue
			}
			ip := ipList[0].String()
			routeIface, err := vpn.GetRouteInterface(ip)
			if err != nil {
				fmt.Printf("❌ Could not determine interface for %s (%s): %v\n", domain, ip, err)
				continue
			}
			if matched {
				fmt.Printf("🔒 %s ➜ Routed via VPN\n", domain)
			} else {
				fmt.Printf("🌐 %s ➜ Direct connection\n", domain)
			}
			fmt.Printf("   ➜ Interface: %s (IP: %s)\n", routeIface, ip)
		case "test":
			if len(parts) < 2 {
				fmt.Println("Usage: test <domain>")
				continue
			}
			domain := parts[1]

			rules, err := dnsmasq.LoadDomainRules("assets/merged_rule.list")
			if err != nil {
				fmt.Println("Failed to load domain rules:", err)
				continue
			}

			if dnsmasq.MatchesRules(domain, rules) {
				fmt.Printf("🔒 %s ➜ Routed via VPN (matched rule)\n", domain)
			} else {
				fmt.Printf("🌐 %s ➜ Direct connection (no match)\n", domain)
			}
		case "clear-logs":
			fmt.Println("Clearing all logs...")
			_ = os.Truncate("logs/app.log", 0)
			_ = os.Truncate("logs/err.log", 0)
			_ = os.Truncate("logs/vpn.log", 0)
			fmt.Println("✅ Logs cleared.")
		case "compress-logs":
			timestamp := time.Now().Format("20060102_150405")
			archive := fmt.Sprintf("logs/archive_%s.zip", timestamp)
			cmd := exec.Command("zip", "-j", archive, "logs/app.log", "logs/err.log", "logs/vpn.log")
			if err := cmd.Run(); err != nil {
				fmt.Println("❌ Failed to compress logs:", err)
			} else {
				fmt.Println("✅ Logs compressed into", archive)
			}
		case "view-log":
			if len(parts) < 2 {
				fmt.Println("Missing argument for view-log")
				continue
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

			for {
				viewLog(filter)
				fmt.Println("\n(Type 'back' to return to console)")
				backInput, err := line.Prompt("view-log> ")
				if err != nil {
					break
				}
				backInput = strings.TrimSpace(strings.ToLower(backInput))
				if backInput == "back" {
					break
				}
			}
		case "clear":
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			_ = cmd.Run()
		default:
			fmt.Println("Unknown command:", parts[0])
		}
	}
}

func main() {
	// Ensure logs directory exists
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		_ = os.Mkdir("logs", 0755)
	}

	err := loadINIConfig("config.ini")
	if err != nil {
		log.Fatalf("Failed to load config.ini: %v", err)
	}

	logFile, err = os.OpenFile("logs/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	errLogFile, _ = os.OpenFile("logs/err.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	vpnLogFile, _ = os.OpenFile("logs/vpn.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	log.SetOutput(logFile)

	if len(os.Args) > 1 && os.Args[1] == "--start" {
		startConsole()
	} else {
		fmt.Println(`Usage:
  sudo ./openvpnadvanced --start     Launch interactive console
  sudo ./openvpnadvanced             Show this help message`)
		os.Exit(0)
	}
}
