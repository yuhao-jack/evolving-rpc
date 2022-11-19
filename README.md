### 项目说明
> evolving-rpc是一个以学习为主要目标的简易RPC框架，该框架由作者从底层的TCP封装开始一点点修改出来的，由于作者水平有限，项目多有瑕疵，但整体上实现了RPC的基本功能，这也让作者在技术领域从一个简单的只会背诵八股文的`API调用工程师`慢慢提升到理解并实现出来的程序员。
> 该框架作者只在测试环境中进行过简单的验证，还不是一个成熟的可投入生产使用的框架，该项目目前仍在持续的优化中，诸多问题读者可在issue中提出，也可以Ｅmail：`154826195@qq.com`联系本人．

### 压测数据
```
goos: linux
goarch: amd64
pkg: github.com/yuhao-jack/evolving-rpc/test
cpu: Intel(R) Core(TM) i5-10400 CPU @ 2.90GHz
BenchmarkDirectly
2022/11/18 10:22:20.988715 /home/yuhao/go/src/evolving-rpc/evolving-client/EvolvingClient.go:104: start evolving-client successful.
2022/11/18 10:22:20.991760 /home/yuhao/go/src/evolving-rpc/evolving-client/EvolvingClient.go:104: start evolving-client successful.
2022/11/18 10:22:21.022299 /home/yuhao/go/src/evolving-rpc/evolving-client/EvolvingClient.go:104: start evolving-client successful.
BenchmarkDirectly-2   	    4312	    318592 ns/op
```
> 从压测数据上来看性能还是不错的　＿＾－－＾＿
### 快速开始　
####    安装
　　```go　get　github.com/yuhao-jack/evolving-rpc```

####    [点我查看直连模式](./test/directly_rpc_test.go)

####    [点我查看分布式模式（还在完善中）](./test/distributed_rpc_test.go) 

### 注意
作者在写该项目时是为了提升自己，完全不想引入第三方库，所以默认使用的`json`作为传输协议，在后续的版本中为了提升性能可能考虑引入`protobuf`作为传输协议
