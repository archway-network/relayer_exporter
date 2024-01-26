package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

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

func main() {
	port := flag.Int("p", 8008, "Server port")
	version := flag.Bool("version", false, "Print version")
	configPath := flag.String("config", "./config.yml", "path to config file")
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

	ctx := context.Background()
	// TODO: Add a feature to refresh paths at configured interval
	paths, err := cfg.IBCPaths(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	rpcs, err := cfg.GetRPCsMap(paths)
	if err != nil {
		log.Fatal(err.Error())
	}

	ibcCollector := collector.IBCCollector{
		RPCs:          rpcs,
		Paths:         paths,
		AckProcessors: map[string]collector.AckProcessor{},
	}

	balancesCollector := collector.WalletBalanceCollector{
		RPCs:     rpcs,
		Accounts: cfg.Accounts,
	}

	prometheus.MustRegister(ibcCollector)
	prometheus.MustRegister(balancesCollector)

	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", *port)
	log.Info(fmt.Sprintf("Starting server on addr: %s", addr))
	log.Fatal(http.ListenAndServe(addr, nil).Error())
}
