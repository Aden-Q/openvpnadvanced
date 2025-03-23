package fetcher

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func FetchAndMergeRules(subscriptionFile, outputFile string) error {
	urls, err := readSubscriptionURLs(subscriptionFile)
	if err != nil {
		return err
	}

	ruleSet := make(map[string]struct{})

	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Failed to fetch %s: %v\n", url, err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Failed to read from %s: %v\n", url, err)
			continue
		}

		lines := strings.Split(string(body), "\n")
		for _, line := range lines {
			rule := strings.TrimSpace(line)
			if rule == "" || strings.HasPrefix(rule, "#") {
				continue
			}
			ruleSet[rule] = struct{}{}
		}
	}

	// Write merged rules to file
	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()

	for rule := range ruleSet {
		out.WriteString(rule + "\n")
	}

	fmt.Printf("✅ Merged %d unique rules into %s\n", len(ruleSet), outputFile)
	return nil
}

func readSubscriptionURLs(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			urls = append(urls, line)
		}
	}
	return urls, nil
}
