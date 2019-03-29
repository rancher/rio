package execution

import (
	"context"

	v1 "github.com/rancher/rio/pkg/generated/controllers/webhookinator.rio.cattle.io/v1"

	webhookv1 "github.com/rancher/rio/pkg/apis/webhookinator.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
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
		v1.UpdateGitWebHookExecutionOnChange(rContext.Webhook.Webhookinator().V1().GitWebHookExecution().Updater(), h.onChange))
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
	copiedService := service.DeepCopy()
	return obj, webhookv1.GitWebHookExecutionConditionHandled.Once(obj, func() (runtime.Object, error) {
		if copiedService.Spec.ImageBuild != nil && obj.Spec.Branch == copiedService.Spec.ImageBuild.Branch {
			if obj.Spec.Tag != "" {
				copiedService.Spec.ImageBuild.Commit = ""
				copiedService.Spec.ImageBuild.Tag = obj.Spec.Tag
			} else {
				copiedService.Spec.ImageBuild.Commit = obj.Spec.Commit
				copiedService.Spec.ImageBuild.Tag = ""
			}
			if _, err := e.serviceClient.Update(copiedService); err != nil {
				return obj, err
			}
		}
		return obj, nil
	})
}
