package device_plugin

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s-device-plugin-test/pkg/common"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"log"
	"net"
	"os"
	"path"
	"syscall"
	"time"
)

type TestDevicePlugin struct {
	server *grpc.Server
	stop   chan struct{} //this channel signals to stop the device plugin
	dm     *DeviceMonitor
}

func NewTestDevicePlugin() *TestDevicePlugin {
	return &TestDevicePlugin{
		server: grpc.NewServer(grpc.EmptyServerOption{}),
		stop:   make(chan struct{}),
		dm:     NewDeviceMonitor(common.DevicePath),
	}
}

// Run start gRPC server and device watcher
func (c *TestDevicePlugin) Run() error {
	// List and Watch socket unix
	err := c.dm.List()
	if err != nil {
		log.Fatalf("list devices error: %v", err)
	}

	go func() {
		if err = c.dm.Watch(); err != nil {
			log.Println("watch devices error")
		}
	}()

	pluginapi.RegisterDevicePluginServer(c.server, c)
	// delete old unix socket before start
	//socket := path.Join(pluginapi.DevicePluginPath, common.DeviceSocket)
	socket := path.Join(pluginapi.DevicePluginPath, common.DeviceSocket)
	err = syscall.Unlink(socket)
	if err != nil && !os.IsNotExist(err) {
		return errors.WithMessagef(err, "delete socket %s failed", socket)
	}

	sock, err := net.Listen("unix", socket)
	if err != nil {
		return errors.WithMessagef(err, "listen unix %s failed", socket)
	}

	go c.server.Serve(sock)

	// wait for server to start by launching a blocking connection
	conn, err := connect(socket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()

	return nil
}

// dial establishes the gPRC communication with the registered device plugin
func connect(socketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	c, err := grpc.DialContext(ctx, socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			if deadline, ok := ctx.Deadline(); ok {
				return net.DialTimeout("unix", addr, time.Until(deadline))
			}
			return net.DialTimeout("unix", addr, common.ConnectTimeout)
		}),
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}
