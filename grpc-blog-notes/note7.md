# [Go gRPC进阶-TLS认证+自定义方法认证（七） ](https://www.cnblogs.com/FireworksEasyCool/p/12710325.html)

### 前言[#](https://www.cnblogs.com/FireworksEasyCool/p/12710325.html#453068258)

前面篇章的gRPC都是明文传输的，容易被篡改数据。本章将介绍如何为gRPC添加安全机制，包括TLS证书认证和Token认证。

### TLS证书认证[#](https://www.cnblogs.com/FireworksEasyCool/p/12710325.html#211646506)

###### 什么是TLS

TLS（Transport Layer Security，安全传输层)，TLS是建立在`传输层`TCP协议之上的协议，服务于应用层，它的前身是SSL（Secure Socket Layer，安全套接字层），它实现了将应用层的报文进行加密后再交由TCP进行传输的功能。

###### TLS的作用

TLS协议主要解决如下三个网络安全问题。

- 保密(message privacy)，保密通过加密encryption实现，所有信息都加密传输，第三方无法嗅探；
- 完整性(message integrity)，通过MAC校验机制，一旦被篡改，通信双方会立刻发现；
- 认证(mutual authentication)，双方认证,双方都可以配备证书，防止身份被冒充；

#### 生成私钥

生成RSA私钥：`openssl genrsa -out server.key 2048`

> 生成RSA私钥，命令的最后一个参数，将指定生成密钥的位数，如果没有指定，默认512

生成ECC私钥：`openssl ecparam -genkey -name secp384r1 -out server.key`

> 生成ECC私钥，命令为椭圆曲线密钥参数生成及操作，本文中ECC曲线选择的是secp384r1

#### 生成公钥

```
openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
```

> openssl req：生成自签名证书，-new指生成证书请求、-sha256指使用sha256加密、-key指定私钥文件、-x509指输出证书、-days 3650为有效期

此后则输入证书拥有者信息

```ruby
CopyCountry Name (2 letter code) [AU]:CN
State or Province Name (full name) [Some-State]:XxXx
Locality Name (eg, city) []:XxXx
Organization Name (eg, company) [Internet Widgits Pty Ltd]:XX Co. Ltd
Organizational Unit Name (eg, section) []:Dev
Common Name (e.g. server FQDN or YOUR name) []:go-grpc-example
Email Address []:xxx@xxx.com
```

#### 服务端构建TLS证书并认证

```go
Copyfunc main() {
	// 监听本地端口
	listener, err := net.Listen(Network, Address)
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
	}
	// 从输入证书文件和密钥文件为服务端构造TLS凭证
	creds, err := credentials.NewServerTLSFromFile("../pkg/tls/server.pem", "../pkg/tls/server.key")
	if err != nil {
		log.Fatalf("Failed to generate credentials %v", err)
	}
	// 新建gRPC服务器实例,并开启TLS认证
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	// 在gRPC服务器注册我们的服务
	pb.RegisterSimpleServer(grpcServer, &SimpleService{})
	log.Println(Address + " net.Listing whth TLS and token...")
	//用服务器 Serve() 方法以及我们的端口信息区实现阻塞等待，直到进程被杀死或者 Stop() 被调用
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("grpcServer.Serve err: %v", err)
	}
}
```

- `credentials.NewServerTLSFromFile`：从输入证书文件和密钥文件为服务端构造TLS凭证
- `grpc.Creds`：返回一个ServerOption，用于设置服务器连接的凭证。

