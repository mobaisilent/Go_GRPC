package main

import (
	"context"
	"fmt"
	"log"

	pb "grpc/hello-server/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 连接到server端，此处禁用安全传输，没有加密和验证
	conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
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
