package clients

import (
	"context"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
)

//go:generate mockgen -destination=./mocks/client_interface.go -source client_interface.go -package mocks

const DefaultNamespace = "default"

var DefaultRefreshRate = time.Second * 30

func DefaultNamespaceIfEmpty(namespace string) string {
	if namespace == "" {
		return DefaultNamespace
	}
	return namespace
}

type ResourceWatch func(ctx context.Context) (<-chan resources.ResourceList, <-chan error, error)

type ResourceWatcher interface {
	Watch(namespace string, opts WatchOpts) (<-chan resources.ResourceList, <-chan error, error)
}

type ResourceClient interface {
	Kind() string
	NewResource() resources.Resource
	// Deprecated: implemented only by the kubernetes resource client. Will be removed from the interface.
	Register() error
	Read(namespace, name string, opts ReadOpts) (resources.Resource, error)
	Write(resource resources.Resource, opts WriteOpts) (resources.Resource, error)
	Delete(namespace, name string, opts DeleteOpts) error
	List(namespace string, opts ListOpts) (resources.ResourceList, error)
	ResourceWatcher
}

type ResourceClients map[string]ResourceClient

func (r ResourceClients) Add(rcs ...ResourceClient) {
	for _, rc := range rcs {
		r[rc.Kind()] = rc
	}
}

func (r ResourceClients) ForResource(resource resources.Resource) (ResourceClient, error) {
	return r.ForKind(resources.Kind(resource))
}

func (r ResourceClients) ForKind(kind string) (ResourceClient, error) {
	rc, ok := r[kind]
	if !ok {
		return nil, errors.Errorf("no resource client registered for kind %v", kind)
	}
	return rc, nil
}

type ReadOpts struct {
	Ctx     context.Context
	Cluster string
}

func (o ReadOpts) WithDefaults() ReadOpts {
	if o.Ctx == nil {
		o.Ctx = context.TODO()
	}
	return o
}

type StorageWriteOpts interface {
	StorageWriteOptsTag()
}

type WriteOpts struct {
	Ctx               context.Context
	OverwriteExisting bool

	// Implementation dependant write opts
	StorageWriteOpts StorageWriteOpts
}

func (o WriteOpts) WithDefaults() WriteOpts {
	if o.Ctx == nil {
		o.Ctx = context.TODO()
	}
	return o
}

type DeleteOpts struct {
	Ctx            context.Context
	IgnoreNotExist bool
	Cluster        string
}

func (o DeleteOpts) WithDefaults() DeleteOpts {
	if o.Ctx == nil {
		o.Ctx = context.TODO()
	}
	return o
}

type ListOpts struct {
	Ctx      context.Context
	Selector map[string]string
	Cluster  string
}

func (o ListOpts) WithDefaults() ListOpts {
	if o.Ctx == nil {
		o.Ctx = context.TODO()
	}
	return o
}

// RefreshRate is currently ignored by the Kubernetes ResourceClient implementation.
// To achieve a similar behavior you can use the KubeResourceClientFactory.ResyncPeriod field. The difference is that it
// will apply to all the watches started by clients built with the factory.
type WatchOpts struct {
	Ctx         context.Context
	Selector    map[string]string
	RefreshRate time.Duration
	// Cluster is ignored by aggregated watches, but is respected by multi cluster clients.
	Cluster string
}

func (o WatchOpts) WithDefaults() WatchOpts {
	if o.Ctx == nil {
		o.Ctx = context.TODO()
	}
	if o.RefreshRate == 0 {
		o.RefreshRate = DefaultRefreshRate
	}
	return o
}
