package publicdomain

import (
	"context"
	"fmt"
	"strings"

	errors2 "github.com/pkg/errors"
	"github.com/rancher/rio/pkg/apply"
	"github.com/rancher/rio/pkg/certs"
	"github.com/rancher/rio/pkg/deploy/stack/populate/istio"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	spacev1beta1 "github.com/rancher/rio/types/apis/space.cattle.io/v1beta1"
	corev1 "github.com/rancher/types/apis/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
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
		secrets:          rContext.Core.Secrets(settings.RioSystemNamespace),
	}
	rContext.Global.PublicDomains(settings.RioSystemNamespace).AddLifecycle(ctx, "public-domain-controller", dc)
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
	secrets          corev1.SecretInterface
}

func (d *domainController) Create(domain *spacev1beta1.PublicDomain) (runtime.Object, error) {
	return domain, nil
}

func (d *domainController) Updated(domain *spacev1beta1.PublicDomain) (runtime.Object, error) {
	if err := apply.Apply([]runtime.Object{certs.AcmeIssuer()}, nil, "", "acme-cluster-issuer"); err != nil {
		return domain, err
	}
	ns, err := d.getNamespace(domain)
	if err != nil {
		return domain, err
	}
	// certificate
	cert := certs.CertificateHttp(domain)
	if err := apply.Apply([]runtime.Object{cert}, nil, settings.RioSystemNamespace, domain.Spec.DomainName); err != nil {
		return domain, err
	}
	service, err := d.serviceLister.Get(ns, domain.Spec.TargetName)
	if err != nil && !errors.IsNotFound(err) {
		return domain, err
	}
	if err == nil {
		if hasKey(service.Annotations[istio.PublicDomainAnnotation], domain.Spec.DomainName) {
			return domain, nil
		}
		service.Annotations[istio.PublicDomainAnnotation] = addKey(service.Annotations[istio.PublicDomainAnnotation], domain.Spec.DomainName)
		_, err = d.services.Services(ns).Update(service)
		return domain, err
	}
	routeset, err := d.routesetLister.Get(ns, domain.Spec.TargetName)
	if err != nil && !errors.IsNotFound(err) {
		return domain, err
	}
	if errors.IsNotFound(err) {
		return domain, errors2.Errorf("can't find target service or route %s", domain.Spec.TargetName)
	}
	if hasKey(routeset.Annotations[istio.PublicDomainAnnotation], domain.Spec.DomainName) {
		return domain, nil
	}
	routeset.Annotations[istio.PublicDomainAnnotation] = addKey(routeset.Annotations[istio.PublicDomainAnnotation], domain.Spec.DomainName)
	_, err = d.routesets.RouteSets(ns).Update(routeset)
	return domain, err
}

func (d *domainController) Remove(domain *spacev1beta1.PublicDomain) (runtime.Object, error) {
	ns, err := d.getNamespace(domain)
	if err != nil {
		return domain, err
	}
	service, err := d.serviceLister.Get(ns, domain.Spec.TargetName)
	if err != nil && !errors.IsNotFound(err) {
		return domain, err
	}
	if err == nil {
		service.Annotations[istio.PublicDomainAnnotation] = rmKey(service.Annotations[istio.PublicDomainAnnotation], domain.Spec.DomainName)
		if service.Annotations[istio.PublicDomainAnnotation] == "" {
			delete(service.Annotations, istio.PublicDomainAnnotation)
		}
		_, err = d.services.Services(ns).Update(service)
		return domain, err
	}
	routeset, err := d.routesetLister.Get(ns, domain.Spec.TargetName)
	if err != nil {
		return domain, err
	}
	if errors.IsNotFound(err) {
		return domain, nil
	}
	routeset.Annotations[istio.PublicDomainAnnotation] = rmKey(routeset.Annotations[istio.PublicDomainAnnotation], domain.Spec.DomainName)
	if routeset.Annotations[istio.PublicDomainAnnotation] == "" {
		delete(routeset.Annotations, istio.PublicDomainAnnotation)
	}
	if _, err := d.routesets.RouteSets(ns).Update(routeset); err != nil {
		return domain, err
	}
	if err := d.secrets.Delete(fmt.Sprintf("%s-tls-certs", domain.Name), &metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		return domain, err
	}
	return domain, nil
}

func (d *domainController) getNamespace(domain *spacev1beta1.PublicDomain) (string, error) {
	r, err := labels.NewRequirement("field.cattle.io/displayName", selection.Equals, []string{domain.Spec.TargetWorkspaceName})
	if err != nil {
		return "", err
	}
	namespaces, err := d.namespacesLister.List("", labels.NewSelector().Add(*r))
	if err != nil {
		return "", err
	}
	for _, ns := range namespaces {
		return namespace.StackNamespace(ns.Name, domain.Spec.TargetStackName), nil
	}
	return "", fmt.Errorf("can't find associated stack")
}

func hasKey(values, key string) bool {
	for _, v := range strings.Split(values, ",") {
		if key == v {
			return true
		}
	}
	return false
}

func addKey(values, key string) string {
	if !hasKey(values, key) {
		values += "," + key
	}
	return strings.Trim(values, ",")
}

func rmKey(values, key string) string {
	r := []string{}
	for _, v := range strings.Split(values, ",") {
		if key != v {
			r = append(r, v)
		}
	}
	return strings.Join(r, ",")
}
