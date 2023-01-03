package main

import (
	evolvingserver "github.com/yuhao-jack/evolving-rpc/evolving-server"
	"github.com/yuhao-jack/evolving-rpc/evolving-server/svr_mgr"
	"github.com/yuhao-jack/evolving-rpc/model"
	"github.com/yuhao-jack/go-toolx/fun"
	"net/http"
)

func main() {
	serverConf := model.EvolvingServerConf{
		BindHost:   "0.0.0.0",
		ServerPort: 6601,
	}
	evolvingServer := evolvingserver.NewEvolvingServer(&serverConf)
	go evolvingServer.Start()
	handleMgr := HandleMgr{EvolvingServer: evolvingServer}
	http.HandleFunc("/serviceInfoList", handleMgr.ServiceInfoList)
	http.ListenAndServe("127.0.0.1:8000", nil)
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
