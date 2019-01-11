package publicdomain

import (
	"context"
	"fmt"
	"strings"

	errors2 "github.com/pkg/errors"
	"github.com/rancher/rio/features/routing/controllers/service/populate"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	corev1 "github.com/rancher/types/apis/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
)

func Register(ctx context.Context, rContext *types.Context) error {
	dc := &domainController{
		namespacesLister:  rContext.Core.Namespace.Cache(),
		stackLister:       rContext.Rio.Stack.Cache(),
		serviceLister:     rContext.Rio.Service.Cache(),
		services:          rContext.Rio.Service,
		routesetLister:    rContext.Rio.RouteSet.Cache(),
		routesets:         rContext.Rio.RouteSet,
		secrets:           rContext.Core.Secret,
		featureController: rContext.Global.Feature,
	}
	rContext.Global.PublicDomain.OnChange(ctx, "public-domain-controller", dc.Updated)
	rContext.Global.PublicDomain.OnRemove(ctx, "public-domain-controller", dc.Remove)
	return nil
}

type domainController struct {
	namespacesLister  corev1.NamespaceClientCache
	stackLister       riov1.StackClientCache
	serviceLister     riov1.ServiceClientCache
	services          riov1.ServiceClient
	routesetLister    riov1.RouteSetClientCache
	routesets         riov1.RouteSetClient
	secrets           corev1.SecretClient
	featureController projectv1.FeatureClient
}

func (d *domainController) Updated(domain *projectv1.PublicDomain) (runtime.Object, error) {
	if domain.Namespace != settings.RioSystemNamespace {
		return domain, nil
	}
	fmt.Println("!!!!!! enqueue")
	d.featureController.Enqueue("", "letsencrypt")

	ns, err := d.getNamespace(domain)
	if err != nil {
		return domain, err
	}

	service, err := d.serviceLister.Get(ns, domain.Spec.TargetName)
	if err != nil && !errors.IsNotFound(err) {
		return domain, err
	}
	if err == nil {
		if hasKey(service.Annotations[populate.PublicDomainAnnotation], domain.Spec.DomainName) {
			return domain, nil
		}
		service = service.DeepCopy()
		service.Annotations[populate.PublicDomainAnnotation] = addKey(service.Annotations[populate.PublicDomainAnnotation], domain.Spec.DomainName)
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
	if hasKey(routeset.Annotations[populate.PublicDomainAnnotation], domain.Spec.DomainName) {
		return domain, nil
	}

	routeset = routeset.DeepCopy()
	routeset.Annotations[populate.PublicDomainAnnotation] = addKey(routeset.Annotations[populate.PublicDomainAnnotation], domain.Spec.DomainName)
	_, err = d.routesets.Update(routeset)
	return domain, err
}

func (d *domainController) Remove(domain *projectv1.PublicDomain) (runtime.Object, error) {
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
		service.Annotations[populate.PublicDomainAnnotation] = rmKey(service.Annotations[populate.PublicDomainAnnotation], domain.Spec.DomainName)
		if service.Annotations[populate.PublicDomainAnnotation] == "" {
			delete(service.Annotations, populate.PublicDomainAnnotation)
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
		routeset.Annotations[populate.PublicDomainAnnotation] = rmKey(routeset.Annotations[populate.PublicDomainAnnotation], domain.Spec.DomainName)
		if routeset.Annotations[populate.PublicDomainAnnotation] == "" {
			delete(routeset.Annotations, populate.PublicDomainAnnotation)
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

func (d *domainController) getNamespace(domain *projectv1.PublicDomain) (string, error) {
	r, err := labels.NewRequirement("field.cattle.io/displayName", selection.Equals, []string{domain.Spec.TargetProjectName})
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
		if key == strings.TrimSpace(v) {
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
		if key != strings.TrimSpace(v) {
			r = append(r, v)
		}
	}
	return strings.Join(r, ",")
}
