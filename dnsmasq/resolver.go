package dnsmasq

import (
	"bufio"
	"log"
	"net"
	"openvpnadvanced/doh"
	"os"
	"strings"
)

type Rule struct {
	Suffix string
}

func MatchesRules(domain string, rules []Rule) bool {
	// 将域名转换为小写，确保不受大小写影响
	domain = strings.ToLower(domain)

	for _, rule := range rules {
		// 将规则后缀转换为小写进行匹配
		if strings.HasSuffix(domain, strings.ToLower(rule.Suffix)) {
			return true
		}
	}
	return false
}

func ResolveRecursive(domain string, rules []Rule, cache *Cache) (bool, string) {
	visited := make(map[string]bool)
	current := domain
	originalDomain := domain // 保持原始域名

	for depth := 0; depth < 10; depth++ {
		if visited[current] {
			log.Printf("⚠️ Circular CNAME detected for %s", domain)
			return false, ""
		}
		visited[current] = true

		// 缓存检查（保留原始域名规则匹配）
		if cachedVal, ok := cache.Get(current); ok {
			if net.ParseIP(cachedVal) != nil {
				log.Printf("[CACHE] %s ➜ %s", current, cachedVal)
				return MatchesRules(originalDomain, rules), cachedVal // 使用原始域名进行匹配
			} else {
				log.Printf("[CACHE-CNAME] %s ➜ %s", current, cachedVal)
				current = cachedVal
				continue
			}
		}

		// DNS查询流程
		ip, cname, err := doh.QueryWithCNAME(current)
		if err == nil && ip != "" {
			log.Printf("[A] %s ➜ %s", current, ip)
			cache.Set(originalDomain, ip) // 使用原始域名缓存
			cache.Set(current, ip)
			return MatchesRules(originalDomain, rules), ip
		}

		ipv6, err := doh.QueryAAAA(current)
		if err == nil && ipv6 != "" {
			log.Printf("[AAAA] %s ➜ %s", current, ipv6)
			cache.Set(originalDomain, ipv6) // 使用原始域名缓存
			cache.Set(current, ipv6)
			return MatchesRules(originalDomain, rules), ipv6
		}

		if cname != "" {
			log.Printf("[CNAME] %s ➜ %s", current, cname)
			current = cname
			continue
		}

		// 后备查询逻辑
		allRecords, err := doh.QueryAll(current)
		if err == nil {
			for recordType, answers := range allRecords {
				for _, answer := range answers {
					if net.ParseIP(answer) != nil {
						log.Printf("[FALLBACK][%s] %s ➜ %s", recordType, current, answer)
						cache.Set(originalDomain, answer) // 使用原始域名缓存
						cache.Set(current, answer)
						return MatchesRules(originalDomain, rules), answer
					}
				}
			}
		}

		break
	}

	log.Printf("❌ Resolution failed for %s", domain)
	return false, ""
}

func LoadDomainRules(path string) ([]Rule, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rules []Rule
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "DOMAIN-SUFFIX,") {
			parts := strings.SplitN(line, ",", 2)
			if len(parts) == 2 {
				rules = append(rules, Rule{Suffix: strings.ToLower(parts[1])})
			}
		}
	}
	return rules, nil
}

func ResolveWithCNAME(domain string, rules []Rule, cache *Cache) (bool, string, string) {
	visited := make(map[string]bool)
	current := domain
	originalDomain := domain
	var firstCNAME string

	for depth := 0; depth < 10; depth++ {
		if visited[current] {
			log.Printf("⚠️ Circular CNAME detected for %s", domain)
			return false, "", ""
		}
		visited[current] = true

		// 缓存检查（保持规则匹配）
		if cachedVal, ok := cache.Get(current); ok {
			if net.ParseIP(cachedVal) != nil {
				log.Printf("[CACHE] %s ➜ %s", current, cachedVal)
				return MatchesRules(originalDomain, rules), cachedVal, firstCNAME
			} else {
				log.Printf("[CACHE-CNAME] %s ➜ %s", current, cachedVal)
				current = cachedVal
				continue
			}
		}

		// DNS查询流程
		ip, cname, err := doh.QueryWithCNAME(current)
		if err == nil && ip != "" {
			log.Printf("[A] %s ➜ %s", current, ip)
			cache.Set(originalDomain, ip) // 使用原始域名缓存
			cache.Set(current, ip)
			return MatchesRules(originalDomain, rules), ip, firstCNAME
		}

		ipv6, err := doh.QueryAAAA(current)
		if err == nil && ipv6 != "" {
			log.Printf("[AAAA] %s ➜ %s", current, ipv6)
			cache.Set(originalDomain, ipv6) // 使用原始域名缓存
			cache.Set(current, ipv6)
			return MatchesRules(originalDomain, rules), ipv6, firstCNAME
		}

		if cname != "" {
			log.Printf("[CNAME] %s ➜ %s", current, cname)
			if firstCNAME == "" {
				firstCNAME = cname
			}
			current = cname
			continue
		}

		// 后备查询逻辑
		allRecords, err := doh.QueryAll(current)
		if err == nil {
			for recordType, answers := range allRecords {
				for _, answer := range answers {
					if net.ParseIP(answer) != nil {
						log.Printf("[FALLBACK][%s] %s ➜ %s", recordType, current, answer)
						cache.Set(originalDomain, answer) // 使用原始域名缓存
						cache.Set(current, answer)
						return MatchesRules(originalDomain, rules), answer, firstCNAME
					}
				}
			}
		}

		break
	}

	log.Printf("❌ Resolution failed for %s", domain)
	return false, "", ""
}
