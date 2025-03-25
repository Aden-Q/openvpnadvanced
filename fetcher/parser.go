package fetcher

import (
	"bufio"
	"os"
	"strings"
)

// ParseRules 读取规则文件并返回规则列表
func ParseRules(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rules []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rules = append(rules, line)
	}
	return rules, scanner.Err()
}

// MatchRule 判断一个域名是否匹配规则列表
func MatchRule(domain string, rules []string) bool {
	for _, rule := range rules {
		if strings.HasPrefix(rule, "DOMAIN-SUFFIX,") {
			suffix := strings.TrimPrefix(rule, "DOMAIN-SUFFIX,")
			if strings.HasSuffix(domain, suffix) {
				return true
			}
		}
	}
	return false
}
