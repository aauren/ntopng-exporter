package main

import (
	"context"
	"fmt"
	"github.com/aauren/ntopng-exporter/internal/config"
	"github.com/aauren/ntopng-exporter/internal/metrics"
	"github.com/aauren/ntopng-exporter/internal/ntopng"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Parse and validate the config
	myConfig, err := config.ParseConfig()
	if err != nil {
		fmt.Printf("ran into the following error while attempting to parse config: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Config: %s\n\n", myConfig)

	// Setup channel for stopping work when done
	stopChan := make(chan struct{})

	// Setup ntopng scrape controller and prime cache, then start it running asynchronously
	ntopControl := ntopng.CreateController(&myConfig, stopChan)
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
	go ntopControl.RunController()

	// Setup goroutine for serving traffic
	srv := serveMetrics(&ntopControl, &myConfig)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Printf("\n\nDetected shutdown - Cleaning Up Now\n\n")
	close(stopChan)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Was unable to gracefully shutdown prometheus http server: %v\n", err)
	}
	fmt.Printf("\nGoodbye")
}

func serveMetrics(ntopController *ntopng.Controller, config *config.Config) *http.Server {
	ntopCollector := metrics.NewNtopNGCollector(ntopController, config)
	prometheus.MustRegister(ntopCollector)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server {
		Addr: fmt.Sprintf("%s:%d", config.Metric.Serve.IP, config.Metric.Serve.Port),
		Handler: mux,
	}
	go func(srv *http.Server) {
		if msg := srv.ListenAndServe(); msg != nil {
			fmt.Printf("Output from HTTP Server: %v", msg)
		}
	}(srv)
	return srv
}