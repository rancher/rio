package generic

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	finalizerKey = "wrangler.cattle.io/"
)

type Updater func(runtime.Object) (runtime.Object, error)

type objectLifecycleAdapter struct {
	name    string
	handler Handler
	updater Updater
}

func NewRemoveHandler(name string, updater Updater, handler Handler) Handler {
	o := objectLifecycleAdapter{
		name:    name,
		handler: handler,
		updater: updater,
	}
	return o.sync
}

func (o *objectLifecycleAdapter) sync(key string, obj runtime.Object) (runtime.Object, error) {
	if obj == nil {
		return nil, nil
	}

	metadata, err := meta.Accessor(obj)
	if err != nil {
		return obj, err
	}

	if metadata.GetDeletionTimestamp() == nil {
		return o.addFinalizer(obj)
	}

	finalizerKey := o.constructFinalizerKey()
	finalizers := metadata.GetFinalizers()
	if len(finalizers) > 0 && finalizers[0] == finalizerKey {
		return obj, nil
	}

	newObj, err := o.handler(key, obj)
	if err != nil {
		return newObj, err
	}

	if newObj != nil {
		obj = newObj
	}

	obj = obj.DeepCopyObject()
	metadata, err = meta.Accessor(obj)
	if err != nil {
		return obj, err
	}

	metadata.SetFinalizers(finalizers[1:])
	return o.updater(obj)
}

func (o *objectLifecycleAdapter) constructFinalizerKey() string {
	return finalizerKey + o.name
}

func (o *objectLifecycleAdapter) addFinalizer(obj runtime.Object) (runtime.Object, error) {
	metadata, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}

	finalizerKey := o.constructFinalizerKey()
	finalizers := metadata.GetFinalizers()
	for _, finalizer := range finalizers {
		if finalizer == finalizerKey {
			return obj, nil
		}
	}

	obj = obj.DeepCopyObject()
	metadata, err = meta.Accessor(obj)
	if err != nil {
		return nil, err
	}

	metadata.SetFinalizers(append(finalizers, finalizerKey))
	return o.updater(obj)
}
