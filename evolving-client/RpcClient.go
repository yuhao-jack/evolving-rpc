package evolving_client

import (
	"encoding/json"
	"errors"
	"github.com/yuhao-jack/evolving-rpc/model"
	"github.com/yuhao-jack/go-toolx/netx"
	"log"
	"sync"
	"time"
)

type ModeType string

const (
	Directly    ModeType = "directly"
	Distributed ModeType = "distributed"
)

type RpcClient struct {
	registerCenterConfigs []*model.EvolvingClientConfig
	serviceInfoMap        map[string][]*model.ServiceInfo
	serviceClientMap      map[string][]*EvolvingClient
	evolvingClient        []*EvolvingClient
	mode                  ModeType
}

func NewRpcClient(registerCenterConfigs []*model.EvolvingClientConfig, dependentServices []string) (c *RpcClient) {
	rpcClient := RpcClient{registerCenterConfigs: registerCenterConfigs, serviceInfoMap: map[string][]*model.ServiceInfo{}}
	for _, config := range registerCenterConfigs {
		evolvingClient := NewEvolvingClient(config)
		rpcClient.evolvingClient = append(rpcClient.evolvingClient, evolvingClient)
	}
	group := sync.WaitGroup{}
	for _, service := range dependentServices {
		group.Add(1)
		var serviceList []*model.ServiceInfo
		waitGroup := sync.WaitGroup{}

		for _, client := range rpcClient.evolvingClient {
			waitGroup.Add(1)
			var tmpErr error
			Err := client.DisCover(service, func(reply netx.IMessage) {
				err := json.Unmarshal(reply.GetBody(), &serviceList)
				if err != nil {
					tmpErr = err
				}
				waitGroup.Done()
			})

			if tmpErr == nil && Err == nil {
				break
			}
		}
		waitGroup.Wait()

		waitGroup.Add(1)
		rpcClient.evolvingClient[0].DisCover(service, func(reply netx.IMessage) {
			err := json.Unmarshal(reply.GetBody(), &serviceList)
			if err != nil {
				log.Default().Println("json.Unmarshal failed,err:", err)
			}
			waitGroup.Done()

		})
		waitGroup.Wait()
		rpcClient.serviceInfoMap[service] = serviceList
		if rpcClient.serviceClientMap == nil {
			rpcClient.serviceClientMap = map[string][]*EvolvingClient{}
		}
		for _, info := range serviceList {
			rpcClient.serviceClientMap[service] = append(rpcClient.serviceClientMap[service], NewEvolvingClient(&model.EvolvingClientConfig{
				EvolvingServerHost: info.ServiceHost,
				EvolvingServerPort: info.ServicePort,
				HeartbeatInterval:  60 * time.Second,
			}))

		}
		group.Done()
	}
	group.Wait()

	return &rpcClient
}

func (c *RpcClient) ExecuteCommand(serviceName, command string, req []byte, isSync bool) (res []byte, err error) {
	clients, ok := c.serviceClientMap[serviceName]
	if !ok {
		return nil, errors.New("service " + serviceName + " not found")
	}
	group := sync.WaitGroup{}
	if isSync {
		group.Add(1)
	}
	clients[0].Execute(netx.NewDefaultMessage([]byte(command), req), func(reply netx.IMessage) {
		res = reply.GetBody()
		if isSync {
			group.Done()
		}
	})
	if isSync {
		group.Wait()
	}

	return res, nil
}
