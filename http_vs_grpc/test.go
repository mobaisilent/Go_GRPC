package main

import (
	"bytes"
	"context"
	student_service "dqq/micro_service/grpc"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const student_id = 10

func TestGrpc(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond) //连接超时设置为1000毫秒
	defer cancel()
	//连接到服务端
	conn, err := grpc.DialContext(
		ctx,
		"127.0.0.1:5678",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Printf("dial failed: %s", err)
		return
	}
	//创建client
	client := student_service.NewStudentServiceClient(conn)

	//准备好请求参数
	request := student_service.StudentID{Id: student_id}
	//发送请求，取得响应
	response, err := client.GetStudent(context.Background(), &request)
	if err != nil {
		fmt.Printf("get student failed: %s", err)
	} else {
		fmt.Println(response.Id)
	}
}

func TestHttp(t *testing.T) {
	client := http.Client{}

	//准备好请求参数
	sid := student_service.StudentID{Id: student_id}
	bs, err := sonic.Marshal(&sid)
	if err != nil {
		fmt.Printf("marshal request failed: %s\n", err)
		return
	}

	request, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:5679/", bytes.NewReader(bs))
	if err != nil {
		fmt.Printf("build request failed: %s\n", err)
		return
	}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Printf("http rpc failed: %s\n", err)
		return
	}
	bs, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("read response stream failed: %s\n", err)
		return
	}
	resp.Body.Close()

	var stu student_service.Student
	err = sonic.Unmarshal(bs, &stu)
	if err != nil {
		fmt.Printf("unmarshal student failed: %s\n", err)
		return
	}
	fmt.Println(stu.Id)
}

func BenchmarkGrpc(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond) //连接超时设置为1000毫秒
	defer cancel()
	//连接到服务端
	conn, err := grpc.DialContext(
		ctx,
		"127.0.0.1:5678",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Printf("dial failed: %s", err)
		return
	}
	//创建client
	client := student_service.NewStudentServiceClient(conn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//准备好请求参数
		request := student_service.StudentID{Id: student_id}
		//调用服务，取得结果，结果反序列化为结构体
		client.GetStudent(context.Background(), &request)
	}
}

func BenchmarkHttp(b *testing.B) {
	client := http.Client{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//准备好请求参数
		sid := student_service.StudentID{Id: student_id}
		bs, err := sonic.Marshal(&sid)
		if err != nil {
			fmt.Printf("marshal request failed: %s\n", err)
			return
		}

		request, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:5679/", bytes.NewReader(bs))
		if err != nil {
			fmt.Printf("build request failed: %s\n", err)
			return
		}
		//调用服务
		resp, err := client.Do(request)
		if err != nil {
			fmt.Printf("http rpc failed: %s\n", err)
			return
		}
		//取得结果
		bs, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("read response stream failed: %s\n", err)
			return
		}
		resp.Body.Close()
		//结果反序列化为结构体
		var stu student_service.Student
		sonic.Unmarshal(bs, &stu)
	}
}

// go test -v .\micro_service\grpc\client\ -run=^TestGrpc$ -count=1
// go test -v .\micro_service\grpc\client\ -run=^TestHttp$ -count=1
// go test .\micro_service\grpc\client\ -bench=^BenchmarkGrpc$ -run=^$ -count=1 -benchmem -benchtime=5s
// go test .\micro_service\grpc\client\ -bench=^BenchmarkHttp$ -run=^$ -count=1 -benchmem -benchtime=5s
