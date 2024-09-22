package main

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func RunLocalServer(pc net.PacketConn, l net.Listener, opts ...func(*dns.Server)) (*dns.Server, string, chan error, error) {
	server := &dns.Server{
		PacketConn: pc,
		Listener:   l,

		ReadTimeout:  time.Hour,
		WriteTimeout: time.Hour,
	}

	waitLock := sync.Mutex{}
	waitLock.Lock()
	server.NotifyStartedFunc = waitLock.Unlock

	for _, opt := range opts {
		opt(server)
	}

	var (
		addr   string
		closer io.Closer
	)
	if l != nil {
		addr = l.Addr().String()
		closer = l
	} else {
		addr = pc.LocalAddr().String()
		closer = pc
	}

	// fin must be buffered so the goroutine below won't block
	// forever if fin is never read from. This always happens
	// if the channel is discarded and can happen in TestShutdownUDP.
	fin := make(chan error, 1)

	go func() {
		fin <- server.ActivateAndServe()
		closer.Close()
	}()

	waitLock.Lock()
	return server, addr, fin, nil
}

func RunLocalUDPServer(laddr string, opts ...func(*dns.Server)) (*dns.Server, string, chan error, error) {
	pc, err := net.ListenPacket("udp", laddr)
	if err != nil {
		return nil, "", nil, err
	}

	return RunLocalServer(pc, nil, opts...)
}

func TestLocalDNSHandler(t *testing.T) {
    // Add a test domain to the map
    mu.Lock()
    domainsToAddresses["example.org"] = "192.168.1.2"
    mu.Unlock()

    tests := []struct {
        name     string
        domain   string
        qtype    uint16
        expected string
    }{
        {
            name:     "Valid A record",
            domain:   "example.org.",
            qtype:    dns.TypeA,
            expected: "192.168.1.2",
        },
        {
            name:     "Non-existent domain",
            domain:   "notfound.",
            qtype:    dns.TypeA,
            expected: "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            dns.HandleFunc(".", localDNS_handler)
            defer dns.HandleRemove(".")
            server, addrstr, _, err := RunLocalUDPServer(":0")
            if err != nil {
                t.Fatalf("unable to run test server: %v", err)
            }
            defer server.Shutdown()

            client := new(dns.Client)
            message := new(dns.Msg)
            message.SetQuestion(tt.domain, tt.qtype)
            response, _, err := client.Exchange(message, addrstr)
            if err != nil {
                t.Fatalf("Failed to exchange DNS message: %v", err)
            }

            if tt.expected == "" {
                if len(response.Answer) != 0 {
                    t.Errorf("Expected no answers, but got %v", response.Answer)
                }
            } else {
                if len(response.Answer) == 0 {
                    t.Fatalf("Expected an answer, but got none")
                }
                aRecord, ok := response.Answer[0].(*dns.A)
                if !ok {
                    t.Fatalf("Expected A record, but got %T", response.Answer[0])
                }
                if aRecord.A.String() != tt.expected {
                    t.Errorf("Expected IP %v, but got %v", tt.expected, aRecord.A.String())
                }
            }
        })
    }
}