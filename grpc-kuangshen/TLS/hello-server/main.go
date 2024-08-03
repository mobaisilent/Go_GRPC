package main

import (
	"context"
	"fmt"
	pb "grpc/hello-server/proto"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	// 自签证书和私钥文件    -- 主要改动代码在这里
	creds, err := credentials.NewServerTLSFromFile("../key/test.pem", "../key/test.key")
	if err != nil {
		log.Fatalf("Failed to generate credentials %v", err)
	}

	// 开启端口
	listen, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("Failed to listen on port 9090: %v", err)
	}

	// 创建grpc服务
	grpcServer := grpc.NewServer(grpc.Creds(creds)) // 传入证书

	// 在grpc服务上注册自己编写的服务
	pb.RegisterSayHelloServer(grpcServer, &server{})

	// 启动grpc服务
	err = grpcServer.Serve(listen)
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
