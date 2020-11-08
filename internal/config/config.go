package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type ntopng struct {
	EndPoint string
	User     string
	Password string
	AuthMethod string
}

type config struct {
	Ntopng ntopng
}

func ParseConfig() (config, error) {
	var config config
	viper.SetConfigName("ntopng-exporter")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.ntopng-exporter")
	viper.AddConfigPath("/etc/ntopng-exporter/")
	viper.AddConfigPath("./config")

	err := viper.ReadInConfig()
	if err != nil {
		return config, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return config, err
	}
	//err = config.validate()
	return config, err
}

func (c *config) validate() error {
	if c.Ntopng.AuthMethod != "cookie" && c.Ntopng.AuthMethod != "basic" {
		return fmt.Errorf("ntopng authMethod must be either cookie or basic")
	}
	return nil
}

func (c config) String() string {
	configOutput := fmt.Sprintf("ntopng:\n\t%s", c.Ntopng)
	return configOutput
}

func (n ntopng) String() string {
	return fmt.Sprintf("%s: '%s'/'%s' - %s", n.EndPoint, n.User, n.Password, n.AuthMethod)
}
