package contents

const (
	Register      = "REGISTER"
	DisCover      = "DISCOVER"
	OK            = "OK"
	ALive         = "ALIVE"
	Default       = "DEFAULT"
	ConnectClosed = "CONNECT_CLOSED"
)

type ServiceProtoc string

const (
	Http  ServiceProtoc = "http"
	Https ServiceProtoc = "https"
	Grpc  ServiceProtoc = "grpc"
)
