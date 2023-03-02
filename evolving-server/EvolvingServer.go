package evolving_server

import (
	"encoding/json"
	"fmt"
	"github.com/yuhao-jack/evolving-rpc/contents"
	"github.com/yuhao-jack/evolving-rpc/errorx"
	"github.com/yuhao-jack/evolving-rpc/evolving-server/svr_mgr"
	"github.com/yuhao-jack/evolving-rpc/model"
	"github.com/yuhao-jack/go-toolx/fun"
	"github.com/yuhao-jack/go-toolx/netx"
	"net"
	"sync"
	"time"
)

// EvolvingServer
// @Description: 服务端连接（非RPC服务端）
type EvolvingServer struct {
	conf            *model.EvolvingServerConf
	dataPackChanMap map[*netx.DataPack]chan netx.IMessage
	commands        map[string]func(dataPack *netx.DataPack, reply netx.IMessage)
	dataPackLock    *sync.RWMutex
	commandLock     *sync.RWMutex
	closeFlag       bool
}

// NewEvolvingServer
//
//	@Description: 创建一个服务端连接（非RPC服务端）
//	@param conf 创建服务端的配置
//	@return *EvolvingServer
func NewEvolvingServer(conf *model.EvolvingServerConf) *EvolvingServer {
	evolvingServer := EvolvingServer{
		conf:            conf,
		dataPackChanMap: make(map[*netx.DataPack]chan netx.IMessage),
		commands:        make(map[string]func(dataPack *netx.DataPack, reply netx.IMessage)),
		commandLock:     &sync.RWMutex{},
		dataPackLock:    &sync.RWMutex{},
	}
	//  heartbeat
	evolvingServer.SetCommand(contents.ALive, func(dataPack *netx.DataPack, reply netx.IMessage) {
		evolvingServer.sendMsg(dataPack, netx.NewDefaultMessage([]byte(contents.ALive), []byte(contents.OK)))
	})
	//  default
	evolvingServer.SetCommand(contents.Default, func(dataPack *netx.DataPack, reply netx.IMessage) {
		Default(reply, dataPack, evolvingServer.sendMsg)
	})
	//  register
	evolvingServer.SetCommand(contents.Register, func(dataPack *netx.DataPack, reply netx.IMessage) {
		Register(reply, dataPack, evolvingServer.sendMsg)
	})
	// discover
	evolvingServer.SetCommand(contents.DisCover, func(dataPack *netx.DataPack, reply netx.IMessage) {
		DisCover(reply, dataPack, evolvingServer.sendMsg)
	})
	return &evolvingServer
}

// Start
//
//	@Description: 启动服务端
//	@receiver s
func (s *EvolvingServer) Start() {
	tcpListener, err := netx.CreateTCPListener(fmt.Sprintf("%s:%d", s.conf.BindHost, s.conf.ServerPort))
	if err != nil {
		contents.RpcLogger.Error("start evolving-server failed,err:%v", err)
		return
	}
	contents.RpcLogger.Info("start evolving-server successful.")
	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			contents.RpcLogger.Error("accept tcp conn failed,err:%v", err)
			continue
		}
		go s.connHandler(tcpConn)
	}
}

func (s *EvolvingServer) Close() {
	for dataPack, _ := range s.dataPackChanMap {
		dataPack.Close()
	}
	s.closeFlag = true
}

// connHandler
//
//	@Description: 新建连接处理
//	@param conn 客户端连接
func (s *EvolvingServer) connHandler(conn *net.TCPConn) {
	dataPack := netx.DataPack{Conn: conn}
	svr_mgr.GetServiceMgrInstance().AddDataPack(&dataPack)
	var serviceInfo model.ServiceInfo
	defer func() { // 客户端端开后广播到其他客户端
		svr_mgr.GetServiceMgrInstance().DelDataPack(&dataPack)
		if !fun.IsBlank(serviceInfo) {
			svr_mgr.GetServiceMgrInstance().ServiceInfoList.ForEach(func(info *model.ServiceInfo) {
				if info.ServiceName == serviceInfo.ServiceName &&
					info.ServiceHost == serviceInfo.ServiceHost &&
					info.ServicePort == serviceInfo.ServicePort {
					info.AdditionalMeta[contents.Status.String()] = contents.Down
					info.AdditionalMeta[contents.LostTime.String()] = time.Now()
				}
			})
		}
		err := conn.Close()
		if err != nil {
			contents.RpcLogger.Error(err.Error())
		}
		if s.dataPackChanMap[&dataPack] != nil {
			close(s.dataPackChanMap[&dataPack])
		}
		delete(s.dataPackChanMap, &dataPack)
		s.broadCast(netx.NewDefaultMessage([]byte(contents.ConnectClosed), []byte(dataPack.RemoteAddr().String()+" disconnected")))
	}()
	for {
		message, err := dataPack.UnPackMessage()
		if err != nil {
			contents.RpcLogger.Error(err.Error())
			break
		}
		command := string(message.GetCommand())
		if command == contents.Register {
			err = json.Unmarshal(message.GetBody(), &serviceInfo)
			if err != nil {
				contents.RpcLogger.Error(err.Error())
			}
		}
		f := s.GetCommand(command)
		fun.IfOr(f != nil, f, s.GetCommand(contents.Default))(&dataPack, message)
	}
}

