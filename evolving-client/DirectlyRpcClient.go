package evolving_client

import (
	"gitee.com/yuhao-jack/evolving-rpc/model"
	"gitee.com/yuhao-jack/go-toolx/netx"
	"sync"
)

type DirectlyRpcClientConfig struct {
	//ServerHost string `json:"server_host"`
	//ServerPort int32 `json:"server_port"`
	model.EvolvingClientConfig
}

type DirectlyRpcClient struct {
	client *EvolvingClient
}

func NewDirectlyRpcClient(config *DirectlyRpcClientConfig) *DirectlyRpcClient {

	client := NewEvolvingClient(&config.EvolvingClientConfig)

	return &DirectlyRpcClient{client}
}

func (d *DirectlyRpcClient) ExecuteCommand(command string, req []byte, isSync bool) (res []byte, err error) {
	group := sync.WaitGroup{}
	if isSync {
		group.Add(1)
	}
	d.client.Execute(netx.NewDefaultMessage([]byte(command), req), func(reply netx.IMessage) {
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
