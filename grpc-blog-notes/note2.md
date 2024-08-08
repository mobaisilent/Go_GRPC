# [Go gRPC教程-简单RPC（二） ](https://www.cnblogs.com/FireworksEasyCool/p/12674120.html)

### 前言[#](https://www.cnblogs.com/FireworksEasyCool/p/12674120.html#621991211)

gRPC主要有4种请求和响应模式，分别是`简单模式(Simple RPC)`、`服务端流式（Server-side streaming RPC）`、`客户端流式（Client-side streaming RPC）`、和`双向流式（Bidirectional streaming RPC）`。

- `简单模式(Simple RPC)`：客户端发起请求并等待服务端响应。
- `服务端流式（Server-side streaming RPC）`：客户端发送请求到服务器，拿到一个流去读取返回的消息序列。 客户端读取返回的流，直到里面没有任何消息。
- `客户端流式（Client-side streaming RPC）`：与服务端数据流模式相反，这次是客户端源源不断的向服务端发送数据流，而在发送结束后，由服务端返回一个响应。
- `双向流式（Bidirectional streaming RPC）`：双方使用读写流去发送一个消息序列，两个流独立操作，双方可以同时发送和同时接收。

本篇文章先介绍简单模式。

### 新建proto文件[#](https://www.cnblogs.com/FireworksEasyCool/p/12674120.html#915362889)

主要是定义我们服务的方法以及数据格式，我们使用上一篇的simple.proto文件。

1.定义发送消息的信息

```protobuf
Copymessage SimpleRequest{
    // 定义发送的参数，采用驼峰命名方式，小写加下划线，如：student_name
    string data = 1;//发送数据
}
```

2.定义响应信息

```protobuf
Copymessage SimpleResponse{
    // 定义接收的参数
    // 参数类型 参数名 标识号(不可重复)
    int32 code = 1;  //状态码
    string value = 2;//接收值
}
```

3.定义服务方法Route

```protobuf
Copy// 定义我们的服务（可定义多个服务,每个服务可定义多个接口）
service Simple{
    rpc Route (SimpleRequest) returns (SimpleResponse){};
}
```

4.编译proto文件

我这里使用上一篇介绍的VSCode-proto3插件，保存后自动编译。

> 指令编译方法，进入simple.proto文件所在目录，运行：
> `protoc --go_out=plugins=grpc:./ ./simple.proto`

### 创建Server端[#](https://www.cnblogs.com/FireworksEasyCool/p/12674120.html#3156337349)

1.定义我们的服务，并实现Route方法

```go
Copyimport (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "go-grpc-example/proto"
)
// SimpleService 定义我们的服务
type SimpleService struct{}

// Route 实现Route方法
func (s *SimpleService) Route(ctx context.Context, req *pb.SimpleRequest) (*pb.SimpleResponse, error) {
	res := pb.SimpleResponse{
		Code:  200,
		Value: "hello " + req.Data,
	}
	return &res, nil
}
```

该方法需要传入RPC的上下文context.Context，它的作用结束`超时`或`取消`的请求。更具体的说请参考[该文章](https://blog.csdn.net/chinawangfei/article/details/86559975)

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
	pb.RegisterSimpleServer(grpcServer, &SimpleService{})

	//用服务器 Serve() 方法以及我们的端口信息区实现阻塞等待，直到进程被杀死或者 Stop() 被调用
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("grpcServer.Serve err: %v", err)
	}
}
```

里面每个方法的作用都有注释，这里就不解析了。
运行服务端

```powershell
Copygo run server.go
:8000 net.Listing...
```

### 创建Client端[#](https://www.cnblogs.com/FireworksEasyCool/p/12674120.html#4226951648)

```go
Copyimport (
	"context"
	"log"

	"google.golang.org/grpc"

	pb "go-grpc-example/proto"
)
const (
	// Address 连接地址
	Address string = ":8000"
)

func main() {
	// 连接服务器
	conn, err := grpc.Dial(Address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("net.Connect err: %v", err)
	}
	defer conn.Close()

	// 建立gRPC连接
	grpcClient := pb.NewSimpleClient(conn)
	// 创建发送结构体
	req := pb.SimpleRequest{
		Data: "grpc",
	}
	// 调用我们的服务(Route方法)
	// 同时传入了一个 context.Context ，在有需要时可以让我们改变RPC的行为，比如超时/取消一个正在运行的RPC
	res, err := grpcClient.Route(context.Background(), &req)
	if err != nil {
		log.Fatalf("Call Route err: %v", err)
	}
	// 打印返回值
	log.Println(res)
}
```

运行客户端

```powershell
Copygo run client.go
code:200 value:"hello grpc"
```

成功调用Server端的Route方法并获取返回的数据。

### 总结[#](https://www.cnblogs.com/FireworksEasyCool/p/12674120.html#3281195411)

本篇介绍了简单RPC模式，客户端发起请求并等待服务端响应。下篇将介绍`服务端流式RPC`.

教程源码地址：https://github.com/Bingjian-Zhu/go-grpc-example
参考：[gRPC官方文档中文版](http://doc.oschina.net/grpc?t=60133)