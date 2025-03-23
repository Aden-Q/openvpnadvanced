package dnsmasq

import (
	"fmt"
	"os"
)

type ResolvedResult struct {
	Domain      string
	IP          string
	ShouldRoute bool
}

// ExportVPNIPs writes all VPN-targeted domain-IP mappings to a file
func ExportVPNIPs(results []ResolvedResult, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, res := range results {
		if res.ShouldRoute && res.IP != "" {
			line := fmt.Sprintf("%s %s\n", res.IP, res.Domain)
			_, err := file.WriteString(line)
			if err != nil {
				return err
			}
		}
	}

	fmt.Printf("✅ Exported VPN IPs to %s\n", outputPath)
	return nil
}
