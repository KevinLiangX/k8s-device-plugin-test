package common

import "time"

const (
	ResourceName   string = "kevinliangx.com/testdevice"
	DevicePath     string = "/home/kevinliangx/Codes/GO/k8s-device-plugin-test/testdevices"
	DeviceSocket   string = "testdevice.sock"
	ConnectTimeout        = time.Second * 5
)
