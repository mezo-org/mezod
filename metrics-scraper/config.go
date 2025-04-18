package metricsscraper

import (
	"encoding/json"
	"time"
)

type Config struct {
	PollRate Duration     `json:"poll_rate"`
	Nodes    []NodeConfig `json:"nodes"`
	ChainID  string       `json:"chain_id"`
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
