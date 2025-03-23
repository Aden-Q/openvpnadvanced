package utils

import "fmt"

const (
	ColorReset = "\033[0m"
	ColorRed   = "\033[31m"
	ColorGreen = "\033[32m"
)

// PrintVPN prints a VPN-tagged message in green
func PrintVPN(domain, ip string) {
	fmt.Printf("%s[VPN]   %-20s ➜ %s%s\n", ColorGreen, domain, ip, ColorReset)
}

// PrintDirect prints a direct-access message in default color
func PrintDirect(domain, ip string) {
	fmt.Printf("[DIRECT] %-20s ➜ %s\n", domain, ip)
}

// PrintError prints an error message in red
func PrintError(domain, msg string) {
	fmt.Printf("%s[ERROR] %-20s ➜ %s%s\n", ColorRed, domain, msg, ColorReset)
}
