# [Go gRPC教程-客户端流式RPC（四） ](https://www.cnblogs.com/FireworksEasyCool/p/12696733.html)

### 前言[#](https://www.cnblogs.com/FireworksEasyCool/p/12696733.html#478517132)

上一篇介绍了`服务端流式RPC`，客户端发送请求到服务器，拿到一个流去读取返回的消息序列。 客户端读取返回的流的数据。本篇将介绍`客户端流式RPC`。

`客户端流式RPC`：与`服务端流式RPC`相反，客户端不断的向服务端发送数据流，而在发送结束后，由服务端返回一个响应。

###### 情景模拟：客户端大量数据上传到服务端。

### 新建proto文件[#](https://www.cnblogs.com/FireworksEasyCool/p/12696733.html#1576715025)

新建client_stream.proto文件

1.定义发送信息

```protobuf
Copy// 定义流式请求信息
message StreamRequest{
    //流式请求参数
    string stream_data = 1;
}
```

2.定义接收信息

```protobuf
Copy// 定义响应信息
message SimpleResponse{
    //响应码
    int32 code = 1;
    //响应值
    string value = 2;
}
```

3.定义服务方法RouteList

客户端流式rpc，只要在请求的参数前添加stream即可

```protobuf
Copyservice StreamClient{
    // 客户端流式rpc，在请求的参数前添加stream
    rpc RouteList (stream StreamRequest) returns (SimpleResponse){};
}
```

4.编译proto文件

进入client_stream.proto所在目录，运行指令:

```
protoc --go_out=plugins=grpc:./ ./client_stream.proto
```

### 创建Server端[#](https://www.cnblogs.com/FireworksEasyCool/p/12696733.html#2783239147)

1.定义我们的服务，并实现RouteList方法

```go
Copy// SimpleService 定义我们的服务
type SimpleService struct{}
// RouteList 实现RouteList方法
func (s *SimpleService) RouteList(srv pb.StreamClient_RouteListServer) error {
	for {
		//从流中获取消息
		res, err := srv.Recv()
		if err == io.EOF {
			//发送结果，并关闭
			return srv.SendAndClose(&pb.SimpleResponse{Value: "ok"})
		}
		if err != nil {
			return err
		}
		log.Println(res.StreamData)
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
	pb.RegisterStreamClientServer(grpcServer, &SimpleService{})

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

### 创建Client端[#](https://www.cnblogs.com/FireworksEasyCool/p/12696733.html#2672603633)

1.创建调用服务端RouteList方法

```go
Copy// routeList 调用服务端RouteList方法
func routeList() {
	//调用服务端RouteList方法，获流
	stream, err := streamClient.RouteList(context.Background())
	if err != nil {
		log.Fatalf("Upload list err: %v", err)
	}
	for n := 0; n < 5; n++ {
		//向流中发送消息
		err := stream.Send(&pb.StreamRequest{StreamData: "stream client rpc " + strconv.Itoa(n)})
		if err != nil {
			log.Fatalf("stream request err: %v", err)
		}
	}
	//关闭流并获取返回的消息
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("RouteList get response err: %v", err)
	}
	log.Println(res)
}
```

2.启动gRPC客户端

```go
Copy// Address 连接地址
const Address string = ":8000"

var streamClient pb.StreamClientClient

func main() {
	// 连接服务器
	conn, err := grpc.Dial(Address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("net.Connect err: %v", err)
	}
	defer conn.Close()

	// 建立gRPC连接
	streamClient = pb.NewStreamClientClient(conn)
	routeList()
}
```

运行客户端

```powershell
Copygo run client.go
code:200 value:"hello grpc"
value:"ok"
```

服务端不断从客户端获取到数据

```powershell
Copystream client rpc 0
stream client rpc 1
stream client rpc 2
stream client rpc 3
stream client rpc 4
```

### 思考[#](https://www.cnblogs.com/FireworksEasyCool/p/12696733.html#587125539)

服务端在没有接受完消息时候能主动停止接收数据吗（很少有这种场景）？

答案：可以的，但是客户端代码需要注意EOF判断

1.我们把服务端的RouteList方法实现稍微修改，当接收到一条数据后马上调用SendAndClose()关闭stream.

```go
Copy// RouteList 实现RouteList方法
func (s *SimpleService) RouteList(srv pb.StreamClient_RouteListServer) error {
	for {
		//从流中获取消息
		res, err := srv.Recv()
		if err == io.EOF {
			//发送结果，并关闭
			return srv.SendAndClose(&pb.SimpleResponse{Value: "ok"})
		}
		if err != nil {
			return err
		}
		log.Println(res.StreamData)
		return srv.SendAndClose(&pb.SimpleResponse{Value: "ok"})
	}
}
```

2.再把客户端调用RouteList方法的实现稍作修改

```go
Copy// routeList 调用服务端RouteList方法
func routeList() {
	//调用服务端RouteList方法，获流
	stream, err := streamClient.RouteList(context.Background())
	if err != nil {
		log.Fatalf("Upload list err: %v", err)
	}
	for n := 0; n < 5; n++ {
		//向流中发送消息
		err := stream.Send(&pb.StreamRequest{StreamData: "stream client rpc " + strconv.Itoa(n)})
		//发送也要检测EOF，当服务端在消息没接收完前主动调用SendAndClose()关闭stream，此时客户端还执行Send()，则会返回EOF错误，所以这里需要加上io.EOF判断
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("stream request err: %v", err)
		}
	}
	//关闭流并获取返回的消息
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("RouteList get response err: %v", err)
	}
	log.Println(res)
}
```

客户端Send()需要检测err是否为EOF，因为当服务端在消息没接收完前主动调用SendAndClose()关闭stream，若此时客户端继续执行Send()，则会返回EOF错误。

### 总结[#](https://www.cnblogs.com/FireworksEasyCool/p/12696733.html#2617162232)

本篇介绍了`客户端流式RPC`的简单使用，下篇将介绍`双向流式RPC`。

教程源码地址：https://github.com/Bingjian-Zhu/go-grpc-example
参考：[gRPC官方文档中文版](http://doc.oschina.net/grpc?t=60133)