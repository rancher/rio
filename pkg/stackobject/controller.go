package stackobject

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/norman/condition"
	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/lifecycle"
	"github.com/rancher/norman/objectclient"
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rio/features/routing/pkg/istio/config"
	"github.com/rancher/rio/pkg/stacknamespace"
	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"github.com/rancher/types/apis/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

var ErrSkipObjectSet = errors.New("skip objectset")

type ClientAccessor interface {
	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
}

type Populator func(obj runtime.Object, stack *riov1.Stack, os *objectset.ObjectSet) error

type Controller struct {
	Processor      *objectset.Processor
	Populator      Populator
	name           string
	stacksCache    riov1.StackClientCache
	namespaceCache v1.NamespaceClientCache
	injector       []config.IstioInjector
}

func NewGeneratingController(ctx context.Context, rContext *types.Context, name string, client ClientAccessor, injector ...config.IstioInjector) *Controller {
	sc := &Controller{
		name:           name,
		Processor:      objectset.NewProcessor(name),
		stacksCache:    rContext.Rio.Stack.Cache(),
		namespaceCache: rContext.Core.Namespace.Cache(),
		injector:       injector,
	}

	lcName := name + "-object-controller"
	lc := lifecycle.NewObjectLifecycleAdapter(lcName, false, sc, client.ObjectClient())
	client.Generic().AddHandler(ctx, name, lc)

	return sc
}

func (o *Controller) Create(obj runtime.Object) (runtime.Object, error) {
	return obj, nil
}

func (o *Controller) Finalize(obj runtime.Object) (runtime.Object, error) {
	return obj, o.Processor.Remove(obj)
}

func (o *Controller) Updated(obj runtime.Object) (runtime.Object, error) {
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

	desireset := o.Processor.NewDesiredSet(obj, os)
	if !stack.Spec.DisableMesh {
		for _, i := range o.injector {
			desireset.AddInjector(i.Inject)
		}
	}

	cond := o.getCondition()

	if err = desireset.Apply(); err != nil {
		cond.False(obj)
		cond.ReasonAndMessageFromError(obj, err)
	} else if cond.GetLastUpdated(obj) != "" {
		cond.True(obj)
		cond.Message(obj, "")
		cond.Reason(obj, "")
	}

	return obj, err
}

func (o *Controller) getCondition() condition.Cond {
	setID := o.Processor.SetID()
	buffer := strings.Builder{}
	buffer.WriteString(string(riov1.StackConditionDeployed))
	for _, part := range strings.Split(setID, "-") {
		buffer.WriteString(convert.Capitalize(part))
	}
	return condition.Cond(buffer.String())
}
