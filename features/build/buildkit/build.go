package buildkit

import (
	"context"
	"time"

	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/controllers/data"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	corev1 "github.com/rancher/types/apis/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		processor: objectset.NewProcessor("system-data").
			Client(rContext.Rio.Stack),
		secrets:   rContext.Core.Secret.Cache(),
		namespace: rContext.Core.Namespace,
	}
	rContext.Core.Namespace.OnChange(ctx, "project-build", h.sync)
	return nil
}

type handler struct {
	processor *objectset.Processor
	secrets   corev1.SecretClientCache
	namespace corev1.NamespaceClient
}

func (h handler) sync(obj *v1.Namespace) (runtime.Object, error) {
	rioCert, _ := h.secrets.Get(settings.CloudNamespace, "rio-certs")
	if rioCert == nil || rioCert.Annotations["certificate-status"] != "ready" {
		h.sleepAndEnqueue(obj)
		return obj, nil
	}

	if obj.Labels["rio.cattle.io/project"] != "true" {
		return obj, nil
	}

	// deploy build-controller in rio-system
	if obj.Name == settings.RioSystemNamespace {
		os := objectset.NewObjectSet()
		stack := data.Stack("build-controller", obj.Name, riov1.StackSpec{
			EnableKubernetesResources: true,
		})
		os.Add(stack)
		return obj, h.processor.NewDesiredSet(obj, os).Apply()
	}

	// deploy buildkit, registry and webhook server in each project
	os := objectset.NewObjectSet()
	stack := data.Stack("build", obj.Name, riov1.StackSpec{})
	os.Add(stack)
	return obj, h.processor.NewDesiredSet(obj, os).Apply()
}

func (h handler) sleepAndEnqueue(obj *v1.Namespace) {
	time.Sleep(15 * time.Second)
	h.namespace.Enqueue("", obj.Name)
}
