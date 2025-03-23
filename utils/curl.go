package utils

import (
	"fmt"
	"os/exec"
)

func CurlHTTPS(host, ip string) error {
	url := fmt.Sprintf("https://%s", host)
	cmd := exec.Command("curl", "-v", "--resolve", fmt.Sprintf("%s:443:%s", host, ip), url)
	cmd.Stdout = nil
	cmd.Stderr = nil

	output, err := cmd.CombinedOutput()
	fmt.Printf("Curl output:\n%s\n", output)
	return err
}
