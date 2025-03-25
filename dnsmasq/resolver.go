package dnsmasq

import (
	"bufio"
	"log"
	"openvpnadvanced/doh"
	"os"
	"strings"
)

type Rule struct {
	Suffix string
}

// MatchesRules checks if a domain matches any of the rules
func MatchesRules(domain string, rules []Rule) bool {
	for _, rule := range rules {
		if strings.HasSuffix(domain, rule.Suffix) {
			return true
		}
	}
	return false
}

// ResolveRecursive performs a full resolution: A, AAAA, CNAME fallback
func ResolveRecursive(domain string, rules []Rule, cache *Cache) (bool, string) {
	visited := make(map[string]bool)
	current := domain

	for depth := 0; depth < 10; depth++ {
		if visited[current] {
			log.Printf("⚠️ Circular CNAME detected for %s", domain)
			return false, ""
		}
		visited[current] = true

		if cachedIP, ok := cache.Get(current); ok {
			log.Printf("[CACHE] %s ➜ %s", current, cachedIP)
			return MatchesRules(domain, rules), cachedIP
		}

		ip, cname, err := doh.QueryWithCNAME(current)
		if err == nil && ip != "" {
			log.Printf("[A] %s ➜ %s", current, ip)
			cache.Set(current, ip)
			cache.Set(domain, ip)
			return MatchesRules(domain, rules), ip
		}

		ipv6, err := doh.QueryAAAA(current)
		if err == nil && ipv6 != "" {
			log.Printf("[AAAA] %s ➜ %s", current, ipv6)
			cache.Set(current, ipv6)
			cache.Set(domain, ipv6)
			return MatchesRules(domain, rules), ipv6
		}

		if cname != "" {
			log.Printf("[CNAME] %s ➜ %s", current, cname)
			cache.Set(current, cname) // Cache CNAME

			// Try A record on cname
			if ip2, err := doh.QueryA(cname); err == nil && ip2 != "" {
				log.Printf("[A+CNAME] %s ➜ %s", cname, ip2)
				cache.Set(cname, ip2)
				cache.Set(domain, ip2)
				return MatchesRules(domain, rules), ip2
			}

			// Try AAAA record on cname
			if ip6, err := doh.QueryAAAA(cname); err == nil && ip6 != "" {
				log.Printf("[AAAA+CNAME] %s ➜ %s", cname, ip6)
				cache.Set(cname, ip6)
				cache.Set(domain, ip6)
				return MatchesRules(domain, rules), ip6
			}

			current = cname
			continue
		}

		allRecords, err := doh.QueryAll(current)
		if err == nil && len(allRecords) > 0 {
			for _, recordList := range allRecords {
				for _, data := range recordList {
					log.Printf("[DNS] %s ➜ %s", current, data)
					cache.Set(current, data)
					cache.Set(domain, data)
					return MatchesRules(domain, rules), data
				}
			}
		}

		break
	}

	log.Printf("❌ Resolution failed for %s", domain)
	return false, ""
}

// LoadDomainRules loads DOMAIN-SUFFIX rules from a file
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
			suffix := strings.TrimPrefix(line, "DOMAIN-SUFFIX,")
			rules = append(rules, Rule{Suffix: suffix})
		}
	}

	return rules, nil

}

// ResolveWithCNAME exposes recursive resolution and returns CNAME (if any)
func ResolveWithCNAME(domain string, rules []Rule, cache *Cache) (bool, string, string) {
	visited := make(map[string]bool)
	current := domain
	var firstCNAME string

	for depth := 0; depth < 10; depth++ {
		if visited[current] {
			log.Printf("⚠️ Circular CNAME detected for %s", domain)
			return false, "", ""
		}
		visited[current] = true

		if cachedIP, ok := cache.Get(current); ok {
			log.Printf("[CACHE] %s ➜ %s", current, cachedIP)
			return MatchesRules(domain, rules), cachedIP, firstCNAME
		}

		ip, cname, err := doh.QueryWithCNAME(current)
		if err == nil && ip != "" {
			log.Printf("[A] %s ➜ %s", current, ip)
			cache.Set(current, ip)
			cache.Set(domain, ip)
			return MatchesRules(domain, rules), ip, firstCNAME
		}

		ipv6, err := doh.QueryAAAA(current)
		if err == nil && ipv6 != "" {
			log.Printf("[AAAA] %s ➜ %s", current, ipv6)
			cache.Set(current, ipv6)
			cache.Set(domain, ipv6)
			return MatchesRules(domain, rules), ipv6, firstCNAME
		}

		if cname != "" {
			log.Printf("[CNAME] %s ➜ %s", current, cname)
			cache.Set(current, cname) // Cache CNAME
			if firstCNAME == "" {
				firstCNAME = cname
			}

			// Try A record on cname
			if ip2, err := doh.QueryA(cname); err == nil && ip2 != "" {
				log.Printf("[A+CNAME] %s ➜ %s", cname, ip2)
				cache.Set(cname, ip2)
				cache.Set(domain, ip2)
				return MatchesRules(domain, rules), ip2, firstCNAME
			}

			// Try AAAA record on cname
			if ip6, err := doh.QueryAAAA(cname); err == nil && ip6 != "" {
				log.Printf("[AAAA+CNAME] %s ➜ %s", cname, ip6)
				cache.Set(cname, ip6)
				cache.Set(domain, ip6)
				return MatchesRules(domain, rules), ip6, firstCNAME
			}

			current = cname
			continue
		}

		allRecords, err := doh.QueryAll(current)
		if err == nil && len(allRecords) > 0 {
			for _, recordList := range allRecords {
				for _, data := range recordList {
					log.Printf("[DNS] %s ➜ %s", current, data)
					cache.Set(current, data)
					cache.Set(domain, data)
					return MatchesRules(domain, rules), data, firstCNAME
				}
			}
		}

		break
	}

	log.Printf("❌ Resolution failed for %s", domain)
	return false, "", ""
}
