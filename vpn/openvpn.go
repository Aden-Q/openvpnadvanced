package vpn

import (
	"errors"
	"net"
	"os/exec"
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
