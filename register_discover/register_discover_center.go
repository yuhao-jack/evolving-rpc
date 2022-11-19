package main

import (
	evolvingserver "github.com/yuhao-jack/evolving-rpc/evolving-server"
	"github.com/yuhao-jack/evolving-rpc/model"
)

func main() {
	serverConf := model.EvolvingServerConf{
		BindHost:   "0.0.0.0",
		ServerPort: 6601,
	}
	evolvingServer := evolvingserver.NewEvolvingServer(&serverConf)
	evolvingServer.Start()

}
