package monitoring

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ethereumSidecarGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mezo_ethereum_sidecar_is_running",
			Help: "the version of the ethereum sidecar",
		},
		[]string{"moniker", "network", "version"},
	)

	connectSidecarGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mezo_connect_sidecar_is_running",
			Help: "the version of the connect sidecar",
		},
		[]string{"moniker", "network", "version"},
	)

	mezodVersionGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mezo_mezod_version",
			Help: "the mezod version of a node",
		},
		[]string{"moniker", "network", "version"},
	)

	mezoLatestBlockGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mezo_latest_block",
			Help: "the latest block processed by a node",
		},
		[]string{"moniker", "network"},
	)

	mezoLatestTimestampGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mezo_latest_timestamp",
			Help: "the latest timestamps at which a node processed a block",
		},
		[]string{"moniker", "network"},
	)
)

func startPrometheus() {
	http.Handle("/metrics", promhttp.Handler())

	err := http.ListenAndServe(":2112", nil) //nolint:gosec
	if err != nil {
		log.Fatalf("error: couldn't start http server: %v", err)
	}
}
