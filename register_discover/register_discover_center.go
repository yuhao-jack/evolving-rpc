package main

import (
	"flag"
	"fmt"
	evolvingserver "github.com/yuhao-jack/evolving-rpc/evolving-server"
	"github.com/yuhao-jack/evolving-rpc/evolving-server/svr_mgr"
	"github.com/yuhao-jack/evolving-rpc/model"
	go_log "github.com/yuhao-jack/go-log"
	"github.com/yuhao-jack/go-toolx/fun"
	"net/http"
	"os"
)

var logger = go_log.DefaultGoLog()

func main() {
	defer logger.Destroy()
	serverConf := model.EvolvingServerConf{}
	var (
		host         string
		port, rdport int
	)
	ip, err := fun.GetLocalIp()
	if err != nil {
		logger.Error("GetLocalIp err:%v", err)
		os.Exit(1)
	}
	ip = ""
	flag.StringVar(&serverConf.BindHost, "rdh", ip, "注册发现服务的host")
	flag.IntVar(&rdport, "rdp", 6601, "注册发现服务的端口")
	serverConf.ServerPort = int32(rdport)
	flag.StringVar(&host, "h", ip, "工具服务host")
	flag.IntVar(&port, "p", 8080, "工具服务端口")
	flag.Parse()
	logger.Info("register and discover center addr:%s:%d", serverConf.BindHost, serverConf.ServerPort)
	logger.Info("tools service addr:%s:%d", host, port)

	evolvingServer := evolvingserver.NewEvolvingServer(&serverConf)
	go evolvingServer.Start()
	handleMgr := HandleMgr{EvolvingServer: evolvingServer}
	http.HandleFunc("/serviceInfoList", handleMgr.ServiceInfoList)
	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
}

type HandleMgr struct {
	EvolvingServer *evolvingserver.EvolvingServer
}

func (h *HandleMgr) ServiceInfoList(w http.ResponseWriter, r *http.Request) {

	if svr_mgr.GetServiceMgrInstance().ServiceInfoList.Size() > 0 {
		w.Write([]byte(fun.StrVal(map[string]any{"msg": "success", "data": svr_mgr.GetServiceMgrInstance().ServiceInfoList.Elements(), "code": 0})))

	} else {
		w.Write([]byte(fun.StrVal(map[string]any{"msg": "no data", "code": 10001})))
	}
}
