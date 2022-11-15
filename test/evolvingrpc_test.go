package test

import (
	"encoding/json"
	evolving_client "github.com/yuhao-jack/evolving-rpc/evolving-client"
	"github.com/yuhao-jack/evolving-rpc/model"
	"log"
	"testing"
	"time"
)

//func init() {
//	config := model.EvolvingClientConfig{
//		EvolvingServerHost: "0.0.0.0",
//		EvolvingServerPort: 6601,
//		HeartbeatInterval:  5 * time.Minute,
//	}
//	serviceInfo := model.ServiceInfo{
//		ServiceName:    "Arith",
//		ServiceHost:    "0.0.0.0",
//		ServicePort:    3301,
//		ServiceProtoc:  "prc",
//		AdditionalMeta: map[string]any{},
//	}
//	rpcServer := evolving_server.NewRpcServer(&config, &serviceInfo)
//	err := rpcServer.Register(new(Arith))
//	if err != nil {
//		log.Default().Println(err)
//		return
//	}
//	go rpcServer.Run()
//	time.Sleep(3 * time.Second)
//}

func TestEvolvingRpc(t *testing.T) {
	var registerCenterConfigs = []*model.EvolvingClientConfig{
		{
			EvolvingServerHost: "0.0.0.0",
			EvolvingServerPort: 6601,
			HeartbeatInterval:  5 * time.Minute,
		},
	}
	rpcClient := evolving_client.NewRpcClient(registerCenterConfigs, []string{"Arith"})
	bytes, err := json.Marshal(&ArithReq{
		A: 10,
		B: 2,
	})
	if err != nil {
		log.Default().Println(err)
		return
	}
	res, err := rpcClient.ExecuteCommand("Arith", "Arith.Multiply", bytes, true)
	log.Default().Println(res, err)

}
