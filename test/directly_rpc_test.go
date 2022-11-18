package test

import (
	"encoding/json"
	evolving_client "github.com/yuhao-jack/evolving-rpc/evolving-client"
	evolving_server "github.com/yuhao-jack/evolving-rpc/evolving-server"
	"github.com/yuhao-jack/evolving-rpc/model"
	"log"
	"testing"
	"time"
)

// Arith
// @Description:
type Arith struct {
}

// ArithReq
// @Description:
type ArithReq struct {
	A int
	B int
}

// ArithReply
// @Description:
type ArithReply struct {
	Pro int // *
	Quo int // /
	Rem int // %
}

func (a *Arith) Multiply(req *ArithReq) (reply ArithReply) {
	reply.Pro = req.A * req.B
	return
}

func (a *Arith) Divide(req *ArithReq) (reply ArithReply) {
	if req.B == 0 {
		panic("divide by zero")
	}
	reply.Quo = req.A / req.B
	reply.Rem = req.A % req.B
	return
}

func init() {
	config := evolving_server.DirectlyRpcServerConfig{EvolvingServerConf: model.EvolvingServerConf{
		BindHost:   "0.0.0.0",
		ServerPort: 3301,
	}}
	server := evolving_server.NewDirectlyRpcServer(&config)
	err := server.Register(new(Arith))
	if err != nil {
		log.Default().Println("err:", err)
		return
	}
	go server.Run()
	time.Sleep(3 * time.Second)
}

func TestDirectly(t *testing.T) {
	config := evolving_client.DirectlyRpcClientConfig{EvolvingClientConfig: model.EvolvingClientConfig{
		EvolvingServerHost: "0.0.0.0",
		EvolvingServerPort: 3301,
		HeartbeatInterval:  5 * time.Minute,
	}}
	client := evolving_client.NewDirectlyRpcClient(&config)
	bytes, _ := json.Marshal(&ArithReq{
		A: 99,
		B: 63,
	})

	res, err := client.ExecuteCommand("Arith.Divide", bytes, true)
	log.Default().Println(string(res), err)
	res, err = client.ExecuteCommand("Arith.Multiply", bytes, true)
	log.Default().Println(string(res), err)

}

func BenchmarkDirectly(b *testing.B) {
	config := evolving_client.DirectlyRpcClientConfig{EvolvingClientConfig: model.EvolvingClientConfig{
		EvolvingServerHost: "0.0.0.0",
		EvolvingServerPort: 3301,
		HeartbeatInterval:  5 * time.Minute,
	}}
	client := evolving_client.NewDirectlyRpcClient(&config)
	bytes, _ := json.Marshal(&ArithReq{
		A: 99,
		B: 63,
	})

	for i := 0; i < b.N; i++ {
		//res, err := client.ExecuteCommand("Arith.Divide", bytes, true)
		//log.Default().Println(string(res), err)
		//res, err = client.ExecuteCommand("Arith.Multiply", bytes, true)
		//log.Default().Println(string(res), err)
		client.ExecuteCommand("Arith.Divide", bytes, true)
		client.ExecuteCommand("Arith.Multiply", bytes, true)
	}
}
