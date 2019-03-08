package domain

import (
	"context"
	"reflect"
	"sort"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/selection"

	"github.com/rancher/norman/pkg/changeset"
	"github.com/rancher/norman/types/slice"
	"github.com/rancher/rancher/pkg/controllers/user/approuter"
	"github.com/rancher/rancher/pkg/ticker"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	v12 "github.com/rancher/types/apis/core/v1"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	local            = "localhost.localdomain"
	serviceNameLabel = "rio.cattle.io/service"
	stackNameLabel   = "rio.cattle.io/stack"
	registry         = "registry"
	build            = "build"
)

var (
	addressTypes = []v1.NodeAddressType{
		v1.NodeExternalIP,
		v1.NodeInternalIP,
	}

	nodeHasGateway = "nodeGatewayIndex"
)

type Controller struct {
	ctx               context.Context
	init              sync.Once
	rdnsClient        *approuter.Client
	servicesLister    v12.ServiceClientCache
	endpointsLister   v12.EndpointsClientCache
	nodeLister        v12.NodeClientCache
	stackLister       riov1.StackClientCache
	stackController   changeset.Enqueuer
	featureController changeset.Enqueuer
	previousIPs       []string
}

func Register(ctx context.Context, rContext *types.Context) error {
	rdnsClient := approuter.NewClient(rContext.Core.Secret.Interface(),
		rContext.Core.Secret.Interface().Controller().Lister(),
		settings.RioSystemNamespace)
	rdnsClient.SetBaseURL(settings.RDNSURL.Get())

	g := &Controller{
		ctx:               ctx,
		rdnsClient:        rdnsClient,
		servicesLister:    rContext.Core.Service.Cache(),
		endpointsLister:   rContext.Core.Endpoints.Cache(),
		nodeLister:        rContext.Core.Node.Cache(),
		stackLister:       rContext.Rio.Stack.Cache(),
		stackController:   rContext.Rio.Stack,
		featureController: rContext.Global.Feature,
	}

	rContext.Core.Endpoints.Cache().Index(nodeHasGateway, g.indexEPByNode)

	changeset.Watch(ctx, "domain-controller",
		g.resolve,
		rContext.Core.Service,
		rContext.Core.Node,
		rContext.Core.Endpoints)

	rContext.Core.Service.OnChange(ctx, "domain-controller", g.sync)

	return nil
}

func isGateway(obj runtime.Object) bool {
	o, err := meta.Accessor(obj)
	if err != nil {
		return false
	}
	if o == nil || reflect.ValueOf(obj).IsNil() {
		return false
	}
	return o.GetName() == settings.IstioGatewayDeploy && o.GetNamespace() == namespace.CloudNamespace
}

func isRegistry(svc *v1.Service) bool {
	return svc.Labels[serviceNameLabel] == registry && svc.Labels[stackNameLabel] == build
}

func (g *Controller) indexEPByNode(ep *v1.Endpoints) ([]string, error) {
	if !isGateway(ep) {
		return nil, nil
	}

	var result []string

	for _, subset := range ep.Subsets {
		for _, addr := range subset.Addresses {
			if addr.NodeName != nil {
				result = append(result, *addr.NodeName)
			}
		}
	}

	return result, nil
}

func (g *Controller) resolve(namespace, name string, obj runtime.Object) ([]changeset.Key, error) {
	switch t := obj.(type) {
	case *v1.Endpoints:
		if isGateway(t) {
			return []changeset.Key{
				{
					Namespace: t.Namespace,
					Name:      t.Name,
				},
			}, nil
		}
	case *v1.Node:
		eps, err := g.endpointsLister.GetIndexed(nodeHasGateway, t.Name)
		if err != nil || len(eps) == 0 {
			return nil, err
		}
		return []changeset.Key{
			{
				Namespace: eps[0].Namespace,
				Name:      eps[0].Name,
			},
		}, nil
	}
	return nil, nil
}

func (g *Controller) sync(svc *v1.Service) (runtime.Object, error) {
	if svc.Namespace != namespace.CloudNamespace {
		return nil, nil
	}

	// We do init here because we need caches synced before we can initialize
	g.init.Do(func() {
		err := g.start()
		if err != nil {
			panic(err)
		}
	})

	if !isGateway(svc) && !isRegistry(svc) {
		return nil, nil
	}

	var ips []string
	ep, err := g.endpointsLister.Get(svc.Namespace, svc.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	for _, subset := range ep.Subsets {
		for _, addr := range subset.Addresses {
			if addr.NodeName == nil {
				continue
			}

			node, err := g.nodeLister.Get("", *addr.NodeName)
			if err != nil {
				return nil, err
			}

			nodeIP := getNodeIP(node)
			if nodeIP != "" {
				ips = append(ips, nodeIP)
			}
		}
	}

	r1, err := labels.NewRequirement(serviceNameLabel, selection.In, []string{registry})
	if err != nil {
		return nil, err
	}
	r2, err := labels.NewRequirement(stackNameLabel, selection.In, []string{build})
	if err != nil {
		return nil, err
	}
	registryServices, err := g.servicesLister.List(settings.CloudNamespace, labels.NewSelector().Add(*r1, *r2))
	if err != nil {
		return nil, err
	}
	subDomains := map[string][]string{}
	for _, svc := range registryServices {
		subDomains[svc.Name] = []string{svc.Spec.ClusterIP}
	}

	if err := g.updateDomain(ips, subDomains); err != nil {
		return nil, err
	}

	return nil, nil
}

func (g *Controller) start() error {
	go func() {
		g.renew()
		for range ticker.Context(g.ctx, 6*time.Hour) {
			if err := g.renew(); err != nil {
				logrus.Errorf("failed to renew domain: %v", err)
			}
		}
	}()

	domain, err := g.rdnsClient.GetDomain()
	if err == nil {
		if domain != nil && domain.Fqdn != "" {
			return g.setDomain(domain.Fqdn)
		}
	} else {
		return g.setDomain(local)
	}

	return nil
}

func (g *Controller) renew() error {
	if _, err := g.rdnsClient.RenewDomain(); err != nil {
		return err
	}
	return nil
}

func (g *Controller) setDomain(fqdn string) error {
	if settings.ClusterDomain.Get() == fqdn {
		return nil
	}

	if err := settings.ClusterDomain.Set(fqdn); err != nil {
		return err
	}

	stacks, err := g.stackLister.List("", labels.Everything())
	if err != nil {
		return err
	}

	for _, stack := range stacks {
		g.stackController.Enqueue(stack.Namespace, stack.Name)
	}

	time.Sleep(time.Second)
	g.featureController.Enqueue(settings.RioSystemNamespace, "letsencrypt")
	return nil
}

func (g *Controller) updateDomain(ips []string, subDomains map[string][]string) error {
	var (
		fqdn string
		err  error
	)

	if len(ips) == 0 {
		return nil
	}

	sort.Strings(ips)
	if slice.StringsEqual(g.previousIPs, ips) {
		return nil
	}

	if len(ips) == 1 && ips[0] == "127.0.0.1" {
		fqdn = local
	} else {
		_, fqdn, err = g.rdnsClient.ApplyDomain(ips, subDomains)
		if err != nil {
			return err
		}
	}

	if err := g.setDomain(fqdn); err != nil {
		return err
	}
	g.previousIPs = ips

	return nil
}

func getNodeIP(node *v1.Node) string {
	for _, addrType := range addressTypes {
		for _, addr := range node.Status.Addresses {
			if addrType == addr.Type {
				return addr.Address
			}
		}
	}

	return ""
}
