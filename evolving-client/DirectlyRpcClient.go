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
	client     *EvolvingClient
	signalLock *sync.Mutex
}

// NewDirectlyRpcClient
//
//	@Description: 创建直连模式下的Rpc客户端
//	@param config  直连模式下的Rpc客户端的配置
//	@return *DirectlyRpcClient 直连模式下的Rpc客户端
func NewDirectlyRpcClient(config *DirectlyRpcClientConfig) *DirectlyRpcClient {
	client := NewEvolvingClient(&config.EvolvingClientConfig)
	if client == nil {
		return nil
	}
	return &DirectlyRpcClient{client: client, signalLock: &sync.Mutex{}}
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
	var signalChan chan struct{}
	if isAsync {
		d.signalLock.Lock()
		signalChan = make(chan struct{})
		defer close(signalChan)
	}
	d.client.Execute(netx.NewDefaultMessage([]byte(command), req), func(reply netx.IMessage) {
		res = reply.GetBody()
		if isAsync {
			d.signalLock.Unlock()
			signalChan <- struct{}{}
		}

	})
	if isAsync {
		<-signalChan
	}
	return res, nil
}

func (d *DirectlyRpcClient) ExecuteCmd(command string, req []byte, callBack func([]byte)) {
	d.client.Execute(netx.NewDefaultMessage([]byte(command), req), func(reply netx.IMessage) {
		callBack(reply.GetBody())
	})
}

// Close
//
//	@Description: 关闭客户端
//	@receiver c
//	@Author yuhao
//	@Data 2023-03-01 21:03:07
func (d *DirectlyRpcClient) Close() {
	d.client.Close()
}
