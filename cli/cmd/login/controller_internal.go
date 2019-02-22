// +build controller_internal

package login

import (
	"context"

	"github.com/rancher/norman"
	"github.com/rancher/rio/pkg/server"
)

func runController(ctx context.Context, kubeconfig string, runDns bool) (context.Context, error) {
	ctx, _, err := server.NewConfig(runDns).Build(ctx, &norman.Options{
		K8sMode:    "external",
		KubeConfig: kubeconfig,
	})
	return ctx, err
}
