package domain

import (
	"context"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/rancher/norman/pkg/changeset"
	"github.com/rancher/norman/types/slice"
	"github.com/rancher/rancher/pkg/controllers/user/approuter"
	"github.com/rancher/rancher/pkg/ticker"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	v12 "github.com/rancher/types/apis/core/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	local = "localhost.localdomain"
)

var (
	refreshInterval = 5 * time.Minute
	addressTypes    = []v1.NodeAddressType{
		v1.NodeExternalIP,
		v1.NodeInternalIP,
	}
)

type Controller struct {
	ctx             context.Context
	init            sync.Once
	rdnsClient      *approuter.Client
	endpointsLister v12.EndpointsLister
	nodeLister      v12.NodeLister
	stackController v1beta1.StackController
	previousIPs     []string
}

func Register(ctx context.Context, rContext *types.Context) error {
	rdnsClient := approuter.NewClient(rContext.Core.Secrets(""),
		rContext.Core.Secrets("").Controller().Lister(),
		settings.RioSystemNamespace)
	rdnsClient.SetBaseURL(settings.RDNSURL.Get())

	g := &Controller{
		ctx:             ctx,
		rdnsClient:      rdnsClient,
		endpointsLister: rContext.Core.Endpoints(settings.IstioExternalLBNamespace).Controller().Lister(),
		nodeLister:      rContext.Core.Nodes("").Controller().Lister(),
		stackController: rContext.Rio.Stacks("").Controller(),
	}

	changeset.Watch(ctx, "domain-controller",
		func(namespace, name string, obj runtime.Object) ([]changeset.Key, error) {
			switch t := obj.(type) {
			case *v1.Endpoints:
				if isLB(t) {
					return []changeset.Key{
						{
							Namespace: t.Namespace,
							Name:      t.Name,
						},
					}, nil
				}
			}
			return nil, nil
		},
		rContext.Core.Services(settings.IstioExternalLBNamespace).Controller().Enqueue,
		rContext.Core.Endpoints(settings.IstioExternalLBNamespace).Controller())

	rContext.Core.Services(settings.IstioExternalLBNamespace).Controller().AddHandler(ctx, "domain-controller", g.sync)

	return nil
}

func isLB(obj runtime.Object) bool {
	o, err := meta.Accessor(obj)
	if err != nil {
		return false
	}
	if o == nil || reflect.ValueOf(obj).IsNil() {
		return false
	}
	return o.GetName() == settings.IstioExternalLB && o.GetNamespace() == settings.IstioExternalLBNamespace
}

func (g *Controller) sync(key string, svc *v1.Service) (runtime.Object, error) {
	// We do init here because we need caches synced before we can initialize
	g.init.Do(func() {
		err := g.start()
		if err != nil {
			panic(err)
		}
	})

	if !isLB(svc) {
		return nil, nil
	}

	var ips []string
	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		if ingress.Hostname == "localhost" {
			ips = append(ips, "127.0.0.1")
		} else if ingress.IP != "" {
			ips = append(ips, ingress.IP)
		}
	}

	if len(ips) == 0 {
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
	}

	return nil, g.updateDomain(ips)
}

func (g *Controller) start() error {
	domain, err := g.rdnsClient.GetDomain()
	if err == nil {
		if domain != nil && domain.Fqdn != "" {
			return g.setDomain(domain.Fqdn)
		}
	} else {
		return g.setDomain(local)
	}

	go func() {
		g.renew()
		for range ticker.Context(g.ctx, 6*time.Hour) {
			g.renew()
		}
	}()

	return nil
}

func (g *Controller) renew() error {
	_, err := g.rdnsClient.RenewDomain()
	return err
}

func (g *Controller) setDomain(fqdn string) error {
	if settings.ClusterDomain.Get() == fqdn {
		return nil
	}

	settings.ClusterDomain.Set(fqdn)

	stacks, err := g.stackController.Lister().List("", labels.Everything())
	if err != nil {
		return err
	}

	for _, stack := range stacks {
		g.stackController.Enqueue(stack.Namespace, stack.Name)
	}

	return nil
}

func (g *Controller) updateDomain(ips []string) error {
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
		_, fqdn, err = g.rdnsClient.ApplyDomain(ips)
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
