package publicdomain

import (
	"context"
	"fmt"

	"github.com/rancher/rio/pkg/deploy/stack/populate/istio"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	spacev1beta1 "github.com/rancher/rio/types/apis/space.cattle.io/v1beta1"
	corev1 "github.com/rancher/types/apis/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func Register(ctx context.Context, rContext *types.Context) {
	dc := &domainController{
		namespacesLister: rContext.Core.Namespaces("").Controller().Lister(),
		stackLister:      rContext.Rio.Stacks("").Controller().Lister(),
		serviceLister:    rContext.Rio.Services("").Controller().Lister(),
		services:         rContext.Rio,
		domains:          rContext.Global.PublicDomains(settings.RioSystemNamespace),
		domainLister:     rContext.Global.PublicDomains(settings.RioSystemNamespace).Controller().Lister(),
		routesetLister:   rContext.Rio.RouteSets("").Controller().Lister(),
		routesets:        rContext.Rio,
	}
	rContext.Global.PublicDomains(settings.RioSystemNamespace).AddLifecycle("public-domain-controller", dc)
}

type domainController struct {
	namespacesLister corev1.NamespaceLister
	stackLister      v1beta1.StackLister
	serviceLister    v1beta1.ServiceLister
	services         v1beta1.ServicesGetter
	domains          spacev1beta1.PublicDomainInterface
	domainLister     spacev1beta1.PublicDomainLister
	routesetLister   v1beta1.RouteSetLister
	routesets        v1beta1.RouteSetsGetter
}

func (d *domainController) Create(domain *spacev1beta1.PublicDomain) (*spacev1beta1.PublicDomain, error) {
	return domain, nil
}

func (d *domainController) Updated(domain *spacev1beta1.PublicDomain) (*spacev1beta1.PublicDomain, error) {
	ns, err := d.getNamespace(domain)
	if err != nil {
		return domain, err
	}
	if domain.Spec.ServiceName != "" {
		service, err := d.serviceLister.Get(ns, domain.Spec.ServiceName)
		if err != nil {
			return domain, err
		}
		if service.Annotations[istio.PublicDomainAnnotation] != domain.Spec.DomainName {
			service.Annotations[istio.PublicDomainAnnotation] = domain.Spec.DomainName
		} else {
			return domain, nil
		}
		_, err = d.services.Services(ns).Update(service)
		return domain, err
	} else if domain.Spec.RouteSetName != "" {
		routeset, err := d.routesetLister.Get(ns, domain.Spec.RouteSetName)
		if err != nil {
			return domain, err
		}
		if routeset.Annotations[istio.PublicDomainAnnotation] != domain.Spec.DomainName {
			routeset.Annotations[istio.PublicDomainAnnotation] = domain.Spec.DomainName
		} else {
			return domain, nil
		}
		_, err = d.routesets.RouteSets(ns).Update(routeset)
		return domain, err
	}
	return domain, nil
}

func (d *domainController) Remove(domain *spacev1beta1.PublicDomain) (*spacev1beta1.PublicDomain, error) {
	ns, err := d.getNamespace(domain)
	if err != nil {
		return domain, err
	}
	if domain.Spec.ServiceName != "" {
		service, err := d.serviceLister.Get(ns, domain.Spec.ServiceName)
		if err != nil {
			return domain, err
		}
		delete(service.Annotations, istio.PublicDomainAnnotation)
		_, err = d.services.Services(ns).Update(service)
		return domain, err
	} else if domain.Spec.RouteSetName != "" {
		routeset, err := d.routesetLister.Get(ns, domain.Spec.RouteSetName)
		if err != nil {
			return domain, err
		}
		delete(routeset.Annotations, istio.PublicDomainAnnotation)
		_, err = d.routesets.RouteSets(ns).Update(routeset)
		return domain, err
	}
	return domain, nil
}

func (d *domainController) getNamespace(domain *spacev1beta1.PublicDomain) (string, error) {
	namespaces, err := d.namespacesLister.List("", labels.Everything())
	if err != nil {
		return "", err
	}
	for _, ns := range namespaces {
		if ns != nil && ns.Labels["field.cattle.io/displayName"] == domain.Spec.SpaceName {
			stack, err := d.stackLister.Get(ns.Name, domain.Spec.StackName)
			if err != nil {
				return "", err
			}
			namespace := namespace.StackToNamespace(stack)
			return namespace, nil
		}
	}
	return "", fmt.Errorf("can't find associated stack")
}
