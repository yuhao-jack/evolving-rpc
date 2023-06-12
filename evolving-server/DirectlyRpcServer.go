package evolving_server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yuhao-jack/evolving-rpc/contents"
	"github.com/yuhao-jack/evolving-rpc/model"
	"github.com/yuhao-jack/go-toolx/containerx"
	"github.com/yuhao-jack/go-toolx/fun"
	"github.com/yuhao-jack/go-toolx/netx"
	"reflect"
	"strings"
)

// DirectlyRpcServerConfig
// @Description: 直连模式下的RPC的服务端的配置
type DirectlyRpcServerConfig struct {
	model.EvolvingServerConf
}

// DirectlyRpcServer
// @Description: 直连模式下的RPC服务端
type DirectlyRpcServer struct {
	serviceMap                map[string]*service
	evolvingServer            *EvolvingServer
	protocUnmarshalHandlerMap *containerx.ConcurrentMap[string, func(in []byte, recv any) error]
	protocMarshalHandlerMap   *containerx.ConcurrentMap[string, func(recv any) ([]byte, error)]
}

// NewDirectlyRpcServer
//
//	@Description: 创建一个直连模式下的RPC服务端
//	@param config 直连模式下的RPC的服务端的配置
//	@return *DirectlyRpcServer 直连模式下的RPC服务端
func NewDirectlyRpcServer(config *DirectlyRpcServerConfig) *DirectlyRpcServer {
	d := &DirectlyRpcServer{evolvingServer: NewEvolvingServer(&config.EvolvingServerConf), serviceMap: map[string]*service{}}
	d.protocUnmarshalHandlerMap = containerx.NewConcurrentMap[string, func(in []byte, recv any) error]()
	d.protocMarshalHandlerMap = containerx.NewConcurrentMap[string, func(recv any) ([]byte, error)]()
	d.SetProtocUnmarshalHandler(contents.Json, func(in []byte, recv any) error {
		return json.Unmarshal(in, recv)
	})
	d.SetProtocMarshalHandler(contents.Json, func(recv any) ([]byte, error) {
		return json.Marshal(recv)
	})
	return d
}

// Register
//
//	@Description: 注册服务
//	@receiver d
//	@param rcvr 具体服务对象的指针
//	@return error 注册失败时的错误信息
func (d *DirectlyRpcServer) Register(rcvr any) error {
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()

	if fun.IsBlank(s.name) {
		return errors.New("no service name for type " + s.typ.String())
	}
	s.method = make(map[string]*methodType)
	buildMethodMap(s)
	if len(s.method) == 0 {
		return errors.New(s.name + " has no exported methods of suitable type")
	}
	d.serviceMap[s.name] = s
	return nil
}

// Run
//
//	@Description: 直连模式下的RPC服务端的启动（该方法阻塞）
//	@receiver d
func (d *DirectlyRpcServer) Run() {
	for n, server := range d.serviceMap {
		for s := range server.method {
			d.evolvingServer.SetCommand(fmt.Sprint(n, ".", s), func(dataPack *netx.DataPack, reply netx.IMessage) {
				splitArr := strings.Split(string(reply.GetCommand()), ".")
				var reqv reflect.Value
				ts := d.serviceMap[splitArr[0]]
				tm := ts.method[splitArr[1]]
				reqv = reflect.New(tm.ReqType)

				var err error
				var protoc = string(reply.GetProtoc())

				unmarshalHandler, b := d.protocUnmarshalHandlerMap.Get(protoc)
				if b {
					err = unmarshalHandler(reply.GetBody(), reqv.Interface())
				} else {
					err = unknownProtocErr
				}
				if err != nil {
					reply.SetBody([]byte(err.Error()))
					d.evolvingServer.Execute(dataPack, reply, nil)
					return
				}

				res := tm.method.Func.Call([]reflect.Value{ts.rcvr, reflect.Indirect(reqv)})[0].Interface()
				if res != nil {
					var bytes []byte
					marshalHandler, b := d.protocMarshalHandlerMap.Get(protoc)
					if b {
						bytes, err = marshalHandler(res)
					} else {
						err = unknownProtocErr
					}
					if err != nil {
						reply.SetBody([]byte(err.Error()))
					} else {
						reply.SetBody(bytes)
					}
				}
				d.evolvingServer.Execute(dataPack, reply, nil)
			})
		}
	}
	d.evolvingServer.Start()
}

func (d *DirectlyRpcServer) SetProtocUnmarshalHandler(protoc string, handler func(in []byte, recv any) error) {
	d.protocUnmarshalHandlerMap.Set(protoc, handler)
	_, b := d.protocMarshalHandlerMap.Get(protoc)
	if !b {
		contents.RpcLogger.Warn("WARNING %s MarshalHandler is empty.", protoc)
	} else {
		contents.RpcLogger.Info("%s protoc both MarshalHandler and UnmarshalHandler are ready.")
	}
}

func (d *DirectlyRpcServer) SetProtocMarshalHandler(protoc string, handler func(recv any) ([]byte, error)) {
	d.protocMarshalHandlerMap.Set(protoc, handler)
	_, b := d.protocUnmarshalHandlerMap.Get(protoc)
	if !b {
		contents.RpcLogger.Warn("WARNING %s UnmarshalHandler is empty.", protoc)
	} else {
		contents.RpcLogger.Info("%s protoc both MarshalHandler and UnmarshalHandler are ready.")
	}
}
