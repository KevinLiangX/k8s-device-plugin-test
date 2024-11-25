```go
//一个用于通过 Unix 域套接字连接 gRPC 服务的函数实现。kubelet.sock是一个unix套接字
func connect(socketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// 使用 context.WithTimeout 创建带超时的上下文（ctx），用于限制 gRPC 的连接时间。
    // 使用 defer cancel() 确保函数返回时，释放上下文资源，避免资源泄漏。
	c, err := grpc.DialContext(ctx, socketPath, //ctx, 控制链接过程的超时时间; socketPath: Unix套接字路径，链接的地址
		grpc.WithTransportCredentials(insecure.NewCredentials()), //设置不使用TLS,因为unix套接字在本地通信中通常不需要加密
		grpc.WithBlock(), // 指定阻塞式，调用DialContext时会阻塞直到连接成功或者超时
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			if deadline, ok := ctx.Deadline(); ok {
				return net.DialTimeout("unix", addr, time.Until(deadline))
			}
			return net.DialTimeout("unix", addr, common.ConnectTimeout)
		}),
		// 自定义拨号逻辑，覆盖默认的网络拨号方式
		// 使用net.DialTimeout 通过Unix套接字发起连接
		// 如果上下文ctx 包含Dealine,根据 time.Until(deadline)计算剩余时间设置超时时间，否则，使用默认超时时间(common.ConnectTimeout)
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

在gRPC中，默认的网络拨号方式是基于TCP/IP的套接字，具体表现如下：
- 如果没有指定自定义拨号器（例如：grpc.WithContextDialer）,gprc.Dial或者grpc.DialContext会使用TCP协议连接到提供的目的地址，通常是host:port格式
- 默认情况下， gRPC会尝试解析目标地址：
    - 如果目标地址是域名或者IP地址端口号（如 localhost:50051或者192.168.1.1:50051）,会使用标准的TCP套接字发起连接
    - 如果目标地址不是标准格式，可能会依赖解析器插件，比如针对不同协议的scheme(如：dns://,unixt:)
conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	
注意事项
    默认 TCP 适用场景：在服务和客户端运行在不同机器上，通过 IP 和端口号连接时，默认 TCP 非常适用。TCP 是跨网络通信的标准协议。
    Unix 域套接字适用场景：客户端和服务运行在同一台机器上，且追求更高性能或更小的延迟时，可以使用 Unix 域套接字。与 TCP 相比，Unix 域套接字可以省去协议栈处理，性能更好。
    指定 Scheme 的地址：如果目标地址带有 scheme（如 dns:// 或 unix:），gRPC 会根据 scheme 选择适当的解析器来处理地址。
```