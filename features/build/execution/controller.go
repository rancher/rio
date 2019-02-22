package execution

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rio/types"
	v1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	webhookv1 "github.com/rancher/rio/types/apis/webhookinator.rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := executionHandler{
		serviceClient: rContext.Rio.Service,
		serviceLister: rContext.Rio.Service.Cache(),
	}

	rContext.Webhook.GitWebHookExecution.OnChange(ctx, "webhook-execution", h.onChange)
	return nil
}

type executionHandler struct {
	serviceLister v1.ServiceClientCache
	serviceClient v1.ServiceClient
}

func (e executionHandler) onChange(obj *webhookv1.GitWebHookExecution) (runtime.Object, error) {
	ns, name := kv.Split(obj.Spec.GitWebHookReceiverName, ":")
	service, err := e.serviceLister.Get(ns, name)
	if err != nil {
		if errors.IsNotFound(err) {
			return obj, nil
		}
		return obj, err
	}
	copiedService := service.DeepCopy()
	return webhookv1.GitWebHookExecutionConditionHandled.Once(obj, func() (runtime.Object, error) {
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
