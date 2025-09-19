package bridgeworker

import (
	"fmt"
	"log"
	"net/http"

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

	http.Handle("/metrics", promhttp.Handler())

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil) //nolint:gosec
	if err != nil {
		log.Fatalf("couldn't start prometheus http server: [%v]", err)
	}
}
