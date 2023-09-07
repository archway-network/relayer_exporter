package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/archway-network/relayer_exporter/pkg/collector"
	"github.com/archway-network/relayer_exporter/pkg/config"
	log "github.com/archway-network/relayer_exporter/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	log.Info(fmt.Sprintf("Getting IBC paths from %s/%s/%s on GitHub", cfg.GitHub.Org, cfg.GitHub.Repo, cfg.GitHub.IBCDir))

	// TODO: Add a feature to refresh paths at configured interval
	paths, err := cfg.IBCPaths()
	if err != nil {
		log.Fatal(err.Error())
	}

	rpcs := cfg.GetRPCsMap()

	clientsCollector := collector.IBCClientsCollector{
		RPCs:  rpcs,
		Paths: paths,
	}

	prometheus.MustRegister(clientsCollector)

	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", *port)
	log.Info(fmt.Sprintf("Starting server on addr: %s", addr))
	log.Fatal(http.ListenAndServe(addr, nil).Error())
}
