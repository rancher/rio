package data

import (
	"context"

	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/pkg/project"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	corev1 "github.com/rancher/types/apis/core/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	d := &dataHandler{
		inCluster: rContext.InCluster,
		processor: objectset.NewProcessor("system-data").
			Client(rContext.Core.Namespace,
				rContext.Rio.Stack),
	}

	rContext.Core.Namespace.OnChange(ctx, "data-controller", d.onChange)
	return addNameSpace(rContext.Core.Namespace)
}

type dataHandler struct {
	inCluster bool
	processor objectset.Processor
}

func (d *dataHandler) onChange(obj *v1.Namespace) (runtime.Object, error) {
	if obj.Name != settings.RioSystemNamespace {
		return obj, nil
	}

	os := addData(d.inCluster)
	return obj, d.processor.NewDesiredSet(obj, os).Apply()
}

func addData(inCluster bool) *objectset.ObjectSet {
	os := objectset.NewObjectSet()

	os.Add(systemStacks(inCluster)...)

	if err := localStacks(os); err != nil {
		os.AddErr(err)
	}

	return os
}

func addNameSpace(client corev1.NamespaceClient) error {
	ns := corev1.NewNamespace("", settings.RioSystemNamespace, v1.Namespace{})
	ns.Labels = map[string]string{
		project.ProjectLabel: "true",
	}

	ns, err := client.Create(ns)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}
