package execution

import (
	"context"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	webhookv1 "github.com/rancher/rio/pkg/apis/webhookinator.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	webhookv1controller "github.com/rancher/rio/pkg/generated/controllers/webhookinator.rio.cattle.io/v1"
	name2 "github.com/rancher/rio/pkg/name"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/kv"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := executionHandler{
		serviceClient: rContext.Rio.Rio().V1().Service(),
		serviceLister: rContext.Rio.Rio().V1().Service().Cache(),
	}

	rContext.Webhook.Webhookinator().V1().GitWebHookExecution().OnChange(ctx, "webhook-execution",
		webhookv1controller.UpdateGitWebHookExecutionOnChange(rContext.Webhook.Webhookinator().V1().GitWebHookExecution().Updater(), h.onChange))
	return nil
}

type executionHandler struct {
	serviceLister riov1controller.ServiceCache
	serviceClient riov1controller.ServiceClient
}

func (e executionHandler) onChange(key string, obj *webhookv1.GitWebHookExecution) (*webhookv1.GitWebHookExecution, error) {
	if obj == nil {
		return nil, nil
	}

	ns, name := kv.Split(obj.Spec.GitWebHookReceiverName, ":")
	service, err := e.serviceLister.Get(ns, name)
	if err != nil {
		if errors.IsNotFound(err) {
			return obj, nil
		}
		return obj, err
	}
	return obj, webhookv1.GitWebHookExecutionConditionHandled.Once(obj, func() (runtime.Object, error) {
		specCopy := service.Spec.DeepCopy()
		specCopy.Build.Repo = obj.Spec.RepositoryURL
		specCopy.Build.Revision = obj.Spec.Commit
		specCopy.Image = ""
		newServiceName := name2.SafeConcatName(name, name2.Hex(obj.Spec.RepositoryURL, 7), obj.Spec.Commit)
		newService := riov1.NewService(service.Namespace, newServiceName, riov1.Service{
			Spec: *specCopy,
		})
		if _, err := e.serviceClient.Create(newService); err != nil {
			return obj, err
		}
		return obj, nil
	})
}
