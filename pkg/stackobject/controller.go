package stackobject

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/mapper/convert"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	corev1 "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stacknamespace"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/rancher/wrangler/pkg/relatedresource"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

var ErrSkipObjectSet = errors.New("skip objectset")

type ControllerWrapper interface {
	Informer() cache.SharedIndexInformer
	AddGenericHandler(ctx context.Context, name string, handler generic.Handler)
	AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler)
	Enqueue(namespace, name string)
	Updater() generic.Updater
}

type Populator func(obj runtime.Object, stack *riov1.Stack, os *objectset.ObjectSet) error

type Controller struct {
	Apply          apply.Apply
	Populator      Populator
	name           string
	stacksCache    v1.StackCache
	indexer        cache.Indexer
	namespaceCache corev1.NamespaceCache
	injectors      []string
}

func NewGeneratingController(ctx context.Context, rContext *types.Context, name string, controller ControllerWrapper, injectors ...string) *Controller {
	sc := &Controller{
		name:           name,
		Apply:          rContext.Apply.WithSetID(name).WithStrictCaching(),
		stacksCache:    rContext.Rio.Rio().V1().Stack().Cache(),
		namespaceCache: rContext.Core.Core().V1().Namespace().Cache(),
		injectors:      injectors,
		indexer:        controller.Informer().GetIndexer(),
	}
	relatedresource.Watch(ctx, "stackchange-"+name, sc.resolve, controller, rContext.Rio.Rio().V1().Stack())

	lcName := name + "-object-controller"
	controller.AddGenericHandler(ctx, lcName, generic.UpdateOnChange(controller.Updater(), sc.OnChange))
	controller.AddGenericRemoveHandler(ctx, lcName, sc.OnRemove)
	return sc
}

func (o *Controller) resolve(ns, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	switch obj.(type) {
	case *riov1.Stack:
		var ret []interface{}
		if err := cache.ListAllByNamespace(o.indexer, obj.(*riov1.Stack).Name, labels.Everything(), func(obj interface{}) {
			ret = append(ret, obj)
		}); err != nil {
			return nil, err
		}
		var key []relatedresource.Key
		for _, o := range ret {
			meta, err := meta.Accessor(o)
			if err != nil {
				return nil, err
			}
			key = append(key, relatedresource.Key{
				Name:      meta.GetName(),
				Namespace: meta.GetNamespace(),
			})
		}
		return key, nil
	}
	return nil, nil
}

func (o *Controller) OnRemove(key string, obj runtime.Object) (runtime.Object, error) {
	return obj, o.Apply.WithOwner(obj).Apply(nil)
}

func (o *Controller) OnChange(key string, obj runtime.Object) (runtime.Object, error) {
	if o.Populator == nil {
		return obj, nil
	}

	meta, err := meta.Accessor(obj)
	if err != nil {
		return obj, err
	}

	stack, err := stacknamespace.GetStack(meta, o.stacksCache, o.namespaceCache)
	if apierrors.IsNotFound(err) {
		return obj, nil
	}
	if err != nil {
		return obj, err
	}

	os := objectset.NewObjectSet()
	if err := o.Populator(obj, stack, os); err != nil {
		if err == ErrSkipObjectSet {
			return obj, nil
		}
		os.AddErr(err)
	}

	desireset := o.Apply.WithOwner(obj)
	if !stack.Spec.DisableMesh {
		for _, i := range o.injectors {
			desireset = desireset.WithInjectorName(i)
		}
	}

	return obj, o.getCondition().Do(func() (runtime.Object, error) {
		return obj, desireset.Apply(os)
	})
}

func (o *Controller) getCondition() condition.Cond {
	buffer := strings.Builder{}
	buffer.WriteString(string(riov1.StackConditionDeployed))
	for _, part := range strings.Split(o.name, "-") {
		buffer.WriteString(convert.Capitalize(part))
	}
	return condition.Cond(buffer.String())
}
