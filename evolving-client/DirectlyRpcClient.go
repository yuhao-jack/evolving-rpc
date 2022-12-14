package evolving_client

import (
	"github.com/yuhao-jack/evolving-rpc/model"
	"github.com/yuhao-jack/go-toolx/netx"
	"sync"
)

// DirectlyRpcClientConfig
// @Description: 直连模式下的Rpc客户端的配置
type DirectlyRpcClientConfig struct {
	model.EvolvingClientConfig
}

// DirectlyRpcClient
// @Description: 直连模式下的Rpc客户端
type DirectlyRpcClient struct {
	client *EvolvingClient
}

// NewDirectlyRpcClient
//
//	@Description: 创建直连模式下的Rpc客户端
//	@param config  直连模式下的Rpc客户端的配置
//	@return *DirectlyRpcClient 直连模式下的Rpc客户端
func NewDirectlyRpcClient(config *DirectlyRpcClientConfig) *DirectlyRpcClient {
	client := NewEvolvingClient(&config.EvolvingClientConfig)
	return &DirectlyRpcClient{client}
}

// ExecuteCommand
//
//	@Description: 执行命令
//	@receiver d
//	@param command 命令 eg:Arith.Multiply
//	@param req 命令入参
//	@param isSync 是否同步
//	@return res 命令结果
//	@return err 失败时的错误信息
func (d *DirectlyRpcClient) ExecuteCommand(command string, req []byte, isAsync bool) (res []byte, err error) {
	group := sync.WaitGroup{}
	if isAsync {
		group.Add(1)
	}
	d.client.Execute(netx.NewDefaultMessage([]byte(command), req), func(reply netx.IMessage) {
		res = reply.GetBody()
		if isAsync {
			group.Done()
		}

	})
	if isAsync {
		group.Wait()
	}
	return res, nil
}
