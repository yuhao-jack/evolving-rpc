package test

import (
	"encoding/json"
	"fmt"
	evolving_client "github.com/yuhao-jack/evolving-rpc/evolving-client"
	evolving_server "github.com/yuhao-jack/evolving-rpc/evolving-server"
	"github.com/yuhao-jack/evolving-rpc/model"
	"log"
	"testing"
	"time"
)

func beforeTestDistributedRpc() {
	config := model.EvolvingClientConfig{
		EvolvingServerHost: "0.0.0.0",
		EvolvingServerPort: 6601,
		HeartbeatInterval:  5 * time.Minute,
	}
	serviceInfo := model.ServiceInfo{
		ServiceName:    "Arith",
		ServiceHost:    "0.0.0.0",
		ServicePort:    3301,
		ServiceProtoc:  "prc",
		AdditionalMeta: map[string]any{},
	}
	for i := 0; i < 3; i++ {
		serviceInfo.ServicePort = serviceInfo.ServicePort + 1
		fmt.Println(fmt.Sprintf("%v", serviceInfo))
		rpcServer := evolving_server.NewDistributedRpcServer(&config, &serviceInfo)
		err := rpcServer.Register(new(Arith))
		if err != nil {
			log.Default().Println(err)
			return
		}
		go rpcServer.Run()
	}
	time.Sleep(3 * time.Second)
}

func TestDistributedRpc(t *testing.T) {
	beforeTestDistributedRpc()
	var registerCenterConfigs = []*model.EvolvingClientConfig{
		{
			EvolvingServerHost: "0.0.0.0",
			EvolvingServerPort: 6601,
			HeartbeatInterval:  5 * time.Minute,
		},
	}
	rpcClient := evolving_client.NewDistributedRpcClient(registerCenterConfigs, []string{"Arith"})
	bytes, err := json.Marshal(&ArithReq{
		A: 10,
		B: 2,
	})
	if err != nil {
		log.Default().Println(err)
		return
	}
	res, err := rpcClient.ExecuteCommand("Arith", "Arith.Multiply", bytes, true)
	log.Default().Println(string(res), err)
	res, err = rpcClient.ExecuteCommand("Arith", "Arith.Divide", bytes, true)
	log.Default().Println(string(res), err)
}
