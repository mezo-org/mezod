package metricsscraper

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ethereumSidecarGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ethereum_sidecar_connectivity",
			Help: "the version of the ethereum sidecar",
		},
		[]string{"moniker", "chain_id", "sidecar_version"},
	)

	connectSidecarGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "connect_sidecar_connectivity",
			Help: "the version of the connect sidecar",
		},
		[]string{"moniker", "chain_id", "sidecar_version"},
	)

	mezodVersionGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mezod_version",
			Help: "the mezod version of a node",
		},
		[]string{"moniker", "chain_id", "version"},
	)

	mezoLatestBlockGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "latest_block",
			Help: "the latest block processed by a node",
		},
		[]string{"moniker", "chain_id"},
	)

	mezoLatestTimestampGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "latest_timestamp",
			Help: "the latest timestamps at which a node processed a block",
		},
		[]string{"moniker", "chain_id"},
	)

	mezodUpGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "up",
			Help: "the state of the node 1 for up, 0 for down",
		},
		[]string{"moniker", "chain_id"},
	)
)

func startPrometheus(port uint) {
	log.Printf("starting prometheus")

	http.Handle("/metrics", promhttp.Handler())

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil) //nolint:gosec
	if err != nil {
		log.Fatalf("couldn't start prometheus http server: [%v]", err)
	}
}
