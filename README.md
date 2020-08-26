
# 介绍
这是一个基于websocket实现的双向通信多路复用库。

- 双向通信：区别于传统rpc式,client主动请求server，server回复。通信两端皆可以主动发起请求。  
- 多路复用：在一个通信链路上，可以传输多种请求.  

# 使用场景  

```
    内网IP-------NAT--------公网IP
```
在内网与公网通信时，只能由内网端发起连接到公网端。  
传统的rpc方式，也是由内网端直接请求公网端,公网端进行回复。  

对于公网端主动请求内网端的情况如何处理？  
首先，内网端建立连接到公网端，并保持这个长连接。  
然后，公网端复用这个连接，发起请求到内网端。  

bimux是为解决上面问题而产生的

# 使用说明
muxer 一个可以被复用的连接  
```
type Muxer interface {
	Send(route int32, data []byte) error
	Rpc(route int32, req []byte, timeout time.Duration) (rsp []byte, err error)
	SetID(id string)
	GetID() string

	//raw read and write
	ReadPacket() (data []byte, err error)
	WritePacket(data []byte) error

	Wait() error
	Close()
}
```

基于websockert的server与client实现,
当连接建立后，server与client的身份就不在固定。其中的参数的回调函数，是它作为被访问方时的处理函数 
``` 
func WsServe(addr, uri string,
	connections chan Muxer,
	rpcServeHook RpcServeFunc,
	onewayServeHook OnewayFunc,
	closeHook CloseFunc,
)  error

func WsDial(addr, uri string,
	rpcServeHook RpcServeFunc,
	onewayServeHook OnewayFunc,
	closeHook CloseFunc,
) (Muxer, error)

```

回调函数及定义  
参数 route：调用者用来区分不同的数据类型进行处理。在通信两端route值不能重复。  
```
rpcServeHook RpcServeFunc,
onewayServeHook OnewayFunc,
stopHook StopFunc,

type RpcServeFunc func(route int32, req []byte, m Muxer) []byte
type OnewayFunc func(route int32, req []byte, m Muxer)
type CloseFunc func(Muxer)
```

# 内部实现

## 协议 

### websocket 
使用websocket主要是原因是ws在tcp层上做了帧定义,不需要再自己实现。如下图所示：  

![ws-data-frame-format.png](./ws-data-frame-format.png)

在ws的payload中，我们定义了自己的消息格式  

### 消息格式
使用protobuf定义    
```
enum Flag {
    request = 0; 
    response = 1; 
    oneway = 2;
};

message Message {
    uint64 number = 1;
    Flag flag = 2;

    int32 route = 3;
    bytes data = 4;
}
```
flag用来区分是rpc消息，还是单向不需要回复的消息。
number用来区分消息，将关联的消息对应  

## 流程说明
rpc
```
client                                  server
rpc-------msg(1,FlagRequest)---------->rpcServeHook
 ^                                        |
 |--------msg(2,FlagResponse)-------------|
```
send
```
client                                  server
send--------msg(1,FlagOneway)---------->onewayHook
```

# 性能测试  
```
go build example/client.go
go build example/server.go
```
mac测试结果并发能到4w左右，具体如下：    
需要安装测试工具hey：go get -u github.com/rakyll/hey  
```
./server 
./client 
hey -c 100 -n 1000000 -t 0 http://127.0.0.1:12000/mux

Summary:
  Total:	25.0644 secs
  Slowest:	0.0670 secs
  Fastest:	0.0002 secs
  Average:	0.0025 secs
  Requests/sec:	39897.1667

  Total data:	10000000 bytes
  Size/request:	10 bytes

Response time histogram:
  0.000 [1]	|
  0.007 [980473]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.014 [19032]	|■
  0.020 [326]	|
  0.027 [85]	|
  0.034 [53]	|
  0.040 [22]	|
  0.047 [6]	|
  0.054 [0]	|
  0.060 [1]	|
  0.067 [1]	|


Latency distribution:
  10% in 0.0015 secs
  25% in 0.0018 secs
  50% in 0.0022 secs
  75% in 0.0028 secs
  90% in 0.0037 secs
  95% in 0.0048 secs
  99% in 0.0078 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0000 secs, 0.0002 secs, 0.0670 secs
  DNS-lookup:	0.0000 secs, 0.0000 secs, 0.0000 secs
  req write:	0.0000 secs, 0.0000 secs, 0.0250 secs
  resp wait:	0.0024 secs, 0.0001 secs, 0.0458 secs
  resp read:	0.0001 secs, 0.0000 secs, 0.0567 secs

Status code distribution:
  [200]	1000000 responses
```


# 测试
GO111MODULE=on go get github.com/golang/mock/mockgen@v1.4.3
mockgen -source conn.go  -package bimux -destination conn_mock.go
