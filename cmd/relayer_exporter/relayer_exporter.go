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
	err := refreshIBCCollector(ctx, cfg, registry)
	if err != nil {
		return err
	}

	err = refreshWalletBalanceCollector(cfg, registry)
	if err != nil {
		return err
	}

	return nil
}

func refreshWalletBalanceCollector(cfg *config.Config, registry *prometheus.Registry) error {
	if len(cfg.Accounts) == 0 {
		log.Warn("No accounts configured, skipping wallet balance collector refresh")
		return nil
	}

	rpcs := cfg.GetRPCsMap()
	// Unregister existing collectors
	registry.Unregister(collector.WalletBalanceCollector{})

	// Create and register new collector
	balancesCollector := collector.WalletBalanceCollector{
		RPCs:     rpcs,
		Accounts: cfg.Accounts,
	}

	registry.MustRegister(balancesCollector)

	return nil
}

// refreshIBCCollectors updates the IBC collector with new paths
func refreshIBCCollector(ctx context.Context, cfg *config.Config, registry *prometheus.Registry) error {
	paths, err := cfg.IBCPaths(ctx)
	if err != nil {
		log.Warn("Failed to get IBC paths, skipping IBC collector refresh", zap.Error(err))
		return nil
	}

	if len(paths) > 0 {
		rpcs := cfg.GetRPCsMap()
		// Unregister existing collector
		registry.Unregister(collector.IBCCollector{})

		// Create and register new collector
		ibcCollector := collector.IBCCollector{
			RPCs:  rpcs,
			Paths: paths,
		}
		registry.MustRegister(ibcCollector)
	}

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

	// Create a context with cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	registry := prometheus.NewRegistry()

	// Initial setup of collectors
	if err := refreshCollectors(ctx, cfg, registry); err != nil {
		log.Fatal(err.Error())
	}

	// Defer cancel after all fatal errors
	defer cancel() // Ensure context is cancelled when main exits

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
