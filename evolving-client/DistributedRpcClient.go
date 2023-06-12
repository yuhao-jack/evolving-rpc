package evolving_client

import (
	"encoding/json"
	"errors"
	"github.com/yuhao-jack/evolving-rpc/contents"
	"github.com/yuhao-jack/evolving-rpc/model"
	"github.com/yuhao-jack/go-toolx/fun"
	"github.com/yuhao-jack/go-toolx/netx"
	"hash/crc32"
	"sync"
	"time"
)

type ModeType string

type DistributedRpcClient struct {
	registerCenterConfigs []*model.EvolvingClientConfig
	serviceInfoMap        map[string][]*model.ServiceInfo
	serviceClientMap      map[string][]*EvolvingClient
	evolvingClient        []*EvolvingClient
	mode                  ModeType
}

func NewDistributedRpcClient(registerCenterConfigs []*model.EvolvingClientConfig, dependentServices []string) (c *DistributedRpcClient) {
	rpcClient := DistributedRpcClient{registerCenterConfigs: registerCenterConfigs, serviceInfoMap: map[string][]*model.ServiceInfo{}, serviceClientMap: map[string][]*EvolvingClient{}}
	for _, config := range registerCenterConfigs {
		evolvingClient := NewEvolvingClient(config)
		if evolvingClient != nil {
			rpcClient.evolvingClient = append(rpcClient.evolvingClient, evolvingClient)
		}
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
		if len(rpcClient.evolvingClient) == 0 {
			contents.RpcLogger.Error("service %s has no available nodes...", service)
			return nil
		}
		waitGroup.Add(1)
		_ = rpcClient.evolvingClient[0].DisCover(service, func(reply netx.IMessage) {
			err := json.Unmarshal(reply.GetBody(), &serviceList)
			if err != nil {
				contents.RpcLogger.Error("json.Unmarshal failed,err:%v", err)
			}
			waitGroup.Done()
		})
		waitGroup.Wait()
		rpcClient.serviceInfoMap[service] = serviceList
		if rpcClient.serviceClientMap == nil {
			rpcClient.serviceClientMap = map[string][]*EvolvingClient{}
		}
		for _, info := range serviceList {
			client := NewEvolvingClient(&model.EvolvingClientConfig{
				EvolvingServerHost: info.ServiceHost,
				EvolvingServerPort: info.ServicePort,
				HeartbeatInterval:  60 * time.Second,
			})
			if client != nil {
				rpcClient.serviceClientMap[service] = append(rpcClient.serviceClientMap[service], client)
			}

		}
		group.Done()
	}
	group.Wait()

	return &rpcClient
}

func (c *DistributedRpcClient) ExecuteCommand(serviceName, command string, req []byte, isSync bool) (res []byte, err error) {
	clients, ok := c.serviceClientMap[serviceName]
	if !ok {
		return nil, errors.New("service " + serviceName + " not found")
	}
	if len(clients) == 0 {
		return nil, errors.New("service " + serviceName + " has no provider")
	}
	group := sync.WaitGroup{}
	if isSync {
		group.Add(1)
	}
	clients[c.getClientsIndex(command, len(clients))].Execute(netx.NewDefaultMessage([]byte(command), req), func(reply netx.IMessage) {
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

// Close
//
//	@Description: 关闭客户端
//	@receiver c
//	@Author yuhao
//	@Data 2023-03-01 21:03:07
func (c *DistributedRpcClient) Close() {
	for _, clients := range c.serviceClientMap {
		for _, client := range clients {
			client.Close()
		}
	}
}

// getClientsIndex
//
//	@Description:
//	@receiver c
//	@param in
//	@param hashBy
//	@return int
func (c *DistributedRpcClient) getClientsIndex(in string, hashBy int) int {
	v := int(crc32.ChecksumIEEE([]byte(in)))
	t := fun.IfOr(v >= 0, v, -v) % hashBy
	return t
}
