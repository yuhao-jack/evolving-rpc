package test

import (
	"encoding/json"
	"fmt"
	evolving_client "github.com/yuhao-jack/evolving-rpc/evolving-client"
	evolving_server "github.com/yuhao-jack/evolving-rpc/evolving-server"
	"github.com/yuhao-jack/evolving-rpc/model"
	"log"
	"math/rand"
	"sync/atomic"
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

type EchoReq struct {
	Key       []byte
	Val       []byte
	Topic     string
	Partition int32
	Timestamp time.Time
}

type EchoReply struct {
	Partition int32
	Offset    int64
	Timestamp time.Time
}

var off int64

func (a *Arith) Echo(req *EchoReq) (reply *EchoReply) {
	reply = &EchoReply{}
	reply.Timestamp = time.Now()
	if req.Key == nil || len(req.Key) == 0 {
		reply.Partition = int32(len(req.Val) % 99)
	}
	reply.Offset = off
	atomic.AddInt64(&off, 1)
	return
}

func beforeTestDirectlyRpc() {
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

func TestDirectlyRpc(t *testing.T) {
	beforeTestDirectlyRpc()
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

func TestSendMsg(t *testing.T) {
	beforeTestDirectlyRpc()
	config := evolving_client.DirectlyRpcClientConfig{EvolvingClientConfig: model.EvolvingClientConfig{
		EvolvingServerHost: "0.0.0.0",
		EvolvingServerPort: 3301,
		HeartbeatInterval:  5 * time.Minute,
	}}
	client := evolving_client.NewDirectlyRpcClient(&config)
	rand.Seed(time.Now().Unix())
	for i := 0; i < 1000; i++ {
		partition := i % 3
		echoReq := &EchoReq{
			Val:       []byte(fmt.Sprintf("val-%d", i*i)),
			Topic:     "go-test",
			Partition: int32(partition),
			Timestamp: time.Now(),
		}
		if rand.Intn(100) > 50 {
			echoReq.Key = []byte(fmt.Sprintf("key-%d", i))
		}
		bytes, _ := json.Marshal(echoReq)
		res, err := client.ExecuteCommand("Arith.Echo", bytes, true)
		fmt.Println(string(res), err)

	}

}
func RandInt(min, max int) int {
	//once.Do(func() {

	//})
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return rand.Intn(max-min) + min
}
func BenchmarkDirectlyRpc(b *testing.B) {
	beforeTestDirectlyRpc()
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