完整[server.go](https://github.com/Bingjian-Zhu/go-grpc-example/tree/master/7-grpc_security/server/server.go)代码

#### 客户端配置TLS连接

```go
Copyvar grpcClient pb.SimpleClient

func main() {
	//从输入的证书文件中为客户端构造TLS凭证
	creds, err := credentials.NewClientTLSFromFile("../pkg/tls/server.pem", "go-grpc-example")
	if err != nil {
		log.Fatalf("Failed to create TLS credentials %v", err)
	}
	// 连接服务器
	conn, err := grpc.Dial(Address, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("net.Connect err: %v", err)
	}
	defer conn.Close()

	// 建立gRPC连接
	grpcClient = pb.NewSimpleClient(conn)
}
```

- `credentials.NewClientTLSFromFile`：从输入的证书文件中为客户端构造TLS凭证。
- `grpc.WithTransportCredentials`：配置连接级别的安全凭证（例如，TLS/SSL），返回一个DialOption，用于连接服务器。

完整[client.go](https://github.com/Bingjian-Zhu/go-grpc-example/tree/master/7-grpc_security/client/client.go)代码

到这里，已经完成TLS证书认证了，gRPC传输不再是明文传输。此外，添加自定义的验证方法能使gRPC相对更安全。下面以Token认证为例，介绍gRPC如何添加自定义验证方法。

### Token认证[#](https://www.cnblogs.com/FireworksEasyCool/p/12710325.html#1307214989)

客户端发请求时，添加Token到上下文`context.Context`中，服务器接收到请求，先从上下文中获取Token验证，验证通过才进行下一步处理。

#### 客户端请求添加Token到上下文中

```go
Copytype PerRPCCredentials interface {
    GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error)
    RequireTransportSecurity() bool
}
```

gRPC 中默认定义了 `PerRPCCredentials`，是提供用于自定义认证的接口，它的作用是将所需的安全认证信息添加到每个RPC方法的上下文中。其包含 2 个方法：

- `GetRequestMetadata`：获取当前请求认证所需的元数据
- `RequireTransportSecurity`：是否需要基于 TLS 认证进行安全传输

接下来我们实现这两个方法

```go
Copy// Token token认证
type Token struct {
	AppID     string
	AppSecret string
}

// GetRequestMetadata 获取当前请求认证所需的元数据（metadata）
func (t *Token) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"app_id": t.AppID, "app_secret": t.AppSecret}, nil
}

// RequireTransportSecurity 是否需要基于 TLS 认证进行安全传输
func (t *Token) RequireTransportSecurity() bool {
	return true
}
```

然后再客户端中调用Dial时添加自定义验证方法进去

```go
Copy//构建Token
	token := auth.Token{
		AppID:     "grpc_token",
		AppSecret: "123456",
	}
	// 连接服务器
	conn, err := grpc.Dial(Address, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&token))
```

完整[client.go](https://github.com/Bingjian-Zhu/go-grpc-example/tree/master/7-grpc_security/client/client.go)代码

#### 服务端验证Token

首先需要从上下文中获取元数据，然后从元数据中解析Token进行验证

```go
Copy// Check 验证token
func Check(ctx context.Context) error {
	//从上下文中获取元数据
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "获取Token失败")
	}
	var (
		appID     string
		appSecret string
	)
	if value, ok := md["app_id"]; ok {
		appID = value[0]
	}
	if value, ok := md["app_secret"]; ok {
		appSecret = value[0]
	}
	if appID != "grpc_token" || appSecret != "123456" {
		return status.Errorf(codes.Unauthenticated, "Token无效: app_id=%s, app_secret=%s", appID, appSecret)
	}
	return nil
}

// Route 实现Route方法
func (s *SimpleService) Route(ctx context.Context, req *pb.SimpleRequest) (*pb.SimpleResponse, error) {
    //检测Token是否有效
	if err := Check(ctx); err != nil {
		return nil, err
	}
	res := pb.SimpleResponse{
		Code:  200,
		Value: "hello " + req.Data,
	}
	return &res, nil
}
```

- `metadata.FromIncomingContext`：从上下文中获取元数据

完整[server.go](https://github.com/Bingjian-Zhu/go-grpc-example/tree/master/7-grpc_security/server/server.go)代码

服务端代码中，每个服务的方法都需要添加Check(ctx)来验证Token，这样十分麻烦。gRPC拦截器，能很好地解决这个问题。gRPC拦截器功能类似中间件，拦截器收到请求后，先进行一些操作，然后才进入服务的代码处理。

### 服务端添加拦截器[#](https://www.cnblogs.com/FireworksEasyCool/p/12710325.html#2714279692)

```go
Copyfunc main() {
	// 监听本地端口
	listener, err := net.Listen(Network, Address)
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
	}
	// 从输入证书文件和密钥文件为服务端构造TLS凭证
	creds, err := credentials.NewServerTLSFromFile("../pkg/tls/server.pem", "../pkg/tls/server.key")
	if err != nil {
		log.Fatalf("Failed to generate credentials %v", err)
	}
	//普通方法：一元拦截器（grpc.UnaryInterceptor）
	var interceptor grpc.UnaryServerInterceptor
	interceptor = func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		//拦截普通方法请求，验证Token
		err = Check(ctx)
		if err != nil {
			return
		}
		// 继续处理请求
		return handler(ctx, req)
	}
	// 新建gRPC服务器实例,并开启TLS认证和Token认证
	grpcServer := grpc.NewServer(grpc.Creds(creds), grpc.UnaryInterceptor(interceptor))
	// 在gRPC服务器注册我们的服务
	pb.RegisterSimpleServer(grpcServer, &SimpleService{})
	log.Println(Address + " net.Listing whth TLS and token...")
	//用服务器 Serve() 方法以及我们的端口信息区实现阻塞等待，直到进程被杀死或者 Stop() 被调用
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("grpcServer.Serve err: %v", err)
	}
}
```

- `grpc.UnaryServerInterceptor`：为一元拦截器，只会拦截简单RPC方法。流式RPC方法需要使用流式拦截器`grpc.StreamInterceptor`进行拦截。

客户端发起请求，当Token不正确时候，会返回

```powershell
CopyCall Route err: rpc error: code = Unauthenticated desc = Token无效: app_id=grpc_token, app_secret=12345
```

### 总结[#](https://www.cnblogs.com/FireworksEasyCool/p/12710325.html#86735058)

本篇介绍如何为gRPC添加TLS证书认证和自定义认证，从而让gRPC更安全。添加gRPC拦截器，从而省略在每个方法前添加Token检测代码，使代码更简洁。

教程源码地址：https://github.com/Bingjian-Zhu/go-grpc-example

参考：

> - https://www.jianshu.com/p/1fc7130eb2c2
> - https://segmentfault.com/a/1190000007933303
> - https://eddycjy.com/posts/go/grpc/2018-10-07-grpc-tls/
> - https://segmentfault.com/a/1190000007997759
> - https://eddycjy.com/posts/go/grpc/2018-10-10-interceptor/