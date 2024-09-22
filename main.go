package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"sync"
)

var defaultDNSName string= "dns.example.com"
var defaultDNSPort int= 2053
var defaultHTTPPort int= 2080

var (
    domainsToAddresses map[string]string = map[string]string{
    }
    mu sync.RWMutex
)


func set_localaddrss(dnsName string) {

    // Get the machine's IP addresses
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        log.Fatalf("Failed to get IP addresses: %v", err)
    }

    // Find a non-loopback IP address
    var machineIP string
    for _, addr := range addrs {
        // Skip loopback addresses
        if addr.String() == "127.0.0.1/8" {
            continue
        }
        if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
            if ipNet.IP.To4() != nil {
                machineIP = ipNet.IP.String()
                break
            }
        }
    }

    if machineIP == "" {
        log.Fatalf("No non-loopback IP address found")
    }

    // Add the machine's IP address to the domainsToAddresses map
    mu.Lock()
    domainsToAddresses[dnsName] = machineIP
    mu.Unlock()

    // Log the current domain to address mappings
    mu.RLock()
    for domain, address := range domainsToAddresses {
        log.Printf("Domain: %s, Address: %s\n", domain, address)
    }
    mu.RUnlock()
}

func getConfig() (string, int, int, bool) {
    var dnsName string = defaultDNSName
    var dnsPort int = defaultDNSPort
    var httpPort int = defaultHTTPPort
    var debug bool = false

    // Check environment variables first
    if dnsNameEnv := os.Getenv("DNS_NAME"); dnsNameEnv != "" {
        dnsName = dnsNameEnv
        if dnsName[len(dnsName)-1] != '.' {
            dnsName += "."
        }
    }

    if dnsPortEnv := os.Getenv("DNS_PORT"); dnsPortEnv != "" {
        if port, err := strconv.Atoi(dnsPortEnv); err == nil {
            dnsPort = port
        }
    }

    if httpPortEnv := os.Getenv("HTTP_PORT"); httpPortEnv != "" {
        if port, err := strconv.Atoi(httpPortEnv); err == nil {
            httpPort = port
        }
    }

    if debugEnv := os.Getenv("DEBUG"); debugEnv != "" {
        if debugEnv == "true" {
            debug = true
        }
    }

    // Override with command line arguments if provided
    for i := 1; i < len(os.Args); i++ {
        switch os.Args[i] {
        case "--name":
            if i+1 < len(os.Args) {
                dnsName = os.Args[i+1]
                if dnsName[len(dnsName)-1] != '.' {
                    dnsName += "."
                }
                i++
            } else {
                log.Fatalf("Missing value for --name")
            }
        case "--dns-port":
            if i+1 < len(os.Args) {
                if port, err := strconv.Atoi(os.Args[i+1]); err == nil {
                    dnsPort = port
                    i++
                } else {
                    log.Fatalf("Invalid value for --dns-port: %v", err)
                }
            } else {
                log.Fatalf("Missing value for --dns-port")
            }
        case "--http-port":
            if i+1 < len(os.Args) {
                if port, err := strconv.Atoi(os.Args[i+1]); err == nil {
                    httpPort = port
                    i++
                } else {
                    log.Fatalf("Invalid value for --http-port: %v", err)
                }
            } else {
                log.Fatalf("Missing value for --http-port")
            }
        case "--debug":
            debug = true
        }
    }
    return dnsName, dnsPort, httpPort, debug
}

func main() {
    dnsName, dnsPort, httpPort, _ := getConfig()

    set_localaddrss(dnsName)
    http_handleRequests(httpPort)
    dns_handleRequests(dnsPort)
    select {}
}
