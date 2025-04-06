package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/peterh/liner"
)

func StartConsole() {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)
	line.SetCompleter(getCompleter())

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

		if err := handleCommand(parts); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func getCompleter() func(string) []string {
	return func(line string) (c []string) {
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
	}
}

func handleCommand(parts []string) error {
	switch parts[0] {
	case "help":
		printHelp()
	case "exit":
		os.Exit(0)
	case "status":
		printStatus()
	case "auto-subscribe":
		return handleAutoSubscribe(parts)
	case "update-period":
		return handleUpdatePeriod(parts)
	case "update-now":
		return handleUpdateNow()
	case "show-config":
		showConfig()
	case "check-openvpn-on":
		return handleCheckOpenVPN(true)
	case "check-openvpn-off":
		return handleCheckOpenVPN(false)
	case "start":
		return handleStart(false)
	case "startv":
		return handleStart(true)
	case "show-iface":
		return handleShowIface()
	case "reload-config":
		return handleReloadConfig()
	case "set-log-level":
		return handleSetLogLevel(parts)
	case "view-log":
		return handleViewLog(parts)
	case "clear-logs":
		return handleClearLogs()
	case "compress-logs":
		return handleCompressLogs()
	case "clear":
		return handleClear()
	case "test":
		return handleTest(parts)
	case "rtest":
		return handleRTest(parts)
	default:
		return fmt.Errorf("unknown command: %s", parts[0])
	}
	return nil
}

// 其他辅助函数...
