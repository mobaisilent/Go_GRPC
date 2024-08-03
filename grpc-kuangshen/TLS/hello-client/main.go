package main

import (
	"context"
	"fmt"
	"log"

	pb "grpc/hello-server/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	creds, err := credentials.NewClientTLSFromFile("../key/test.pem", "test.com")
	if err != nil {
		log.Fatalf("Failed to generate credentials %v", err)
	}

	// 连接到server端，此处启用安全传输，包括加密和验证
	conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	// go语言标准连接处理

	// 创建一个客户端
	client := pb.NewSayHelloClient(conn)

	resp, err := client.SayHello(context.Background(), &pb.HelloRequest{RequestName: "mobaiclient", Age: 18, Name: []string{"mobai", "mobai2"}})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	fmt.Println(resp.GetResponseMsg())
}
