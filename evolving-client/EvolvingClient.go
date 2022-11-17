package evolving_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yuhao-jack/evolving-rpc/contents"
	"github.com/yuhao-jack/evolving-rpc/model"
	"github.com/yuhao-jack/go-toolx/fun"
	"github.com/yuhao-jack/go-toolx/netx"
	"log"
	"sync"
	"time"
)

var logger = log.Default()

func init() {
	logger.SetFlags(log.Llongfile | log.Ldate | log.Lmicroseconds)
}

// EvolvingClient
// @Description: 客户端连接（非RPC客户端）
type EvolvingClient struct {
	msgChan  chan netx.IMessage
	dataPack *netx.DataPack
	conf     *model.EvolvingClientConfig
	commands map[string]func(message netx.IMessage)
	lock     *sync.RWMutex
}

// NewEvolvingClient
//
//	@Description: 创建客户端连接（非RPC客户端）
//	@param conf 创建客户端的配置
//	@return *EvolvingClient 客户端连接
func NewEvolvingClient(conf *model.EvolvingClientConfig) *EvolvingClient {
	evolvingClient := EvolvingClient{msgChan: make(chan netx.IMessage, 1024),
		conf:     conf,
		commands: make(map[string]func(message netx.IMessage)),
		lock:     &sync.RWMutex{},
	}
	evolvingClient.createConn()
	evolvingClient.SetCommand(contents.Default, func(reply netx.IMessage) {
		logger.Println(string(reply.GetCommand()), string(reply.GetBody()))
	})
	go evolvingClient.processMsg()
	go evolvingClient.sendMsg()
	return &evolvingClient
}

// Execute
//
//	@Description: 连接执行的命令
//	@receiver c
//	@param req 入参
//	@param callBack 回调方法
func (c *EvolvingClient) Execute(req netx.IMessage, callBack func(reply netx.IMessage)) {
	if callBack != nil {
		c.SetCommand(string(req.GetCommand()), callBack)
	}
	c.msgChan <- req
}

// SetCommand
//
//	@Description: 设置命令
//	@receiver c
//	@param command 命令
//	@param f 执行方法
func (c *EvolvingClient) SetCommand(command string, f func(reply netx.IMessage)) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.commands[command] = f
}

// GetCommand
//
//	@Description: 获取命令
//	@receiver c
//	@param command 命令
//	@return f 执行方法
func (c *EvolvingClient) GetCommand(command string) (f func(reply netx.IMessage)) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	f = c.commands[command]
	return f
}

// start
//
//	@Description: 创建连接
//	@receiver c
func (c *EvolvingClient) createConn() {
	conn, err := netx.CreateTcpConn(fmt.Sprintf("%s:%d", c.conf.EvolvingServerHost, c.conf.EvolvingServerPort))
	if err != nil {
		logger.Println("start evolving-client failed,err:", err)
		return
	}

	dataPack := netx.DataPack{}
	dataPack.Conn = conn
	c.dataPack = &dataPack
	logger.Println("start evolving-client successful.")
}

// sendMsg
//
//	@Description: 发送消息，这里真正将数据包发送到网络上
//	@receiver c
func (c *EvolvingClient) sendMsg() {
	ticker := time.NewTicker(c.conf.HeartbeatInterval)
	for {
		select {
		case <-ticker.C:
			err := c.dataPack.Pack([]byte(contents.ALive), nil)
			if err != nil {
				logger.Println(err)
				continue
			}
		case msg := <-c.msgChan:
			err := c.dataPack.PackMessage(msg)
			if err != nil {
				logger.Println(err)
				continue
			}
		}
	}
}

// processMsg
//
//	@Description: 处理接受的消息，这里是真正的从网络上拿到数据包并执行对应的函数
//	@receiver c
func (c *EvolvingClient) processMsg() {
	for {
		message, err := c.dataPack.UnPackMessage()
		if err != nil {
			logger.Println(err)
			break
		}
		f := c.GetCommand(string(message.GetCommand()))
		fun.IfOr(f != nil, f, c.GetCommand(contents.Default))(message)
	}
}

// RegisterService
//
//	@Description: 把服务注册到注册中心
//	@receiver c
//	@param info 服务的详情信息
//	@param callBack 注册后的回调方法
//	@return error 注册失败时的错误信息
func (c *EvolvingClient) RegisterService(info *model.ServiceInfo, callBack func(reply netx.IMessage)) error {
	if info == nil {
		return errors.New("info or dataPack is nil")
	}
	bytes, err := json.Marshal(info)
	if err != nil {
		return err
	}
	iMessage := netx.NewDefaultMessage([]byte(contents.Register), bytes)
	c.Execute(iMessage, callBack)
	return nil
}

// DisCover
//
//	@Description: 发现服务
//	@receiver c
//	@param serviceName 服务名
//	@param callBack 发现服务后的回调函数
//	@return error 发现失败时的错误信息
func (c *EvolvingClient) DisCover(serviceName string, callBack func(reply netx.IMessage)) error {
	if serviceName == "" {
		return errors.New("serviceName is nil ")
	}
	iMessage := netx.NewDefaultMessage([]byte(contents.DisCover), []byte(serviceName))
	c.Execute(iMessage, callBack)
	return nil
}
