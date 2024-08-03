package main

import (
	"context"
	"fmt"
	"log"

	pb "grpc/hello-server/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
type PerRPCCredentials interface {
	GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error)
	RequireTransportSecurity() bool
}
*/

type ClientTokenAuth struct{}

func (c ClientTokenAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"appId":  "mobaia",
		"appKey": "123123",
	}, nil
}

func (c ClientTokenAuth) RequireTransportSecurity() bool {
	return false
}

func main() {
	// 连接到server端，此处禁用安全传输，没有加密和验证
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	opts = append(opts, grpc.WithPerRPCCredentials(new(ClientTokenAuth)))
	conn, err := grpc.Dial("localhost:9090", opts...)
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

/*
// 证书认证
creds, err := credentials.NewClientTLSFromFile("path/to/cert.pem", "")
if err != nil {
    log.Fatalf("Failed to generate credentials %v", err)
}

var opts []grpc.DialOption
opts = append(opts, grpc.WithTransportCredentials(creds))
opts = append(opts, grpc.WithPerRPCCredentials(new(ClientTokenAuth)))

conn, err := grpc.Dial("localhost:9090", opts...)
if err != nil {
    log.Fatalf("did not connect: %v", err)
}
*/
