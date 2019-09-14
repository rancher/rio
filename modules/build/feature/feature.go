package feature

import (
	"context"
	"strings"

	"github.com/rancher/rio/modules/build/controllers/build"
	"github.com/rancher/rio/modules/build/controllers/gitcommit"
	"github.com/rancher/rio/modules/build/controllers/pod"
	"github.com/rancher/rio/modules/build/controllers/proxy"
	"github.com/rancher/rio/modules/build/controllers/service"
	stack1 "github.com/rancher/rio/modules/build/controllers/stack"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/start"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	dockerSocket     = "/run/docker.sock"
	containerdSocket = "/run/k3s/containerd/containerd.sock"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	runtime := containerRuntime(rContext)
	socket := ""
	if runtime == "docker" {
		socket = dockerSocket
	} else if runtime == "containerd" {
		socket = containerdSocket
	}
	feature := &features.FeatureController{
		FeatureName: "build",
		FeatureSpec: v1.FeatureSpec{
			Description: "Rio Build, from source code to deployment",
			Enabled:     !constants.DisableBuild,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Namespace, "build"),
			stack.NewSystemStack(apply, rContext.Namespace, "buildkit-local"),
		},
		Controllers: []features.ControllerRegister{
			service.Register,
			build.Register,
			gitcommit.Register,
			proxy.Register,
			stack1.Register,
			pod.Register,
		},
		FixedAnswers: map[string]string{
			"NAMESPACE":      rContext.Namespace,
			"RUNTIME":        runtime,
			"SOCKET_ADDRESS": socket,
		},
		OnStart: func(feature *v1.Feature) error {
			return start.All(ctx, 5,
				rContext.Build,
				rContext.Webhook,
			)
		},
	}
	return feature.Register()
}

func containerRuntime(rContext *types.Context) string {
	nodes, err := rContext.Core.Core().V1().Node().List(metav1.ListOptions{})
	if err == nil {
		if len(nodes.Items) == 1 {
			r := nodes.Items[0].Status.NodeInfo.ContainerRuntimeVersion
			if strings.Contains(r, "containerd") {
				return "containerd"
			} else if strings.Contains(r, "docker") {
				return "docker"
			}
		} else if len(nodes.Items) > 1 {
			return "multiple-nodes"
		}
	}
	return ""
}
