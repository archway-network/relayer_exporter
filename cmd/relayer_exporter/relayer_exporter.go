package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/archway-network/relayer_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getVersion() string {
	version, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	return strings.TrimSpace(version.Main.Version)
}

func main() {
	port := flag.Int("p", 8008, "Server port")
	version := flag.Bool("version", false, "Print version")
	rly := flag.String("rly", "/home/archway/go/bin/rly", "Path to rly binary")

	flag.Parse()

	if *version {
		fmt.Println(getVersion())
		os.Exit(0)
	}

	rc := collector.RelayerCollector{Rly: *rly}
	prometheus.MustRegister(rc)

	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server on addr: %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
