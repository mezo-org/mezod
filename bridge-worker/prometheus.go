package bridgeworker

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var pendingBTCWithdrawalsGauge = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "pending_btc_withdrawals",
		Help: "the number of pending BTC withdrawals",
	},
)

func startPrometheus(port uint) {
	log.Printf("starting prometheus")

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	// Timeouts close slow or idle connections, so a client cannot hold
	// connections open and use up the process's file descriptors.
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("couldn't start prometheus http server: [%v]", err)
	}
}
