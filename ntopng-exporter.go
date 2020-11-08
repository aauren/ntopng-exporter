package main

import (
	"fmt"
	"github.com/aauren/ntopng-exporter/internal/config"
	"os"
)

func main() {
	config, err := config.ParseConfig()
	if err != nil {
		fmt.Errorf("ran into the following error while attempting to parse config: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Config: %s", config)
}
