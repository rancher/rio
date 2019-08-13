package modules

import (
	"context"
	"fmt"

	"github.com/rancher/rio/modules/autoscale"
	"github.com/rancher/rio/modules/build"
	"github.com/rancher/rio/modules/istio"
	istio2 "github.com/rancher/rio/modules/istio/controllers/istio"
	"github.com/rancher/rio/modules/monitoring"
	"github.com/rancher/rio/modules/service"
	"github.com/rancher/rio/modules/system"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rioContext *types.Context) error {
	if !constants.DisableIstio {
		mesh := stack.NewSystemStack(rioContext.Apply, rioContext.Namespace, "mesh")
		answer := map[string]string{
			"HTTP_PORT":         constants.DefaultHTTPOpenPort,
			"HTTPS_PORT":        constants.DefaultHTTPSOpenPort,
			"TELEMETRY_ADDRESS": fmt.Sprintf("%s.%s.svc.cluster.local", constants.IstioTelemetry, rioContext.Namespace),
			"NAMESPACE":         rioContext.Namespace,
			"TAG":               "1.1.3",
		}
		if err := mesh.Deploy(answer); err != nil {
			return err
		}
		if err := istio2.RegisterInjectors(ctx, rioContext); err != nil {
			return err
		}
	}
	if err := istio.Register(ctx, rioContext); err != nil {
		return err
	}
	if err := system.Register(ctx, rioContext); err != nil {
		return err
	}
	if err := service.Register(ctx, rioContext); err != nil {
		return err
	}
	if err := monitoring.Register(ctx, rioContext); err != nil {
		return err
	}
	if err := autoscale.Register(ctx, rioContext); err != nil {
		return err
	}
	if err := build.Register(ctx, rioContext); err != nil {
		return err
	}
	return nil
}
