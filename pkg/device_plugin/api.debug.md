```go
func (c *TestDevicePlugin) GetDevicePluginOptions(_ context.Context, _ *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{PreStartRequired: true}, nil
}
```
context.Context： 上下文对象，通常用于控制方法的生命周期，例如超时和取消。此处未使用，因此使用 _ 忽略。
_ *pluginapi.Empty： 空参数，来自 pluginapi 的定义，用于占位或符合接口规范。此处未使用。

return &pluginapi.DevicePluginOptions{PreStartRequired: true}, nil： 返回一个 DevicePluginOptions 指针对象。
PreStartRequired: true： 表示此设备插件要求在分配设备之前，Kubelet 必须调用设备插件的 PreStartContainer 方法。 该选项通常用于设备插件在设备使用前需要完成一些初始化操作（例如加载驱动或配置硬件）。

这里并没有对PreStartRequired进行具体操作，返回一个空的返回

```go
func (c *TestDevicePlugin) PreStartContainer(_ context.Context, _ *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}
```

```go
// gRPC 的 Stream 方法，建立长连接，用于将设备的状态和列表实时通知Kubernetes Kubelet
func (c *TestDevicePlugin) ListAndWatch(_ *pluginapi.Empty, srv pluginapi.DevicePlugin_ListAndWatchServer) error {
	// _ *pluginapi.Empty 空参数，符合gRPC方法签名规范，不需要实际内容，使用_忽略
	// srv 一个gRPC流对象，支持向Kubelet发送设备列表更新
	devs := c.dm.Devices()
	klog.Infof("find devices [%s]", String(devs))

	err := srv.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
	//srv.Send 通过gRPC流 srv发送设备列表到Kubelet,设备列表包含每个设备的状态和属性
	if err != nil {
		return errors.WithMessage(err, "send device failed")
	}

	klog.Infoln("waiting for device update")
	for range c.dm.notify {
		// 监听c.dm.notify通道，用于通知设备状态的变化。每当通道收到消息时，重新获取设备列表并发送更新
		devs = c.dm.Devices()
		klog.Infof("device update, new device list [%s]", String(devs))
		_ = srv.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
		// 再次获取设备列表并发送更新给kubelet, 使用_ 来忽略错误，用于表示这些错误不会终端流程
	}

	return nil
}
```
功能说明
1. 初始设备列表发送： 方法启动后会立即获取当前设备列表并发送给 Kubelet。
2. 动态设备更新： 通过 c.dm.notify 通道监听设备变化，每当设备变化时，重新获取设备列表并发送更新。
3. 流式通信： 使用 gRPC 双向流向 Kubelet 实时发送设备信息，这符合 Kubernetes 设备插件协议要求。


```go
// 告知kubelet 怎么将设备分配给容器，不同的异构厂家不一样，一般在对应容器中增加一个环境变量，Test=$deviceId
func (c *TestDevicePlugin) Allocate(_ context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	ret := &pluginapi.AllocateResponse{}
    // 构建单个容器的分配响应
	for _, req := range reqs.ContainerRequests {
		klog.Infof("[Allocate] received request : %v", strings.Join(req.DevicesIDs, ","))
		resp := pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				"Test": strings.Join(req.DevicesIDs, ","),
			},
        // 设置环境变量，Envs 是一个环境变量，Kubelet会将这些变量注入到容器中，设备ID拼接成字符串，存储在环境变量，该环境变量与底层GPU设备有关系。
		}
		ret.ContainerResponses = append(ret.ContainerResponses, &resp)
	}
	return ret, nil
}

```

功能说明
1. 设备分配流程：
    - Kubernetes 请求设备插件为容器分配设备。
    - 设备插件返回分配结果，包括设备信息和需要注入到容器的环境变量。 将 GPU ID 写入环境变量，容器内的应用可以使用这些 ID 来访问特定 GPU。
2. 设置环境变量：
    - Kubernetes 会将设备插件返回的环境变量注入到容器中，容器内的应用可以通过环境变量获取设备信息。
3. 多容器支持：
    - 通过循环处理每个容器的分配请求，支持同一 Pod 中多个容器的设备分配。