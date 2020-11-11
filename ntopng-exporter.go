package main

import (
	"fmt"
	"github.com/aauren/ntopng-exporter/internal/config"
	"github.com/aauren/ntopng-exporter/internal/metrics"
	"github.com/aauren/ntopng-exporter/internal/ntopng"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
)

func main() {
	myConfig, err := config.ParseConfig()
	if err != nil {
		fmt.Printf("ran into the following error while attempting to parse config: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Config: %s\n\n", myConfig)
	ntopControl := ntopng.CreateController(&myConfig)
	err = ntopControl.CacheInterfaceIds()
	if err != nil {
		fmt.Printf("failed to cache interface ids: %v\n", err)
		os.Exit(2)
	}
	err = ntopControl.ScrapeHostEndpointForAllInterfaces()
	if err != nil {
		fmt.Printf("failed to scrape host endpoint: %v\n", err)
		os.Exit(3)
	}
	err = serveMetrics(&ntopControl, &myConfig)
	if err != nil {
		fmt.Printf("error while listening on metric port: %v\n", err)
		os.Exit(4)
	}
}

func serveMetrics(ntopController *ntopng.Controller, config *config.Config) error {
	ntopCollector := metrics.NewNtopNGCollector(ntopController, config)
	prometheus.MustRegister(ntopCollector)
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":3001", nil)
}