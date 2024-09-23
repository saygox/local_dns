package main

import (
	"log"
	"net"
	"strconv"

	"github.com/miekg/dns"
)

// ServeDNS handles DNS requests and provides responses based on the query type.
// It currently supports handling A records. The function reads the domain name
// from the query, checks if it exists in the domainsToAddresses map, and if so,
// constructs a DNS response with the corresponding IP address.
//
// Parameters:
// - w: dns.ResponseWriter to write the DNS response.
// - r: *dns.Msg containing the DNS request.
//
// The function uses a read lock to ensure thread-safe access to the domainsToAddresses map.
func localDNS_handler(w dns.ResponseWriter, r *dns.Msg) {
    var isLoopback bool
    isLoopback = false
    msg := dns.Msg{}
    msg.SetReply(r)
    switch r.Question[0].Qtype {
    case dns.TypeA:
        mu.RLock() // Lock for reading
        defer mu.RUnlock()
        msg.Authoritative = true
        domain := msg.Question[0].Name
        var search_domain string = domain[:len(domain)-1]
        address, ok := domainsToAddresses[search_domain]
        if ok {
            msg.Answer = append(msg.Answer, &dns.A{
                Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
                A:   net.ParseIP(address),
            })
            isLoopback = true
        }
    }
    if isLoopback {
        w.WriteMsg(&msg)
        return
    } else if !config.UseFallback{
        dns.HandleFailed(w, r)
        return
    }

    if config.IsDebug {
        log.Printf("Domain not found: %s\n", r.Question[0].Name)
    }
    // If the domain is not found, return a SERVFAIL response
    // Forward the request to an external DNS server (e.g., 8.8.8.8:53) if the domain is not found
    client := new(dns.Client)
    message := new(dns.Msg)
    message.SetQuestion(r.Question[0].Name, r.Question[0].Qtype)
    response, _, err := client.Exchange(message, config.FallbackIP)
    if err != nil {
        log.Printf("Failed to forward request to %s: %v", config.FallbackIP, err)
        dns.HandleFailed(w, r)
        return
    }
    if response.Rcode != dns.RcodeSuccess {
        log.Printf("DNS query failed with Rcode: %d", response.Rcode)
        dns.HandleFailed(w, r)
        return
    }
    w.WriteMsg(response.SetReply(r))
}

// dns_handleRequests starts a DNS server on the specified port.
// The server listens for UDP requests and uses a custom handler to process them.
func dns_handleRequests() {
    go func() {
        dns.HandleFunc(".", localDNS_handler)
        defer dns.HandleRemove(".")

        srv := &dns.Server{Addr: ":" + strconv.Itoa(config.DNSPort), Net: "udp"}
        if err := srv.ListenAndServe(); err != nil {
            log.Fatalf("Failed to set udp listener %s\n", err.Error())
        }
    }()
}
