package main

import (
	"k8s-device-plugin-test/pkg/device_plugin"
	"k8s-device-plugin-test/pkg/utils"
	"k8s.io/klog"
)

func main() {
	klog.Infof("Device plugin staring")
	dp := device_plugin.NewTestDevicePlugin()
	go dp.Run()

	// register when device plugin start
	if err := dp.Register(); err != nil {
		klog.Fatalf("register to kubelet failed: %v", err)
	}

	// watch kubelet.sock when kubelet restart, exit device plugin,then will restart by DaemonSet
	stop := make(chan struct{})
	err := utils.WatchKubelet(stop)
	if err != nil {
		klog.Fatalf("start to kubelet failed: %v", err)
	}
	<-stop
	klog.Infof("kubelet restart ,exiting")
}