// Execute
//
//	@Description: 执行命令
//	@receiver s
//	@param dataPack
//	@param req
//	@param callBack
func (s *EvolvingServer) Execute(dataPack *netx.DataPack, req netx.IMessage, callBack func(dataPack *netx.DataPack, reply netx.IMessage)) {
	s.SetCommand(string(req.GetCommand()), callBack)
	s.sendMsg(dataPack, req)
}

// SetCommand
//
//	@Description:
//	@receiver s
//	@param command
//	@param f
func (s *EvolvingServer) SetCommand(command string, f func(dataPack *netx.DataPack, reply netx.IMessage)) {
	s.commandLock.Lock()
	defer s.commandLock.Unlock()
	if f != nil {
		s.commands[command] = f
	}
}

// GetCommand
//
//	@Description:
//	@receiver s
//	@param command
//	@return f
func (s *EvolvingServer) GetCommand(command string) (f func(dataPack *netx.DataPack, reply netx.IMessage)) {
	s.commandLock.RLock()
	defer s.commandLock.RUnlock()
	f = s.commands[command]
	return f
}

// SetDataPackChanMap
//
//	@Description:
//	@receiver s
//	@param dataPack
//	@param c
func (s *EvolvingServer) SetDataPackChanMap(dataPack *netx.DataPack, c chan netx.IMessage) {
	s.dataPackLock.Lock()
	defer s.dataPackLock.Unlock()
	if c != nil {
		s.dataPackChanMap[dataPack] = c
	}
}

// GetDataPackChanMap
//
//	@Description:
//	@receiver s
//	@param dataPack
//	@return c
func (s *EvolvingServer) GetDataPackChanMap(dataPack *netx.DataPack) (c chan netx.IMessage) {
	s.dataPackLock.RLock()
	defer s.dataPackLock.RUnlock()
	c = s.dataPackChanMap[dataPack]
	return c
}

// broadCast
//
//	@Description:  广播
//	@param msg 需要广播的消息
func (s *EvolvingServer) broadCast(msg netx.IMessage) {
	svr_mgr.GetServiceMgrInstance().DataPackMap.Each(func(key string, val *netx.DataPack) {
		s.sendMsg(val, msg)
	})
}

// sendMsg
//
//	@Description: 消息发送
//	@param dataPack
//	@param message
func (s *EvolvingServer) sendMsg(dataPack *netx.DataPack, message netx.IMessage) {
	dataPackChanMap := s.GetDataPackChanMap(dataPack)
	if dataPackChanMap == nil {
		s.SetDataPackChanMap(dataPack, make(chan netx.IMessage, 1024))
		go func() {
			for !s.closeFlag {
				select {
				case msg, ok := <-s.GetDataPackChanMap(dataPack):
					if ok {
						err := dataPack.PackMessage(msg)
						if err != nil {
							contents.RpcLogger.Error(err.Error())
						}
					} else {
						contents.RpcLogger.Warn(dataPack.RemoteAddr().String() + ": closed")
						break
					}
				}
			}
			contents.RpcLogger.Warn("Server closed...")
		}()
	}
	s.GetDataPackChanMap(dataPack) <- message
}

// KeepAlive
//
//	@Description:
//	@param message
//	@param dataPack
func KeepAlive(message netx.IMessage, dataPack *netx.DataPack, sendMsg func(dataPack *netx.DataPack, message netx.IMessage)) {
	sendMsg(dataPack, message)
}

// Register
//
//	@Description:
//	@param message
//	@param dataPack
func Register(message netx.IMessage, dataPack *netx.DataPack, sendMsg func(dataPack *netx.DataPack, message netx.IMessage)) {
	var serviceInfo model.ServiceInfo
	err := json.Unmarshal(message.GetBody(), &serviceInfo)
	if err != nil {
		contents.RpcLogger.Error(err.Error())
		contents.RpcLogger.Warn(string(message.GetBody()))
		return
	}
	needInsert := true
	svr_mgr.GetServiceMgrInstance().ServiceInfoList.ForEach(func(info *model.ServiceInfo) {
		if info.ServiceName == serviceInfo.ServiceName &&
			info.ServiceHost == serviceInfo.ServiceHost &&
			info.ServicePort == serviceInfo.ServicePort {
			info.AdditionalMeta = serviceInfo.AdditionalMeta
			info.AdditionalMeta[contents.Status.String()] = contents.Up
			delete(info.AdditionalMeta, contents.LostTime.String())
			info.ServiceProtoc = serviceInfo.ServiceProtoc
			needInsert = false
		}
	})
	if needInsert {
		svr_mgr.GetServiceMgrInstance().AddServiceInfo(&serviceInfo)
	}
	KeepAlive(message, dataPack, sendMsg)
}

// DisCover
//
//	@Description:
//	@param message
//	@param dataPack
func DisCover(message netx.IMessage, dataPack *netx.DataPack, sendMsg func(dataPack *netx.DataPack, message netx.IMessage)) {
	list := svr_mgr.GetServiceMgrInstance().FindServiceInfosByServiceName(string(message.GetBody()))
	bytes, err := json.Marshal(list)
	if err != nil {
		contents.RpcLogger.Error(err.Error())
		return
	}
	message.SetBody(bytes)
	sendMsg(dataPack, message)
}

// Default
//
//	@Description:
//	@param message
//	@param dataPack
func Default(message netx.IMessage, dataPack *netx.DataPack, sendMsg func(dataPack *netx.DataPack, message netx.IMessage)) {
	message.SetBody([]byte(errorx.UnknownCommandErr.Error()))
	sendMsg(dataPack, message)
}
