package config

import (
	"fmt"
	"github.com/spf13/viper"
	"net"
	"strings"
	"time"
)

const (
	AllScrape       = "all"
	HostScrape      = "hosts"
	InterfaceScrape = "interfaces"
	L7Protocols     = "l7protocols"
)

var (
	AvailableScrapeTargets = map[string]bool{
		AllScrape:       true,
		HostScrape:      true,
		InterfaceScrape: true,
		L7Protocols:     true}
)

type ntopng struct {
	EndPoint       string
	User           string
	Password       string
	AuthMethod     string
	ScrapeInterval string
	ScrapeTargets  []string
	AllowUnsafeTLS bool
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
	viper.SetDefault("ntopng.metric.serve.ip", "0.0.0.0")
	viper.SetDefault("ntopng.metric.serve.port", 3001)
	viper.SetDefault("ntopng.scrapeTargets", "all")
	viper.SetDefault("ntopng.allowUnsafeTLS", false)

	// Unmarshal config into struct
	err = viper.Unmarshal(&config)
	if err != nil {
		return config, err
	}
	err = config.validate()
	return config, err
}

func (c *Config) validate() error {
	if c.Ntopng.AuthMethod != "cookie" && c.Ntopng.AuthMethod != "basic" && c.Ntopng.AuthMethod != "none" {
		return fmt.Errorf("ntopng authMethod must be either cookie, basic, or none")
	}
	if c.Host.InterfacesToMonitor == nil || len(c.Host.InterfacesToMonitor) < 1 {
		return fmt.Errorf("must specify at least one interface to monitor")
	}
	for _, ifName := range c.Host.InterfacesToMonitor {
		if ifName == "" {
			return fmt.Errorf("interface name cannot be null or blank")
		}
	}
	if c.Metric.LocalSubnetsOnly != nil && len(c.Metric.LocalSubnetsOnly) > 0 {
		for _, subnet := range c.Metric.LocalSubnetsOnly {
			if _, _, err := net.ParseCIDR(subnet); err != nil {
				return fmt.Errorf("subnet specified: '%s', is not a valid subnet: %v", subnet, err)
			}
		}
	}
	if _, err := time.ParseDuration(c.Ntopng.ScrapeInterval); err != nil {
		return fmt.Errorf("was not able to parse configured duration: %s - %v", c.Ntopng.ScrapeInterval, err)
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
	return fmt.Sprintf("\t%s: '%s'/'%s' - %s - Allow Unsafe TLS? %t\n\tScrape Interval: %s\n\tScrape Targets: %s",
		n.EndPoint, n.User, n.Password, n.AuthMethod, n.AllowUnsafeTLS, n.ScrapeInterval, n.ScrapeTargets)
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
