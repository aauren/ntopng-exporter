package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aauren/ntopng-exporter/internal"
	"github.com/aauren/ntopng-exporter/internal/config"
	ntopPrometheus "github.com/aauren/ntopng-exporter/internal/metrics/prometheus"
	"github.com/aauren/ntopng-exporter/internal/ntopng"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Parse and validate the config
	myConfig, err := config.ParseConfig()
	if err != nil {
		fmt.Printf("ran into the following error while attempting to parse config: %v", err)
		os.Exit(1)
	}
	fmt.Printf("Config: %s\n\n", myConfig)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// ParseConfig has already validated this, so a failure here means a programming error on our side
	requestTimeout, err := time.ParseDuration(myConfig.Ntopng.RequestTimeout)
	if err != nil {
		fmt.Printf("failed to parse request timeout '%s': %v\n", myConfig.Ntopng.RequestTimeout, err)
		os.Exit(3)
	}

	// Setup ntopng scrape controller and prime cache, then start it running asynchronously
	ntopControl := ntopng.CreateController(ctx, &myConfig, requestTimeout)
	err = ntopControl.CacheInterfaceIds()
	if err != nil {
		fmt.Printf("failed to cache interface ids: %v\n", err)
		os.Exit(2)
	}
	ntopControl.ScrapeAllConfiguredTargets()
	go ntopControl.RunController()

	// Setup goroutine for serving traffic
	srv := serveMetrics(&ntopControl, &myConfig)
	<-ctx.Done()
	stop()

	fmt.Printf("\n\nDetected shutdown - Cleaning Up Now\n\n")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Was unable to gracefully shutdown prometheus http server: %v\n", err)
	}
	fmt.Printf("\nGoodbye")
}

func serveMetrics(ntopController *ntopng.Controller, myConfig *config.Config) *http.Server {
	// Request timing and error counts aren't tied to any one scrape target, so they're always on
	prometheus.MustRegister(ntopController.Collectors()...)
	if internal.IsItemInArray(myConfig.Ntopng.ScrapeTargets, config.HostScrape) ||
		internal.IsItemInArray(myConfig.Ntopng.ScrapeTargets, config.AllScrape) {
		ntopCollector := ntopPrometheus.NewNtopNGHostCollector(ntopController, myConfig)
		prometheus.MustRegister(ntopCollector)
	}
	if internal.IsItemInArray(myConfig.Ntopng.ScrapeTargets, config.InterfaceScrape) ||
		internal.IsItemInArray(myConfig.Ntopng.ScrapeTargets, config.AllScrape) {
		ntopCollector := ntopPrometheus.NewNtopNGInterfaceCollector(ntopController, myConfig)
		prometheus.MustRegister(ntopCollector)
	}
	if internal.IsItemInArray(myConfig.Ntopng.ScrapeTargets, config.L7Protocols) ||
		internal.IsItemInArray(myConfig.Ntopng.ScrapeTargets, config.AllScrape) {
		ntopCollector := ntopPrometheus.NewNtopNGL7Collector(ntopController, myConfig)
		prometheus.MustRegister(ntopCollector)
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", myConfig.Metric.Serve.IP, myConfig.Metric.Serve.Port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func(srv *http.Server) {
		if msg := srv.ListenAndServe(); msg != nil {
			fmt.Printf("Output from HTTP Server: %v", msg)
		}
	}(srv)
	return srv
}
