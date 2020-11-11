package config

import (
	"fmt"
	"github.com/spf13/viper"
	"net"
)

type ntopng struct {
	EndPoint string
	User     string
	Password string
	AuthMethod string
}

type host struct {
	InterfacesToMonitor []string
}

type metric struct {
	LocalSubnetsOnly []string
	ExcludeDNSMetrics bool
}

type Config struct {
	Ntopng ntopng
	Host host
	Metric metric
}

func ParseConfig() (Config, error) {
	// Configure paths and read config
	var config Config
	viper.SetConfigName("ntopng-exporter")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.ntopng-exporter")
	viper.AddConfigPath("/etc/ntopng-exporter/")
	viper.AddConfigPath("./Config")

	err := viper.ReadInConfig()
	if err != nil {
		return config, err
	}
	
	// Set default values
	viper.SetDefault("metric.excludeDNSMetrics", false)

	// Unmarshal config into struct
	err = viper.Unmarshal(&config)
	if err != nil {
		return config, err
	}
	//err = Config.validate()
	return config, err
}

func (c *Config) validate() error {
	if c.Ntopng.AuthMethod != "cookie" && c.Ntopng.AuthMethod != "basic" && c.Ntopng.AuthMethod != "none" {
		return fmt.Errorf("ntopng authMethod must be either cookie, basic, or none")
	}
	if c.Host.InterfacesToMonitor != nil || len(c.Host.InterfacesToMonitor) < 1 {
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
	return nil
}

func (c Config) String() string {
	configOutput := fmt.Sprintf("ntopng:\n\t%s", c.Ntopng)
	return configOutput
}

func (n ntopng) String() string {
	return fmt.Sprintf("%s: '%s'/'%s' - %s", n.EndPoint, n.User, n.Password, n.AuthMethod)
}
