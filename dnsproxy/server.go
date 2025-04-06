package dnsproxy

import (
	"log"
	"net"
	"openvpnadvanced/dnsmasq"
	"openvpnadvanced/utils"
	"openvpnadvanced/vpn"
	"strings"

	"github.com/miekg/dns"
)

type DNSServer struct {
	Rules    []dnsmasq.Rule
	Cache    *dnsmasq.Cache
	Fallback string
	VPNIface string
}

func NewServer(rules []dnsmasq.Rule, cache *dnsmasq.Cache, fallback string, vpnIface string) *DNSServer {
	return &DNSServer{
		Rules:    rules,
		Cache:    cache,
		Fallback: "127.0.0.1:53",
		VPNIface: vpnIface,
	}
}

func (s *DNSServer) Start() {
	handler := dns.NewServeMux()
	handler.HandleFunc(".", s.handleDNSRequest)

	go func() {
		server := &dns.Server{Addr: ":53", Net: "udp", Handler: handler}
		log.Println("🌀 DNS server (UDP) listening on :53")
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start UDP DNS server: %v", err)
		}
	}()

	go func() {
		server := &dns.Server{Addr: ":53", Net: "tcp", Handler: handler}
		log.Println("🌀 DNS server (TCP) listening on :53")
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start TCP DNS server: %v", err)
		}
	}()
}

func (s *DNSServer) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)

	if len(r.Question) == 0 {
		_ = w.WriteMsg(msg)
		return
	}

	q := r.Question[0]
	domain := strings.TrimSuffix(q.Name, ".")

	switch q.Qtype {
	case dns.TypeA:
		// Handle A record normally
	case dns.TypeAAAA, dns.TypeHTTPS, dns.TypeSVCB, dns.TypePTR, dns.TypeSOA:
		msg.Answer = []dns.RR{}
		_ = w.WriteMsg(msg)
		return
	default:
		log.Printf("⚠️ Unsupported query type: %d for %s", q.Qtype, domain)
		msg.Answer = []dns.RR{}
		_ = w.WriteMsg(msg)
		return
	}

	// 使用递归解析逻辑（带缓存）
	shouldRoute, ip := dnsmasq.ResolveRecursive(domain, s.Rules, s.Cache)

	log.Printf("🔍 Domain: %s | IP: %s | VPN: %v", domain, ip, shouldRoute)

	if ip == "" {
		utils.PrintError(domain, "failed to resolve")
		_ = w.WriteMsg(msg)
		return
	}

	msg.Answer = append(msg.Answer, makeARecord(domain, ip))
	_ = w.WriteMsg(msg)

	printDNSLog(domain, ip, shouldRoute)

	// 添加静态路由（确保 VPN 拦截）
	if shouldRoute {
		if err := vpn.AddRoute(ip, s.VPNIface); err != nil {
			log.Printf("⚠️ Failed to add route for %s ➜ %s: %v", ip, s.VPNIface, err)
		} else {
			log.Printf("✅ Route added: %s ➜ %s", ip, s.VPNIface)
		}
	}
}

// nolint: all
func (s *DNSServer) forwardToFallback(domain string) (string, error) {
	client := new(dns.Client)
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	resp, _, err := client.Exchange(msg, s.Fallback)
	if err != nil {
		return "", err
	}

	for _, ans := range resp.Answer {
		if a, ok := ans.(*dns.A); ok {
			return a.A.String(), nil
		}
	}
	return "", nil
}

func makeARecord(domain, ip string) dns.RR {
	return &dns.A{
		Hdr: dns.RR_Header{
			Name:   dns.Fqdn(domain),
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    300,
		},
		A: net.ParseIP(ip),
	}
}

func printDNSLog(domain, ip string, vpn bool) {
	if ip == "" {
		utils.PrintError(domain, "no A record")
	} else if vpn {
		utils.PrintVPN(domain, ip)
	} else {
		utils.PrintDirect(domain, ip)
	}
}
