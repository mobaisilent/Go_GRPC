# [Go gRPC进阶-超时设置（六） ](https://www.cnblogs.com/FireworksEasyCool/p/12702959.html)

### 前言[#](https://www.cnblogs.com/FireworksEasyCool/p/12702959.html#2563099071)

gRPC默认的请求的超时时间是很长的，当你没有设置请求超时时间时，所有在运行的请求都占用大量资源且可能运行很长的时间，导致服务资源损耗过高，使得后来的请求响应过慢，甚至会引起整个进程崩溃。

为了避免这种情况，我们的服务应该设置超时时间。前面的[入门教程](https://github.com/Bingjian-Zhu/go-grpc-example)提到，当客户端发起请求时候，需要传入上下文`context.Context`，用于结束`超时`或`取消`的请求。

本篇以[简单RPC](https://bingjian-zhu.github.io/2020/04/10/Go-gRPC教程-简单RPC（二）/)为例，介绍如何设置gRPC请求的超时时间。

### 客户端请求设置超时时间[#](https://www.cnblogs.com/FireworksEasyCool/p/12702959.html#2271942445)

修改调用服务端方法

1.把超时时间设置为当前时间+3秒

```go
Copy	clientDeadline := time.Now().Add(time.Duration(3 * time.Second))
	ctx, cancel := context.WithDeadline(ctx, clientDeadline)
	defer cancel()
```

2.响应错误检测中添加超时检测

```go
Copy       // 传入超时时间为3秒的ctx
	res, err := grpcClient.Route(ctx, &req)
	if err != nil {
		//获取错误状态
		statu, ok := status.FromError(err)
		if ok {
			//判断是否为调用超时
			if statu.Code() == codes.DeadlineExceeded {
				log.Fatalln("Route timeout!")
			}
		}
		log.Fatalf("Call Route err: %v", err)
	}
	// 打印返回值
	log.Println(res.Value)
```

完整的[client.go](https://github.com/Bingjian-Zhu/go-grpc-example/blob/master/6-grpc_deadlines/client/client.go)代码

### 服务端判断请求是否超时[#](https://www.cnblogs.com/FireworksEasyCool/p/12702959.html#619785162)

当请求超时后，服务端应该停止正在进行的操作，避免资源浪费。

```go
Copy// Route 实现Route方法
func (s *SimpleService) Route(ctx context.Context, req *pb.SimpleRequest) (*pb.SimpleResponse, error) {
	data := make(chan *pb.SimpleResponse, 1)
	go handle(ctx, req, data)
	select {
	case res := <-data:
		return res, nil
	case <-ctx.Done():
		return nil, status.Errorf(codes.Canceled, "Client cancelled, abandoning.")
	}
}

func handle(ctx context.Context, req *pb.SimpleRequest, data chan<- *pb.SimpleResponse) {
	select {
	case <-ctx.Done():
		log.Println(ctx.Err())
		runtime.Goexit() //超时后退出该Go协程
	case <-time.After(4 * time.Second): // 模拟耗时操作
		res := pb.SimpleResponse{
			Code:  200,
			Value: "hello " + req.Data,
		}
		// //修改数据库前进行超时判断
		// if ctx.Err() == context.Canceled{
		// 	...
		// 	//如果已经超时，则退出
		// }
		data <- &res
	}
}
```

一般地，在写库前进行超时检测，发现超时就停止工作。

完整[server.go](https://github.com/Bingjian-Zhu/go-grpc-example/tree/master/6-grpc_deadlines/server/server.go)代码

### 运行结果[#](https://www.cnblogs.com/FireworksEasyCool/p/12702959.html#869484100)

服务端：

```powershell
Copy:8000 net.Listing...
goroutine still running
```

客户端：

```powershell
CopyRoute timeout!
```

### 总结[#](https://www.cnblogs.com/FireworksEasyCool/p/12702959.html#2154918882)

超时时间的长短需要根据自身服务而定，例如返回一个`hello grpc`，可能只需要几十毫秒，然而处理大量数据的同步操作则可能要很长时间。需要考虑多方面因素来决定这个超时时间，例如系统间端到端的延时，哪些RPC是串行的，哪些是可以并行的等等。

教程源码地址：https://github.com/Bingjian-Zhu/go-grpc-example
参考：https://grpc.io/blog/deadlines/