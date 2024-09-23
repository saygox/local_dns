package main

import (
	"log"
	"os"
	"regexp"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	DNSPort       int
	HTTPPort      int
	IsDebug       bool
	FallbackIP    string
	LocalhostOnly bool
	UseFallback   bool
    // UseLocakSocket bool
    // LocalSocketFile string
}

var (
    domainsToAddresses map[string]string = map[string]string{
    }
    mu sync.RWMutex
)

var config Config
var ipPortRegex = regexp.MustCompile(`^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(:\d{1,5})?$`)


var rootCmd = &cobra.Command{
	Use:   "localDNS",
	Short: "easy to use local DNS service",
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("help").Changed {
			// Display help and exit
			cmd.Help()
			os.Exit(0)
		}

		// ENV settings is not set
		if viper.IsSet("dns_port"){
			config.DNSPort = viper.GetInt("dns_port")
		}
		if viper.IsSet("http_port"){
			config.HTTPPort = viper.GetInt("http_port")
		}
		if viper.IsSet("debug"){
			config.IsDebug = viper.GetBool("debug")
		}
		if viper.IsSet("localhost_only"){
			config.LocalhostOnly = viper.GetBool("localhost_only")
		}
		if viper.IsSet("fallback_ip"){
			config.FallbackIP = viper.GetString("fallback_ip")
		}

		config.UseFallback = false
		if config.FallbackIP != "" {
			matches := ipPortRegex.FindStringSubmatch(config.FallbackIP)
			if matches != nil {
				config.UseFallback = true
				if matches[2] == "" {
					config.FallbackIP = matches[1] + ":53"
				}
			}
		}

		log.Println("Start local_dns: ",version)
		if config.IsDebug {
			// Processing when the command is executed
			log.Println("DNS Port:", config.DNSPort)
			if config.UseFallback {
				log.Println("Fallback IP:", config.FallbackIP)
			}
			log.Println("HTTP Port:", config.HTTPPort)
			log.Println("Localhost Only:", config.LocalhostOnly)
			log.Println("Is Debug:", config.IsDebug)
		}

		http_handleRequests()
		dns_handleRequests()
		select {}
	},
}


func init(){

	rootCmd.Version = version
	rootCmd.SetVersionTemplate("local_dns: {{.Version}}\n")

	// Initialize flags (map to struct fields)
	rootCmd.PersistentFlags().IntVar(&config.DNSPort, "dns-port", 2053, "DNS port")
	rootCmd.PersistentFlags().IntVar(&config.HTTPPort, "http-port", 2080, "HTTP port")
	rootCmd.PersistentFlags().BoolVar(&config.IsDebug, "debug", false, "Enable debug logs")
	rootCmd.PersistentFlags().BoolVar(&config.LocalhostOnly, "localhost-only", true, "Limit HTTP port to localhost")
	rootCmd.PersistentFlags().StringVar(&config.FallbackIP, "fallback-ip", "", "DNS fallback IP")

	// Synchronize environment variables and flags with Viper
	viper.BindPFlag("dns_port", rootCmd.PersistentFlags().Lookup("dns-port"))
	viper.BindPFlag("http_port", rootCmd.PersistentFlags().Lookup("http-port"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("localhost_only", rootCmd.PersistentFlags().Lookup("localhost-only"))
	viper.BindPFlag("fallback_ip", rootCmd.PersistentFlags().Lookup("fallback-ip"))

	// Set the prefix for environment variables (e.g., LOCALDNS_DNS_PORT, LOCALDNS_FALLBACK_IP, etc.)
	viper.SetEnvPrefix("LOCALDNS")
	viper.BindEnv("DNS_PORT")
	viper.BindEnv("HTTP_PORT")
	viper.BindEnv("DEBUG")
	viper.BindEnv("LOCALHOST_ONLY")
	viper.BindEnv("FALLBACK_IP")
	viper.AutomaticEnv()


	// Read in environment variables and set config values
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
}


func main() {
	if err := rootCmd.Execute(); err != nil {
        log.Fatalf("%v\n", err)
	}
}
