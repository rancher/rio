package buildkit

import (
	"context"

	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/controllers/data"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	corev1 "github.com/rancher/types/apis/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	buildController = "build-controller"
	build           = "build"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		processor: objectset.NewProcessor("system-data").
			Client(rContext.Rio.Stack),
		secrets:       rContext.Core.Secret.Cache(),
		secretsClient: rContext.Core.Secret,
		namespace:     rContext.Core.Namespace,
		stacks:        rContext.Rio.Stack,
	}
	rContext.Core.Namespace.OnChange(ctx, "project-build", h.sync)
	return nil
}

type handler struct {
	processor     *objectset.Processor
	secrets       corev1.SecretClientCache
	secretsClient corev1.SecretClient
	namespace     corev1.NamespaceClient
	stacks        riov1.StackClient
}

func (h handler) sync(obj *v1.Namespace) (runtime.Object, error) {
	if obj.Labels["rio.cattle.io/project"] != "true" {
		return obj, nil
	}

	//deploy build-controller in rio-system
	if obj.Name == settings.RioSystemNamespace {
		stack := data.Stack(buildController, obj.Name, riov1.StackSpec{
			EnableKubernetesResources: true,
		})
		if _, err := h.stacks.Create(stack.(*riov1.Stack)); err != nil && !errors.IsAlreadyExists(err) {
			return obj, nil
		}
		return obj, nil
	}

	// deploy buildkit, registry and webhook server in each project
	stack := data.Stack(build, obj.Name, riov1.StackSpec{
		EnableKubernetesResources: true,
	})
	if _, err := h.stacks.Create(stack.(*riov1.Stack)); err != nil && !errors.IsAlreadyExists(err) {
		return obj, nil
	}

	return obj, nil
}
