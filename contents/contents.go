package contents

import go_log "github.com/yuhao-jack/go-log"

const (
	Register      = "REGISTER"
	DisCover      = "DISCOVER"
	OK            = "OK"
	ALive         = "ALIVE"
	Default       = "DEFAULT"
	ConnectClosed = "CONNECT_CLOSED"
)
const (
	Json = "json"
	Pb   = "pb"
)

type ServiceProtoc string

func (k ServiceProtoc) String() string { return string(k) }

const (
	Http  ServiceProtoc = "http"
	Https ServiceProtoc = "https"
	Grpc  ServiceProtoc = "grpc"
)

var RpcLogger = go_log.DefaultGoLog()

type ServiceStatus string

func (k ServiceStatus) String() string { return string(k) }

const (
	Up   ServiceStatus = "UP"
	Down ServiceStatus = "DOWN"
)

type AdditionalMetaKey string

func (k AdditionalMetaKey) String() string { return string(k) }

const (
	LostTime AdditionalMetaKey = "lost_time"
	Status   AdditionalMetaKey = "status"
)
