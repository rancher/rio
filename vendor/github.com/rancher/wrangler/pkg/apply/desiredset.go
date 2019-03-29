package apply

import (
	"github.com/rancher/mapper"
	"github.com/rancher/wrangler/pkg/apply/injectors"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

type desiredSet struct {
	a                *apply
	defaultNamespace string
	strictCaching    bool
	pruneTypes       map[schema.GroupVersionKind]cache.SharedIndexInformer
	patchers         map[schema.GroupVersionKind]Patcher
	remove           bool
	setID            string
	objs             *objectset.ObjectSet
	codeVersion      string
	owner            runtime.Object
	injectors        []injectors.ConfigInjector
	injectorNames    []string
	errs             []error
}

func (o *desiredSet) err(err error) error {
	o.errs = append(o.errs, err)
	return o.Err()
}

func (o desiredSet) Err() error {
	return mapper.NewErrors(append(o.errs, o.objs.Err())...)
}

func (o desiredSet) Apply(set *objectset.ObjectSet) error {
	o.objs = set
	return o.apply()
}

func (o desiredSet) WithSetID(id string) Apply {
	o.setID = id
	return o
}

func (o desiredSet) WithOwner(obj runtime.Object) Apply {
	o.owner = obj
	return o
}

func (o desiredSet) WithInjector(injs ...injectors.ConfigInjector) Apply {
	o.injectors = append(o.injectors, injs...)
	return o
}

func (o desiredSet) WithInjectorName(injs ...string) Apply {
	o.injectorNames = append(o.injectorNames, injs...)
	return o
}

func (o desiredSet) WithCacheTypes(igs ...InformerGetter) Apply {
	pruneTypes := map[schema.GroupVersionKind]cache.SharedIndexInformer{}
	for k, v := range o.pruneTypes {
		pruneTypes[k] = v
	}

	for _, ig := range igs {
		pruneTypes[ig.GroupVersionKind()] = ig.Informer()
	}

	o.pruneTypes = pruneTypes
	return o
}

func (o desiredSet) WithPatcher(gvk schema.GroupVersionKind, patcher Patcher) Apply {
	patchers := map[schema.GroupVersionKind]Patcher{}
	for k, v := range o.patchers {
		patchers[k] = v
	}
	patchers[gvk] = patcher
	o.patchers = patchers
	return o
}

func (o desiredSet) WithStrictCaching() Apply {
	o.strictCaching = true
	return o
}

func (o desiredSet) WithDefaultNamespace(ns string) Apply {
	o.defaultNamespace = ns
	return o
}
