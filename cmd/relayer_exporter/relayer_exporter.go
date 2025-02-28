package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/archway-network/relayer_exporter/pkg/collector"
	"github.com/archway-network/relayer_exporter/pkg/config"
	log "github.com/archway-network/relayer_exporter/pkg/logger"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func getVersion() string {
	return fmt.Sprintf("version: %s commit: %s date: %s", version, commit, date)
}

// refreshCollectors updates the collectors with new configuration
func refreshCollectors(ctx context.Context, cfg *config.Config, registry *prometheus.Registry) error {
	paths, err := cfg.IBCPaths(ctx)
	if err != nil {
		return fmt.Errorf("failed to get IBC paths: %w", err)
	}

	rpcs, err := cfg.GetRPCsMap()
	if err != nil {
		return fmt.Errorf("failed to get RPCs map: %w", err)
	}

	// Unregister existing collectors
	registry.Unregister(collector.IBCCollector{})
	registry.Unregister(collector.WalletBalanceCollector{})

	// Create and register new collectors
	ibcCollector := collector.IBCCollector{
		RPCs:  rpcs,
		Paths: paths,
	}

	balancesCollector := collector.WalletBalanceCollector{
		RPCs:     rpcs,
		Accounts: cfg.Accounts,
	}

	registry.MustRegister(ibcCollector)
	registry.MustRegister(balancesCollector)

	return nil
}

func main() {
	port := flag.Int("p", 8008, "Server port")
	version := flag.Bool("version", false, "Print version")
	configPath := flag.String("config", "./config.yml", "path to config file")
	refreshInterval := flag.Duration("refresh", 5*time.Minute, "Configuration refresh interval")
	logLevel := log.LevelFlag()

	flag.Parse()
	log.SetLevel(*logLevel)

	if *version {
		fmt.Println(getVersion())
		os.Exit(0)
	}

	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info(
		fmt.Sprintf(
			"Github IBC registry: %s/%s",
			cfg.GitHub.Org,
			cfg.GitHub.Repo,
		),
		zap.String("Mainnet Directory", cfg.GitHub.IBCDir),
		zap.String("Testnet Directory", cfg.GitHub.TestnetsIBCDir),
	)

	// Create a context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure context is cancelled when main exits

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	registry := prometheus.NewRegistry()

	// Initial setup of collectors
	if err := refreshCollectors(ctx, cfg, registry); err != nil {
		log.Fatal(err.Error())
	}

	// Start periodic refresh in background
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(*refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Info("Stopping collector refresh routine")
				return
			case <-ticker.C:
				log.Info("Refreshing configuration and collectors")
				if err := refreshCollectors(ctx, cfg, registry); err != nil {
					log.Error(fmt.Sprintf("Failed to refresh collectors: %v", err))
					continue
				}
				log.Info("Successfully refreshed configuration and collectors")
			}
		}
	}()

	// Setup HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: nil,
	}

	// Setup HTTP handler with custom registry
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	http.Handle("/metrics", handler)

	// Start server in a goroutine
	go func() {
		log.Info(fmt.Sprintf("Starting server on addr: %s", server.Addr))
		log.Info(fmt.Sprintf("Configuration refresh interval: %s", refreshInterval.String()))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(fmt.Sprintf("Server error: %v", err))
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Info("Received shutdown signal")

	// Initiate graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Trigger context cancellation to stop the refresh routine
	cancel()

	// Shutdown the HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error(fmt.Sprintf("Server shutdown error: %v", err))
	}

	// Wait for background tasks to complete
	wg.Wait()
	log.Info("Shutdown complete")
}
