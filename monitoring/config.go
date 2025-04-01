package monitoring

import (
	"encoding/json"
	"time"
)

type Config struct {
	PollRate Duration     `json:"poll_rate"`
	Nodes    []NodeConfig `json:"nodes"`
}

type NodeConfig struct {
	RPCURL    string `json:"rpc_url"`
	CometURL  string `json:"comet_url"`
	Moniker   string `json:"moniker"`
	NetworkID string `json:"network_id"`
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
