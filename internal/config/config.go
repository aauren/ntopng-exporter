package config

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	AllScrape              = "all"
	HostScrape             = "hosts"
	InterfaceScrape        = "interfaces"
	L7Protocols            = "l7protocols"
	DefaultMetricServePort = 3001
	DefaultRequestTimeout  = 30 * time.Second
	DefaultParallelWorkers = 1
)

var (
	AvailableScrapeTargets = map[string]bool{
		AllScrape:       true,
		HostScrape:      true,
		InterfaceScrape: true,
		L7Protocols:     true}
)

type ntopng struct {
	EndPoint        string
	User            string
	Password        string
	Token           string
	AuthMethod      string
	ScrapeInterval  string
	RequestTimeout  string
	ScrapeTargets   []string
	AllowUnsafeTLS  bool
	ParallelWorkers int
}

type host struct {
	InterfacesToMonitor []string
}

type metric struct {
	LocalSubnetsOnly  []string
	ExcludeDNSMetrics bool
	Serve             metricServe
}

type metricServe struct {
	IP   string
	Port int
}

type Config struct {
	Ntopng ntopng
	Host   host
	Metric metric
}

func ParseConfig() (Config, error) {
	// Configure paths and read config
	var config Config
	viper.SetConfigName("ntopng-exporter")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.ntopng-exporter")
	viper.AddConfigPath("/etc/ntopng-exporter/")
	viper.AddConfigPath("./config")

	err := viper.ReadInConfig()
	if err != nil {
		return config, err
	}

	// Set default values
	viper.SetDefault("metric.excludeDNSMetrics", false)
	viper.SetDefault("ntopng.scrapeInterval", "1m")
	// We have to register the scrapeInterval default before GetDuration so that it can fall back to it, an
	// unparsable user value comes back as 0 here and gets rejected properly by validate() later on
	viper.SetDefault("ntopng.requestTimeout", defaultRequestTimeout(viper.GetDuration("ntopng.scrapeInterval")).String())
	viper.SetDefault("ntopng.metric.serve.ip", "0.0.0.0")
	viper.SetDefault("ntopng.metric.serve.port", DefaultMetricServePort)
	viper.SetDefault("ntopng.scrapeTargets", "all")
	viper.SetDefault("ntopng.allowUnsafeTLS", false)
	viper.SetDefault("ntopng.parallelWorkers", DefaultParallelWorkers)

	// Unmarshal config into struct
	err = viper.Unmarshal(&config)
	if err != nil {
		return config, err
	}
	// Check if environment variable NTOPNG_TOKEN is set
	if tokenEnv, exists := os.LookupEnv("NTOPNG_TOKEN"); exists {
		config.Ntopng.Token = tokenEnv
	}
	err = config.validate()
	return config, err
}

// defaultRequestTimeout caps the default at the scrape interval so that short-interval configs don't fail
// validation out of the box, ignoring non-positive intervals which validate() rejects with a better message
func defaultRequestTimeout(scrapeInterval time.Duration) time.Duration {
	if scrapeInterval > 0 {
		return min(DefaultRequestTimeout, scrapeInterval)
	}
	return DefaultRequestTimeout
}

