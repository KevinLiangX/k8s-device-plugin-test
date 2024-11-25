package device_plugin

import (
	"context"
	"github.com/pkg/errors"
	"k8s-device-plugin-test/pkg/common"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"path"
)

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
		Endpoint:     path.Base(common.DeviceSocket),
		ResourceName: common.ResourceName,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return errors.WithMessage(err, "register to kubelet failed")
	}
	return nil
}
