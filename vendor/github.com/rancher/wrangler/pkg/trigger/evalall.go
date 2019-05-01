package trigger

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/relatedresource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	counter int64
)

type AllHandler func() error

type Controller interface {
	AddGenericHandler(ctx context.Context, name string, handler generic.Handler)
	GroupVersionKind() schema.GroupVersionKind
	Enqueue(namespace, name string)
}

type Trigger interface {
	Trigger()
	OnTrigger(ctx context.Context, name string, handler AllHandler)
	Key() relatedresource.Key
}

type trigger struct {
	key        string
	controller Controller
}

func New(controller Controller) Trigger {
	return &trigger{
		key:        fmt.Sprintf("__trigger__%d__", atomic.AddInt64(&counter, 1)),
		controller: controller,
	}
}

func (e *trigger) Key() relatedresource.Key {
	return relatedresource.Key{
		Namespace: "__trigger__",
		Name:      e.key,
	}
}

func (e *trigger) Trigger() {
	e.controller.Enqueue("__trigger__", e.key)
}

func (e *trigger) OnTrigger(ctx context.Context, name string, handler AllHandler) {
	e.controller.AddGenericHandler(ctx, name, func(queueKey string, _ runtime.Object) (runtime.Object, error) {
		if queueKey == "__trigger__/"+e.key {
			return nil, handler()
		}
		return nil, nil
	})
	e.Trigger()
}
