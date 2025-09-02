package metricsscraper

import (
	"encoding/json"
	"time"
)

type Config struct {
	ChainID string `json:"chain_id"`

	NodePollRate Duration     `json:"node_poll_rate"`
	Nodes        []NodeConfig `json:"nodes"`

	BridgePollRate Duration `json:"bridge_poll_rate"`
}

type NodeConfig struct {
	RPCURL  string `json:"rpc_url"`
	Moniker string `json:"moniker"`
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	parsed, err := time.ParseDuration(v)
	if err != nil {
		return err
	}

	d.Duration = parsed
	return nil
}
