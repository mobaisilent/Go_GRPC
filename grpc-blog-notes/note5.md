# [Go gRPC教程-双向流式RPC（五） ](https://www.cnblogs.com/FireworksEasyCool/p/12698194.html)

### 前言[#](https://www.cnblogs.com/FireworksEasyCool/p/12698194.html#3039756210)

上一篇介绍了`客户端流式RPC`，客户端不断的向服务端发送数据流，在发送结束或流关闭后，由服务端返回一个响应。本篇将介绍`双向流式RPC`。

`双向流式RPC`：客户端和服务端双方使用读写流去发送一个消息序列，两个流独立操作，双方可以同时发送和同时接收。

###### 情景模拟：双方对话（可以一问一答、一问多答、多问一答，形式灵活）。

### 新建proto文件[#](https://www.cnblogs.com/FireworksEasyCool/p/12698194.html#3690201140)

新建both_stream.proto文件

1.定义发送信息

```protobuf
Copy// 定义流式请求信息
message StreamRequest{
    //流请求参数
    string question = 1;
}
```

2.定义接收信息

```protobuf
Copy// 定义流式响应信息
message StreamResponse{
    //流响应数据
    string answer = 1;
}
```

3.定义服务方法Conversations

双向流式rpc，只要在请求的参数前和响应参数前都添加stream即可

```protobuf
Copyservice Stream{
    // 双向流式rpc，同时在请求参数前和响应参数前加上stream
    rpc Conversations(stream StreamRequest) returns(stream StreamResponse){};
}
```

4.编译proto文件

进入both_stream.proto所在目录，运行指令:

```
protoc --go_out=plugins=grpc:./ ./both_stream.proto
```

### 创建Server端[#](https://www.cnblogs.com/FireworksEasyCool/p/12698194.html#172733478)

1.定义我们的服务，并实现RouteList方法

这里简单实现对话中一问一答的形式

```go
Copy// StreamService 定义我们的服务
type StreamService struct{}
// Conversations 实现Conversations方法
func (s *StreamService) Conversations(srv pb.Stream_ConversationsServer) error {
	n := 1
	for {
		req, err := srv.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		err = srv.Send(&pb.StreamResponse{
			Answer: "from stream server answer: the " + strconv.Itoa(n) + " question is " + req.Question,
		})
		if err != nil {
			return err
		}
		n++
		log.Printf("from stream client question: %s", req.Question)
	}
}
```

2.启动gRPC服务器

```go
Copyconst (
	// Address 监听地址
	Address string = ":8000"
	// Network 网络通信协议
	Network string = "tcp"
)

func main() {
	// 监听本地端口
	listener, err := net.Listen(Network, Address)
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
	}
	log.Println(Address + " net.Listing...")
	// 新建gRPC服务器实例
	grpcServer := grpc.NewServer()
	// 在gRPC服务器注册我们的服务
	pb.RegisterStreamServer(grpcServer, &StreamService{})

	//用服务器 Serve() 方法以及我们的端口信息区实现阻塞等待，直到进程被杀死或者 Stop() 被调用
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("grpcServer.Serve err: %v", err)
	}
}
```

运行服务端

```powershell
Copygo run server.go
:8000 net.Listing...
```

### 创建Client端[#](https://www.cnblogs.com/FireworksEasyCool/p/12698194.html#2647767188)

1.创建调用服务端Conversations方法

```go
Copy// conversations 调用服务端的Conversations方法
func conversations() {
	//调用服务端的Conversations方法，获取流
	stream, err := streamClient.Conversations(context.Background())
	if err != nil {
		log.Fatalf("get conversations stream err: %v", err)
	}
	for n := 0; n < 5; n++ {
		err := stream.Send(&pb.StreamRequest{Question: "stream client rpc " + strconv.Itoa(n)})
		if err != nil {
			log.Fatalf("stream request err: %v", err)
		}
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Conversations get stream err: %v", err)
		}
		// 打印返回值
		log.Println(res.Answer)
	}
	//最后关闭流
	err = stream.CloseSend()
	if err != nil {
		log.Fatalf("Conversations close stream err: %v", err)
	}
}
```

2.启动gRPC客户端

```go
Copy// Address 连接地址
const Address string = ":8000"

var streamClient pb.StreamClient

func main() {
	// 连接服务器
	conn, err := grpc.Dial(Address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("net.Connect err: %v", err)
	}
	defer conn.Close()

	// 建立gRPC连接
	streamClient = pb.NewStreamClient(conn)
	conversations()
}
```

运行客户端，获取到服务端的应答

```powershell
Copygo run client.go
from stream server answer: the 1 question is stream client rpc 0
from stream server answer: the 2 question is stream client rpc 1
from stream server answer: the 3 question is stream client rpc 2
from stream server answer: the 4 question is stream client rpc 3
from stream server answer: the 5 question is stream client rpc 4
```

服务端获取到来自客户端的提问

```powershell
Copyfrom stream client question: stream client rpc 0
from stream client question: stream client rpc 1
from stream client question: stream client rpc 2
from stream client question: stream client rpc 3
from stream client question: stream client rpc 4
```

### 总结[#](https://www.cnblogs.com/FireworksEasyCool/p/12698194.html#1544572755)

本篇介绍了`双向流式RPC`的简单使用。

教程源码地址：https://github.com/Bingjian-Zhu/go-grpc-example
参考：[gRPC官方文档中文版](http://doc.oschina.net/grpc?t=60133)