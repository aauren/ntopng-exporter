package main

import (
	"fmt"
	"github.com/aauren/ntopng-exporter/internal/config"
	"github.com/aauren/ntopng-exporter/internal/ntopng"
	"os"
)

func main() {
	config, err := config.ParseConfig()
	if err != nil {
		fmt.Printf("ran into the following error while attempting to parse config: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Config: %s\n\n", config)
	ntopControl := ntopng.CreateController(config)
	err = ntopControl.ScrapeHostEndpoint(1)
	if err != nil {
		fmt.Printf("failed to scrape host endpoint: %v", err)
		os.Exit(2)
	}
}