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
		namespacesLister: rContext.Core.Namespace.Cache(),
		stackLister:      rContext.Rio.Stack.Cache(),
		serviceLister:    rContext.Rio.Service.Cache(),
		services:         rContext.Rio.Service,
		routesetLister:   rContext.Rio.RouteSet.Cache(),
		routesets:        rContext.Rio.RouteSet,
		secrets:          rContext.Core.Secret,
	}
	rContext.Global.PublicDomain.OnChange(ctx, "public-domain-controller", dc.Updated)
	rContext.Global.PublicDomain.OnRemove(ctx, "public-domain-controller", dc.Remove)
}

type domainController struct {
	namespacesLister corev1.NamespaceClientCache
	stackLister      v1beta1.StackClientCache
	serviceLister    v1beta1.ServiceClientCache
	services         v1beta1.ServiceClient
	routesetLister   v1beta1.RouteSetClientCache
	routesets        v1beta1.RouteSetClient
	secrets          corev1.SecretClient
}

func (d *domainController) Updated(domain *spacev1beta1.PublicDomain) (runtime.Object, error) {
	if domain.Namespace != settings.RioSystemNamespace {
		return domain, nil
	}

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
		service = service.DeepCopy()
		service.Annotations[istio.PublicDomainAnnotation] = addKey(service.Annotations[istio.PublicDomainAnnotation], domain.Spec.DomainName)
		_, err = d.services.Update(service)
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
	routeset = routeset.DeepCopy()
	routeset.Annotations[istio.PublicDomainAnnotation] = addKey(routeset.Annotations[istio.PublicDomainAnnotation], domain.Spec.DomainName)
	_, err = d.routesets.Update(routeset)
	return domain, err
}

func (d *domainController) Remove(domain *spacev1beta1.PublicDomain) (runtime.Object, error) {
	if domain.Namespace != settings.RioSystemNamespace {
		return domain, nil
	}

	ns, err := d.getNamespace(domain)
	if err != nil {
		return domain, err
	}
	service, err := d.serviceLister.Get(ns, domain.Spec.TargetName)
	if err != nil && !errors.IsNotFound(err) {
		return domain, err
	}
	if err == nil {
		service = service.DeepCopy()
		service.Annotations[istio.PublicDomainAnnotation] = rmKey(service.Annotations[istio.PublicDomainAnnotation], domain.Spec.DomainName)
		if service.Annotations[istio.PublicDomainAnnotation] == "" {
			delete(service.Annotations, istio.PublicDomainAnnotation)
		}
		_, err = d.services.Update(service)
		return domain, err
	}
	routeset, err := d.routesetLister.Get(ns, domain.Spec.TargetName)
	if err != nil && !errors.IsNotFound(err) {
		return domain, err
	}
	if err == nil {
		routeset = routeset.DeepCopy()
		routeset.Annotations[istio.PublicDomainAnnotation] = rmKey(routeset.Annotations[istio.PublicDomainAnnotation], domain.Spec.DomainName)
		if routeset.Annotations[istio.PublicDomainAnnotation] == "" {
			delete(routeset.Annotations, istio.PublicDomainAnnotation)
		}
		if _, err := d.routesets.Update(routeset); err != nil {
			return domain, err
		}
	}
	if err := d.secrets.Delete(settings.RioSystemNamespace, fmt.Sprintf("%s-tls-certs", domain.Name), &metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
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
	var r []string
	for _, v := range strings.Split(values, ",") {
		if key != v {
			r = append(r, v)
		}
	}
	return strings.Join(r, ",")
}
