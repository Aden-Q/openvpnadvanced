package vpn

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
)

// IsTunnelblickRunning checks if Tunnelblick is running
func IsTunnelblickRunning() bool {
	out, err := exec.Command("pgrep", "-f", "Tunnelblick").Output()
	return err == nil && len(out) > 0
}

// FindVPNInterface returns the first utun interface that has an IPv4 address
func FindVPNInterface() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if strings.HasPrefix(iface.Name, "utun") && iface.Flags&net.FlagUp != 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
					return iface.Name, nil
				}
			}
		}
	}
	return "", errors.New("no active utun interface with IPv4 found")
}

// CorrectDefaultRoute resets the system default route to the local gateway
func CorrectDefaultRoute() error {
	// Step 1: Get current default gateway via `route -n get default`
	cmd := exec.Command("route", "-n", "get", "default")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get default route: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	var gateway string
	re := regexp.MustCompile(`gateway:\s+([0-9.]+)`)

	for scanner.Scan() {
		line := scanner.Text()
		if matches := re.FindStringSubmatch(line); len(matches) == 2 {
			gateway = matches[1]
			break
		}
	}

	if gateway == "" {
		return errors.New("could not find default gateway from route output")
	}

	// Step 2: Delete all default routes (may need to run multiple times)
	for i := 0; i < 3; i++ {
		_ = exec.Command("sudo", "route", "delete", "default").Run()
	}

	// Step 3: Add back default route to real gateway
	cmd = exec.Command("sudo", "route", "add", "default", gateway)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add corrected default route: %w", err)
	}

	fmt.Printf("✅ Corrected default route to local gateway: %s\n", gateway)
	return nil
}
