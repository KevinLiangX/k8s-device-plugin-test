```go
// List all devices，发现设备 遍历 /home/kevinliangx/Codes/GO/k8s-device-plugin-test/testdevices 目录下的所有文件，每个文件都会当作一个设备
func (d *DeviceMonitor) List() error {
	err := filepath.Walk(d.path, func(path string, info fs.FileInfo, err error) error {
		fmt.Printf("Path: %v, info: %v\n", path, info)
		if info.IsDir() {
			klog.Infof("%s is dir,skip", path)
			return nil
		}
		d.devices[info.Name()] = &pluginapi.Device{
			ID:     info.Name(),
			Health: pluginapi.Healthy,
		}
		return nil
	})

	return errors.WithMessagef(err, "walk [%s] failed", d.path)
}
```
以上代码实现的查询d.path，路径中的文件，d.path可以理解为设备文件的目录，在pkg/common/constant.go中定义的
DevicePath     string = "/home/kevinliangx/Codes/GO/k8s-device-plugin-test/testdevices"
调用 filepath.Walk 函数
    filepath.Walk 是 Go 的标准库函数，用于递归地遍历指定路径下的所有文件和文件夹。
    参数解析：
        d.path：表示起始路径，List 方法中的 DeviceMonitor 结构体的 path 字段。
        func(path string, info fs.FileInfo, err error) error：匿名函数（回调函数），在每个文件/文件夹上调用。
        path：当前遍历到的文件或文件夹的路径。
        info：文件或文件夹的元信息（类型、大小等），类型为 fs.FileInfo。
        err：在访问当前文件/文件夹时可能遇到的错误。
设备目前是通过一个字典的方式进行存储，也就是存入在内容中
        d.devices[info.Name()] = &pluginapi.Device{
            ID:     info.Name(), 设备名称
            Health: pluginapi.Healthy, 设备状态
        }
打印出的日志：
/home/kevinliangx/Codes/GO/k8s-device-plugin-test/testdevices is dir,skip
Path: /home/kevinliangx/Codes/GO/k8s-device-plugin-test/testdevices/testdevice-01, 
info: &{testdevice-01 0 420 {0 63867683139 0xd80a80} {2049 933809 1 33188 0 0 0 0 0 4096 0 {1732086340 0} {1732086339 0} {1732086340 464427489} [0 0 0]}}
    
```go
// Watch devices change，启动Goroutine监听设备的变化，该文件下文件发生变化时，通过chan发送通知，更新内存device map，并将最新的设备信息发送给Kubelet
func (d *DeviceMonitor) Watch() error {
	klog.Infoln("keep watching devices for changing")

	w, err := fsnotify.NewWatcher()
	```
	创建文件监视器,调用 fsnotify.NewWatcher 创建一个文件系统监视器对象。w 是监视器实例。
    用途：监控指定路径及其子路径下的文件系统事件（如文件创建、删除等）。
	```
	if err != nil {
		return errors.WithMessagef(err, "new watcher failed")
	}
	defer w.Close()
    ```
    延迟关闭监视器,使用 defer 确保在方法结束时，资源 w 被释放。
    用途：避免资源泄露。
    ```
	errChan := make(chan error)
	```
	创建错误通道,定义一个 chan error 类型的通道，用于在监控的协程中传递可能发生的错误。
    用途：实现协程与主协程的错误通信。
	```
	go func() { //监控文件系统事件和错误，保证主协程不被阻塞。
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("device wather panic:%v", r)
			}
			```
            捕获异常,使用 recover 捕获可能的运行时崩溃，避免 goroutine 崩溃导致程序退出。
            用途：提高方法的鲁棒性，将 panic 转换为错误并发送到 errChan

            ```
		}()
		for { //使用无限循环，保证实时处理文件系统事件。
			select {
			case event, ok := <-w.Events:
				if !ok {
					continue
				}
				klog.Infof("fsnotify device event: %s %s", event.Name, event.Op.String())

				if event.Op == fsnotify.Create {
					dev := path.Base(event.Name)
					d.devices[dev] = &pluginapi.Device{
						ID:     dev,
						Health: pluginapi.Healthy,
					}
					d.notify <- struct{}{}
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					dev := path.Base(event.Name)
					delete(d.devices, dev)
					d.notify <- struct{}{}
					klog.Infof("device [%s] removed", dev)
				}
			case err, ok := <-w.Errors:
				if !ok {
					continue
				}
				klog.Errorf("fsnotify watch device failed:%v", err)
			}
		}
	}()

	err = w.Add(d.path)
	if err != nil {
		return fmt.Errorf("watch device error:%v", err)
	}
	```
    添加监控路径,使用 w.Add 将 d.path 添加到监控器中。如果失败，返回错误信息。
     用途：设置监控器的监视范围。

    ```

	return <-errChan
	```
    返回错误,阻塞等待 errChan 通道中的错误。如果 goroutine 运行中没有错误，此方法会一直阻塞。
    ```
}


```
详细讲解一下 上面的代码的逻辑：
1. go func() 的执行：
    go func() 启动了一个新的 goroutine，它在独立的协程中运行，主要负责监听 w.Events 和 w.Errors 通道中的事件或错误。
    它的启动是非阻塞的，即在主协程中执行 go func() 后，主协程不会等待 goroutine 完成，会立即继续执行下面的代码。
2. w.Add(d.path) 的执行：
    这是在主协程中执行的。它的目的是将 d.path 添加到 w（文件系统监视器）的监视范围内。
    如果 w.Add(d.path) 执行失败，程序会直接返回错误，不会继续等待 go func() 执行结果。
目前有个问题， fsnotify只能监听程序启动后的第一次操作。奶奶的。
