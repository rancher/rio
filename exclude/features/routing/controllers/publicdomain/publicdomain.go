package publicdomain

import (
	"context"
	"fmt"
	"strings"

	errors2 "github.com/pkg/errors"
	"github.com/rancher/rio/features/routing/controllers/service/populate"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	corev1controller "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	dc := &domainController{
		namespace:         rContext.Namespace,
		namespacesLister:  rContext.Core.Core().V1().Namespace().Cache(),
		stackLister:       rContext.Rio.Rio().V1().Stack().Cache(),
		serviceLister:     rContext.Rio.Rio().V1().Service().Cache(),
		services:          rContext.Rio.Rio().V1().Service(),
		routesetLister:    rContext.Rio.Rio().V1().Router().Cache(),
		routesets:         rContext.Rio.Rio().V1().Router(),
		secrets:           rContext.Core.Core().V1().Secret(),
		featureController: rContext.Global.Project().V1().Feature(),
	}
	rContext.Global.Project().V1().PublicDomain().OnChange(ctx, "public-domain-controller", dc.Updated)
	rContext.Global.Project().V1().PublicDomain().OnRemove(ctx, "public-domain-controller", dc.Remove)
	return nil
}

type domainController struct {
	namespace         string
	namespacesLister  corev1controller.NamespaceCache
	stackLister       riov1controller.StackCache
	serviceLister     riov1controller.ServiceCache
	services          riov1controller.ServiceClient
	routesetLister    riov1controller.RouterCache
	routesets         riov1controller.RouterClient
	secrets           corev1controller.SecretClient
	featureController projectv1controller.FeatureController
}

func (d *domainController) Updated(key string, domain *projectv1.PublicDomain) (*projectv1.PublicDomain, error) {
	if domain == nil {
		return nil, nil
	}

	if domain.Namespace != d.namespace {
		return domain, nil
	}

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

func (d *domainController) Remove(key string, domain *projectv1.PublicDomain) (*projectv1.PublicDomain, error) {
	if domain.Namespace != d.namespace {
		return domain, nil
	}

	ns, err := d.getNamespace(domain)
	if err != nil {
		return domain, err
	}

	service, err := d.serviceLister.Get(ns, domain.Spec.TargetName)
	if err != nil && !errors.IsNotFound(err) {
		return domain, err
	} else if err == nil {
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
	} else if err == nil {
		routeset = routeset.DeepCopy()
		routeset.Annotations[populate.PublicDomainAnnotation] = rmKey(routeset.Annotations[populate.PublicDomainAnnotation], domain.Spec.DomainName)
		if routeset.Annotations[populate.PublicDomainAnnotation] == "" {
			delete(routeset.Annotations, populate.PublicDomainAnnotation)
		}
		if _, err := d.routesets.Update(routeset); err != nil {
			return domain, err
		}
	}

	if err := d.secrets.Delete(d.namespace, fmt.Sprintf("%s-tls-certs", domain.Name), &metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		return domain, err
	}

	return domain, nil
}

func (d *domainController) getNamespace(domain *projectv1.PublicDomain) (string, error) {
	return domain.Namespace, nil
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