func (c *Config) validate() error {
	if c.Ntopng.AuthMethod != "cookie" && c.Ntopng.AuthMethod != "basic" && c.Ntopng.AuthMethod != "token" && c.Ntopng.AuthMethod != "none" {
		return fmt.Errorf("ntopng authMethod must be either cookie, basic, token or none")
	}
	if c.Ntopng.AuthMethod == "cookie" || c.Ntopng.AuthMethod == "basic" {
		if c.Ntopng.User == "" || c.Ntopng.Password == "" {
			return fmt.Errorf("ntopng user and password must be set when using cookie or basic auth")
		}
	}
	if c.Ntopng.AuthMethod == "token" {
		if c.Ntopng.Token == "" {
			return fmt.Errorf("ntopng token must be set when using token auth")
		}
	}
	if len(c.Host.InterfacesToMonitor) < 1 {
		return fmt.Errorf("must specify at least one interface to monitor")
	}
	for _, ifName := range c.Host.InterfacesToMonitor {
		if ifName == "" {
			return fmt.Errorf("interface name cannot be null or blank")
		}
	}
	if len(c.Metric.LocalSubnetsOnly) > 0 {
		for _, subnet := range c.Metric.LocalSubnetsOnly {
			if _, _, err := net.ParseCIDR(subnet); err != nil {
				return fmt.Errorf("subnet specified: '%s', is not a valid subnet: %v", subnet, err)
			}
		}
	}
	scrapeInterval, err := time.ParseDuration(c.Ntopng.ScrapeInterval)
	if err != nil {
		return fmt.Errorf("was not able to parse configured duration: %s - %v", c.Ntopng.ScrapeInterval, err)
	}
	if scrapeInterval <= 0 {
		return fmt.Errorf("ntopng scrapeInterval must be greater than zero")
	}
	requestTimeout, err := time.ParseDuration(c.Ntopng.RequestTimeout)
	if err != nil {
		return fmt.Errorf("was not able to parse configured request timeout: %s - %v", c.Ntopng.RequestTimeout, err)
	}
	if requestTimeout <= 0 {
		return fmt.Errorf("ntopng requestTimeout must be greater than zero")
	}
	if requestTimeout > scrapeInterval {
		// We clamp rather than error so that a config tuned for a slow ntopng keeps working
		// when someone later shortens the scrape interval
		fmt.Printf("ntopng requestTimeout (%s) is greater than scrapeInterval (%s), clamping requestTimeout to %s\n",
			c.Ntopng.RequestTimeout, c.Ntopng.ScrapeInterval, scrapeInterval)
		c.Ntopng.RequestTimeout = scrapeInterval.String()
	}
	if c.Ntopng.ParallelWorkers < 1 {
		return fmt.Errorf("ntopng parallelWorkers must be at least 1, got: %d", c.Ntopng.ParallelWorkers)
	}
	if c.Metric.Serve.IP != "0.0.0.0" {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return fmt.Errorf("was not able to get list of interface addresses: %v", err)
		}
		foundIP := false
		for _, addr := range addrs {
			if strings.Contains(addr.String(), c.Metric.Serve.IP) {
				foundIP = true
			}
		}
		if !foundIP {
			return fmt.Errorf("it looks like address isn't present on the host to bind to: %s", c.Metric.Serve.IP)
		}
	}
	if len(c.Ntopng.ScrapeTargets) < 1 {
		return fmt.Errorf("you must specify at least one scrape target in the config")
	}
	for _, target := range c.Ntopng.ScrapeTargets {
		if !AvailableScrapeTargets[target] {
			return fmt.Errorf("'%s' is not an available scrape target: %v",
				target, AvailableScrapeTargets)
		}
	}
	return nil
}

func (c Config) String() string {
	configOutput := fmt.Sprintf("ntopng:\n%s\n\nhost:\n%s\n\nmetric:\n%s", c.Ntopng, c.Host, c.Metric)
	return configOutput
}

func (n ntopng) String() string {
	return fmt.Sprintf(
		"\t%s: '%s'/*HIDDEN* - %s - Allow Unsafe TLS? %t\n\tScrape Interval: %s\n\tRequest Timeout: %s"+
			"\n\tScrape Targets: %s\n\tParallel Workers: %d",
		n.EndPoint, n.User, n.AuthMethod, n.AllowUnsafeTLS, n.ScrapeInterval, n.RequestTimeout,
		n.ScrapeTargets, n.ParallelWorkers)
}

func (h host) String() string {
	return fmt.Sprintf("\tInterface List: %v", h.InterfacesToMonitor)
}

func (m metric) String() string {
	return fmt.Sprintf("\tLocal Subnets: %v\n\tExclude DNS Metrics? %t\n\tServe:\n%s",
		m.LocalSubnetsOnly, m.ExcludeDNSMetrics, m.Serve)
}

func (ms metricServe) String() string {
	return fmt.Sprintf("\t\tIP: %s\n\t\tPort: %d", ms.IP, ms.Port)
}
