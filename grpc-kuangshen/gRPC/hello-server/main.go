package main

import (
	"context"
	"fmt"
	pb "grpc/hello-server/proto"
	"net"

	"google.golang.org/grpc"
)

// 前面的grpc是module名字，后面的hello-server是目录名字 就是定义时 go mod init grpc/hello-server

// hello server
type server struct {
	pb.UnimplementedSayHelloServer
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	fmt.Println("request name:", req.RequestName)
	return &pb.HelloResponse{
		ResponseMsg: "hello " + req.RequestName,
	}, nil
}

func main() {
	// 开启端口
	listen, _ := net.Listen("tcp", ":9090")
	// 创建grpc服务
	grpcServer := grpc.NewServer()
	// 在grpc服务上注册自己编写的服务
	pb.RegisterSayHelloServer(grpcServer, &server{})

	// 启动grpc服务
	err := grpcServer.Serve(listen)
	if err != nil {
		fmt.Printf("failed to serve:%v\n", err)
		return
	}
}
