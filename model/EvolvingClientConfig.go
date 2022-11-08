package model

import "time"

// EvolvingClientConfig
// @Description:
type EvolvingClientConfig struct {
	EvolvingServerHost string        `json:"evolving_server_host"`
	EvolvingServerPort int32         `json:"evolving_server_port"`
	HeartbeatInterval  time.Duration `json:"heartbeat_interval"`
}
