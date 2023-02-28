package test

import (
	"encoding/json"
	"fmt"
	evolving_client "github.com/yuhao-jack/evolving-rpc/evolving-client"
	evolving_server "github.com/yuhao-jack/evolving-rpc/evolving-server"
	"github.com/yuhao-jack/evolving-rpc/model"
	go_log "github.com/yuhao-jack/go-log"
	"log"

	"testing"
	"time"
)

var logger = go_log.GetSingleGoLog()

func beforeTestDistributedRpc() {
	//  开启注册服务
	//serverConf := model.EvolvingServerConf{
	//	BindHost:   "0.0.0.0",
	//	ServerPort: 6601,
	//}
	//evolvingServer := evolving_server.NewEvolvingServer(&serverConf)
	//go evolvingServer.Start()
	time.Sleep(time.Second)

	//  注册中心的配置
	config := model.EvolvingClientConfig{
		EvolvingServerHost: "0.0.0.0",
		EvolvingServerPort: 6601,
		HeartbeatInterval:  5 * time.Minute,
	}
	//  要注册的服务的信息
	serviceInfo := model.ServiceInfo{
		ServiceName:    "Arith",
		ServiceHost:    "0.0.0.0",
		ServicePort:    3301,
		ServiceProtoc:  "prc",
		AdditionalMeta: map[string]any{},
	}
	//  这里尝试注册3个
	for i := 0; i < 3; i++ {
		serviceInfo.ServicePort = serviceInfo.ServicePort + 1
		logger.Info(fmt.Sprintf("%v", serviceInfo))
		rpcServer := evolving_server.NewDistributedRpcServer(&config, &serviceInfo)
		err := rpcServer.Register(new(Arith))
		if err != nil {
			log.Default().Println(err)
			return
		}
		go rpcServer.Run()
	}
	//  让子弹飞一会
	time.Sleep(3 * time.Second)
}

func TestDistributedRpc(t *testing.T) {
	defer logger.Destroy()
	beforeTestDistributedRpc()
	//  注册中心的配置
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
		logger.Error(err.Error())
		return
	}
	res, err := rpcClient.ExecuteCommand("Arith", "Arith.Multiply", bytes, true)
	logger.Warn("%s,%v", string(res), err)
	res, err = rpcClient.ExecuteCommand("Arith", "Arith.Divide", bytes, true)
	logger.Warn("%s,%v", string(res), err)
}
