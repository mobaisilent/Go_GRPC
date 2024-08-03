package main

import (
	"context"
	"errors"
	"fmt"
	pb "grpc/hello-server/proto"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type server struct {
	pb.UnimplementedSayHelloServer
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing token")
	}
	var appId, appKey string
	if val, ok := md["appid"]; ok {
		appId = val[0]
	}
	if val, ok := md["appkey"]; ok {
		appKey = val[0]
	}
	if appId != "mobai" || appKey != "123123" {
		return nil, errors.New("invalid token")
	}
	// 上面是token认证

	fmt.Println("request name:", req.RequestName)
	return &pb.HelloResponse{
		ResponseMsg: "hello " + req.RequestName,
	}, nil
}

func main() {
	listen, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Printf("Failed to listen on port 9090: %v\n", err)
		return
	}
	grpcServer := grpc.NewServer()
	pb.RegisterSayHelloServer(grpcServer, &server{})
	err = grpcServer.Serve(listen)
	if err != nil {
		fmt.Printf("Failed to serve: %v\n", err)
	}
}
