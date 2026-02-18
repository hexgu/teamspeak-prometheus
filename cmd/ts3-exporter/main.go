package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/hexgu/teamspeak-prometheus/internal/collector"
	"github.com/hexgu/teamspeak-prometheus/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Create collector
	ts3Collector := collector.NewTS3Collector(cfg)

	// Register collector
	prometheus.MustRegister(ts3Collector)

	// Start HTTP server
	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", cfg.MetricsPort)
	log.Printf("Starting exporter on %s", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
