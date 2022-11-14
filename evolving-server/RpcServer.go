package evolving_server

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitee.com/yuhao-jack/evolving-rpc/contents"
	evolvingclient "gitee.com/yuhao-jack/evolving-rpc/evolving-client"
	"gitee.com/yuhao-jack/evolving-rpc/model"

	"gitee.com/yuhao-jack/go-toolx/fun"
	"gitee.com/yuhao-jack/go-toolx/netx"
	"go/token"
	"log"
	"reflect"
	"strings"
	"sync"
)

var unknownProtocErr = errors.New("error: unknown protoc")

type IRpcServer interface {
	Register(rcvr any) error
}
type methodType struct {
	sync.Mutex
	method    reflect.Method
	ReqType   reflect.Type
	ReplyType reflect.Type
}

type service struct {
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
}

type RpcServer struct {
	serviceMap           map[string]*service
	registerCenterConfig *model.EvolvingClientConfig
	serverConfig         *model.ServiceInfo
	evolvingServer       *EvolvingServer
}

func NewRpcServer(registerCenterConfig *model.EvolvingClientConfig, serverConfig *model.ServiceInfo) *RpcServer {
	rpcServer := RpcServer{
		serviceMap:           map[string]*service{},
		registerCenterConfig: registerCenterConfig,
		serverConfig:         serverConfig,
	}
	evolvingClient := evolvingclient.NewEvolvingClient(registerCenterConfig)
	if err := evolvingClient.RegisterService(serverConfig, func(reply netx.IMessage) {
		log.Default().Println(string(reply.GetBody()))
	}); err != nil {
		panic("register service to register-center failed ,err:" + err.Error())
	}
	rpcServer.evolvingServer = NewEvolvingServer(&model.EvolvingServerConf{
		BindHost:   serverConfig.ServiceHost,
		ServerPort: serverConfig.ServicePort,
	})
	return &rpcServer
}

func (r *RpcServer) Register(rcvr any) error {
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
	r.serviceMap[s.name] = s
	return nil
}

func (r *RpcServer) Run() {
	for n, server := range r.serviceMap {
		for s := range server.method {
			r.evolvingServer.SetCommand(fmt.Sprint(n, ".", s), func(dataPack *netx.DataPack, reply netx.IMessage) {
				splitArr := strings.Split(string(reply.GetCommand()), ".")
				var reqv reflect.Value
				ts := r.serviceMap[splitArr[0]]
				tm := ts.method[splitArr[1]]
				reqv = reflect.New(tm.ReqType)

				var err error
				var command = string(reply.GetProtoc())
				switch command {
				case contents.Json:
					err = json.Unmarshal(reply.GetBody(), reqv.Interface())
				default:
					err = unknownProtocErr
				}
				if err != nil {
					reply.SetBody([]byte(err.Error()))
					r.evolvingServer.Execute(dataPack, reply, nil)
					return
				}

				res := tm.method.Func.Call([]reflect.Value{ts.rcvr, reflect.Indirect(reqv)})[0].Interface()
				if res != nil {
					var bytes []byte
					switch command {
					case contents.Json:
						bytes, err = json.Marshal(res)
					default:
						err = unknownProtocErr
					}

					if err != nil {
						reply.SetBody([]byte(err.Error()))
					} else {
						reply.SetBody(bytes)
					}
				}
				r.evolvingServer.Execute(dataPack, reply, nil)
			})
		}

	}
	r.evolvingServer.Start()
}

func buildMethodMap(s *service) {
	for m := 0; m < s.typ.NumMethod(); m++ {
		method := s.typ.Method(m)
		if !method.IsExported() {
			continue
		}
		if method.Type.NumIn() != 2 {
			continue
		}
		reqType := method.Type.In(1)
		if !isExportedOrBuiltinType(reqType) {
			continue
		}

		if method.Type.NumOut() != 1 {
			continue
		}
		replyType := method.Type.Out(0) // must be a pointer.
		if !isExportedOrBuiltinType(replyType) {
			continue
		}

		s.method[method.Name] = &methodType{Mutex: sync.Mutex{}, method: method, ReqType: reqType, ReplyType: replyType}
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return token.IsExported(t.Name()) || t.PkgPath() == ""
}
