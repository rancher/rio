// +build k3s

package server

import (
	"net"
	"time"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rancher/pkg/remotedialer"
	"k8s.io/kubernetes/cmd/kube-apiserver/app"
)

func setupK3s(tunnelServer *remotedialer.Server) {
	app.DefaultProxyDialerFn = func(network, address string) (net.Conn, error) {
		_, port, _ := net.SplitHostPort(address)
		addr := "127.0.0.1"
		if port != "" {
			addr += ":" + port
		}
		nodeName, _ := kv.Split(address, ":")
		return tunnelServer.Dial(nodeName, 15*time.Second, "tcp", addr)
	}
}
