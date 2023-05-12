package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/archway-network/relayer_exporter/pkg/collector"
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
	rly := flag.String("rly", "/home/archway/go/bin/rly", "Path to rly binary")
	logLevel := log.LevelFlag()

	flag.Parse()

	log.SetLevel(*logLevel)

	if *version {
		fmt.Println(getVersion())
		os.Exit(0)
	}

	rc := collector.RelayerCollector{Rly: *rly}
	prometheus.MustRegister(rc)

	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", *port)
	log.Info(fmt.Sprintf("Starting server on addr: %s", addr))
	log.Fatal(http.ListenAndServe(addr, nil).Error())
}
