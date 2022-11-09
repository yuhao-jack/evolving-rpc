package evolving_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitee.com/yuhao-jack/evolving-rpc/contents"
	"gitee.com/yuhao-jack/evolving-rpc/model"
	"gitee.com/yuhao-jack/go-toolx/netx"
	"log"
	"sync"
	"time"
)

var logger = log.Default()

func init() {
	logger.SetFlags(log.Llongfile | log.Ldate | log.Lmicroseconds)
}

// EvolvingClient
// @Description:
type EvolvingClient struct {
	msgChan  chan netx.IMessage
	dataPack *netx.DataPack
	conf     *model.EvolvingClientConfig
	commands map[string]func(message netx.IMessage)
	lock     sync.RWMutex
}

// NewEvolvingClient
//
//	@Description:
//	@param conf
//	@return *EvolvingClient
func NewEvolvingClient(conf *model.EvolvingClientConfig) *EvolvingClient {
	evolvingClient := EvolvingClient{msgChan: make(chan netx.IMessage, 1024), conf: conf, commands: make(map[string]func(message netx.IMessage))}
	evolvingClient.start()
	evolvingClient.SetCommand(contents.Default, func(reply netx.IMessage) {
		logger.Println(string(reply.GetCommand()), string(reply.GetBody()))
	})
	go evolvingClient.processMsg()
	go evolvingClient.sendMsg()
	return &evolvingClient
}

// Execute
//
//	@Description:
//	@receiver c
//	@param req
//	@param callBack
func (c *EvolvingClient) Execute(req netx.IMessage, callBack func(reply netx.IMessage)) {
	if callBack != nil {
		c.SetCommand(string(req.GetCommand()), callBack)
	}
	c.msgChan <- req
}

// SetCommand
//
//	@Description:
//	@receiver c
//	@param command
//	@param f
func (c *EvolvingClient) SetCommand(command string, f func(reply netx.IMessage)) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.commands[command] = f
}

// GetCommand
//
//	@Description:
//	@receiver c
//	@param command
//	@return f
func (c *EvolvingClient) GetCommand(command string) (f func(reply netx.IMessage)) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	f = c.commands[command]
	return f
}

// start
//
//	@Description:
//	@receiver c
func (c *EvolvingClient) start() {
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
//	@Description:
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
//	@Description:
//	@receiver c
func (c *EvolvingClient) processMsg() {
	for {
		message, err := c.dataPack.UnPackMessage()
		if err != nil {
			logger.Println(err)
			break
		}
		f := c.GetCommand(string(message.GetCommand()))
		if f != nil {
			f(message)
		} else {
			c.GetCommand(contents.Default)(message)
		}
	}
}

// RegisterService
//
//	@Description:
//	@receiver c
//	@param info
//	@param callBack
//	@return error
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
//	@Description:
//	@receiver c
//	@param serviceName
//	@param callBack
//	@return error
func (c *EvolvingClient) DisCover(serviceName string, callBack func(reply netx.IMessage)) error {
	if serviceName == "" {
		return errors.New("serviceName is nil ")
	}
	iMessage := netx.NewDefaultMessage([]byte(contents.DisCover), []byte(serviceName))
	c.Execute(iMessage, callBack)
	return nil
}
