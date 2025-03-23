package doh

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type DoHAnswer struct {
	Name string `json:"name"`
	Type int    `json:"type"`
	TTL  int    `json:"TTL"`
	Data string `json:"data"`
}

type DoHResponse struct {
	Answer []DoHAnswer `json:"Answer"`
}

// DNS record types (https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml)
const (
	TypeA     = 1
	TypeNS    = 2
	TypeCNAME = 5
	TypeSOA   = 6
	TypePTR   = 12
	TypeMX    = 15
	TypeTXT   = 16
	TypeAAAA  = 28
	TypeSRV   = 33
)

// Query returns the first A record (IPv4)
func Query(domain string) (string, error) {
	return querySingleType(domain, TypeA)
}

// QueryAAAA returns the first AAAA record (IPv6)
func QueryAAAA(domain string) (string, error) {
	return querySingleType(domain, TypeAAAA)
}

// QueryTXT returns the first TXT record
func QueryTXT(domain string) (string, error) {
	return querySingleType(domain, TypeTXT)
}

// QueryMX returns the first MX record
func QueryMX(domain string) (string, error) {
	return querySingleType(domain, TypeMX)
}

// QueryNS returns the first NS record
func QueryNS(domain string) (string, error) {
	return querySingleType(domain, TypeNS)
}

// QueryCNAME returns the first CNAME
func QueryCNAME(domain string) (string, error) {
	return querySingleType(domain, TypeCNAME)
}

// QueryAll returns all records of all known types for a domain
func QueryAll(domain string) (map[string][]string, error) {
	types := []int{TypeA, TypeAAAA, TypeCNAME, TypeMX, TypeTXT, TypeNS, TypeSOA, TypePTR, TypeSRV}
	results := make(map[string][]string)

	for _, t := range types {
		records, err := queryRaw(domain, t)
		if err == nil && len(records) > 0 {
			typeStr := dnsTypeToString(t)
			for _, rec := range records {
				results[typeStr] = append(results[typeStr], rec.Data)
			}
		}
	}

	return results, nil
}

// QueryWithCNAME returns IP or next CNAME if found (for routing fallback)
func QueryWithCNAME(domain string) (ip string, cname string, err error) {
	url := fmt.Sprintf("https://cloudflare-dns.com/dns-query?name=%s&type=A", domain)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/dns-json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var dohRes DoHResponse
	err = json.Unmarshal(body, &dohRes)
	if err != nil {
		return "", "", err
	}

	for _, answer := range dohRes.Answer {
		switch answer.Type {
		case TypeA:
			return answer.Data, "", nil
		case TypeCNAME:
			return "", strings.TrimSuffix(answer.Data, "."), nil
		}
	}

	return "", "", fmt.Errorf("no A record or CNAME found")
}

// querySingleType fetches the first answer of a given DNS type
func querySingleType(domain string, t int) (string, error) {
	records, err := queryRaw(domain, t)
	if err != nil || len(records) == 0 {
		return "", fmt.Errorf("no %s record found", dnsTypeToString(t))
	}
	return records[0].Data, nil
}

// queryRaw returns all answers of the specified type
func queryRaw(domain string, t int) ([]DoHAnswer, error) {
	url := fmt.Sprintf("https://cloudflare-dns.com/dns-query?name=%s&type=%d", domain, t)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/dns-json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var dohRes DoHResponse
	err = json.Unmarshal(body, &dohRes)
	if err != nil {
		return nil, err
	}

	return dohRes.Answer, nil
}

// dnsTypeToString maps DNS type code to human-readable name
func dnsTypeToString(t int) string {
	switch t {
	case TypeA:
		return "A"
	case TypeAAAA:
		return "AAAA"
	case TypeCNAME:
		return "CNAME"
	case TypeMX:
		return "MX"
	case TypeTXT:
		return "TXT"
	case TypeNS:
		return "NS"
	case TypeSOA:
		return "SOA"
	case TypePTR:
		return "PTR"
	case TypeSRV:
		return "SRV"
	default:
		return fmt.Sprintf("TYPE%d", t)
	}
}
