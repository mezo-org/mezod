package monitoring

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ethereumSidecarGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ethereum_sidecar_is_running",
			Help: "the version of the ethereum sidecar",
		},
		[]string{"moniker", "network", "version"},
	)

	connectSidecarGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "connect_sidecar_is_running",
			Help: "the version of the connect sidecar",
		},
		[]string{"moniker", "network", "version"},
	)

	mezodVersionGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mezod_version",
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
)

func startPrometheus() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
