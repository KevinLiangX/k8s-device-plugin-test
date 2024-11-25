

```go
// Register make the gRPC client to register given resourceName to Kubelet
func (c *TestDevicePlugin) Register() error {
	conn, err := connect(pluginapi.KubeletSocket, common.ConnectTimeout)
	if err != nil {
		return errors.WithMessagef(err, "connect to %s failed", pluginapi.KubeletSocket)
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(common.DeviceSocket), // Endpoint是指定Unix Socket的文件名，这里只会取该文件名字，而不是全路径
		ResourceName: common.ResourceName,  // ResourceName 必须是小写
	}
    // Kubelet会在默认的设备插件路径(/var/lib/kubelet/device-plugins/)下寻找插件的socket文件,即便后面注册是其他地址，也会从该默认路径开始查找
	// /var/lib/kubelet/device-plugins/home/kevinliangx/Codes/GO/k8s-device-plugin-test/testdevices/testdevice.sock
	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return errors.WithMessage(err, "register to kubelet failed")
	}
	return nil
}



```




```shell
# 注册的路径与kubelet去通讯的路径不一致导致，注册不了 journalctl -u kubelet | grep "device-plugin"
11月 25 12:39:51 kevinliangx-HP-ZBook-15-G3 kubelet[16582]:   "Addr": "/var/lib/kubelet/device-plugins/testdevice.sock",
11月 25 12:39:51 kevinliangx-HP-ZBook-15-G3 kubelet[16582]: }. Err: connection error: desc = "transport: Error while dialing: dial unix /var/lib/kubelet/device-plugins/testdevice.sock: connect: no such file or directory"
11月 25 12:39:52 kevinliangx-HP-ZBook-15-G3 kubelet[16582]:   "Addr": "/var/lib/kubelet/device-plugins/testdevice.sock",
11月 25 12:39:52 kevinliangx-HP-ZBook-15-G3 kubelet[16582]: }. Err: connection error: desc = "transport: Error while dialing: dial unix /var/lib/kubelet/device-plugins/testdevice.sock: connect: no such file or directory"
11月 25 12:39:54 kevinliangx-HP-ZBook-15-G3 kubelet[16582]:   "Addr": "/var/lib/kubelet/device-plugins/testdevice.sock",
11月 25 12:39:54 kevinliangx-HP-ZBook-15-G3 kubelet[16582]: }. Err: connection error: desc = "transport: Error while dialing: dial unix /var/lib/kubelet/device-plugins/testdevice.sock: connect: no such file or directory"
11月 25 12:39:57 kevinliangx-HP-ZBook-15-G3 kubelet[16582]:   "Addr": "/var/lib/kubelet/device-plugins/testdevice.sock",
11月 25 12:39:57 kevinliangx-HP-ZBook-15-G3 kubelet[16582]: }. Err: connection error: desc = "transport: Error while dialing: dial unix /var/lib/kubelet/device-plugins/testdevice.sock: connect: no such file or directory"
11月 25 12:40:01 kevinliangx-HP-ZBook-15-G3 kubelet[16582]:   "Addr": "/var/lib/kubelet/device-plugins/testdevice.sock",
11月 25 12:40:01 kevinliangx-HP-ZBook-15-G3 kubelet[16582]: }. Err: connection error: desc = "transport: Error while dialing: dial unix /var/lib/kubelet/device-plugins/testdevice.sock: connect: no such file or directory"
11月 25 12:40:01 kevinliangx-HP-ZBook-15-G3 kubelet[16582]: E1125 12:40:01.992461   16582 client.go:69] "Unable to connect to device plugin client with socket path" err="failed to dial device plugin: context deadline exceeded" path="/var/lib/kubelet/device-plugins/testdevice.sock"

```