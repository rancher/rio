package service

import (
	"context"
	"fmt"

	"github.com/knative/build/pkg/apis/build/v1alpha1"
	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/modules/istio/pkg/domains"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/rancher/wrangler/pkg/relatedresource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	buildkitAddr = "buildkit"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "stack-service-build", rContext.Rio.Rio().V1().Service())
	c.Apply = c.Apply.WithCacheTypes(
		rContext.Build.Build().V1alpha1().Build(),
		rContext.Webhook.Gitwatcher().V1().GitWatcher(),
		rContext.Core.Core().V1().ServiceAccount(),
		rContext.Core.Core().V1().Secret(),
	).WithStrictCaching()

	p := populator{
		systemNamespace:    rContext.Namespace,
		secretsCache:       rContext.Core.Core().V1().Secret().Cache(),
		clusterDomainCache: rContext.Global.Admin().V1().ClusterDomain().Cache(),
		serviceCache:       rContext.Rio.Rio().V1().Service().Cache(),
	}

	c.Populator = p.populate

	relatedresource.Watch(ctx, "webhook-service", p.resolve,
		rContext.Rio.Rio().V1().Service(), rContext.Rio.Rio().V1().Service())

	return nil
}

type populator struct {
	systemNamespace    string
	customRegistry     string
	secretsCache       corev1controller.SecretCache
	serviceCache       v1.ServiceCache
	clusterDomainCache projectv1controller.ClusterDomainCache
}

func (p *populator) isWebhook(obj runtime.Object) bool {
	if s, ok := obj.(*riov1.Service); ok {
		return s.Namespace == p.systemNamespace && s.Name == "webhook"
	}
	return false
}

func (p *populator) resolve(namespace, name string, obj runtime.Object) (result []relatedresource.Key, err error) {
	if !p.isWebhook(obj) {
		return nil, nil
	}

	svcs, err := p.serviceCache.List("", labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, svc := range svcs {
		if p.isWebhook(svc) {
			continue
		}
		result = append(result, relatedresource.Key{
			Namespace: svc.Namespace,
			Name:      svc.Name,
		})
	}

	return
}

func (p *populator) populate(obj runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	service := obj.(*riov1.Service)

	if service == nil || service.Spec.Build == nil || service.Spec.Build.Repo == "" {
		return nil
	}

	clusterDomain, err := p.clusterDomainCache.Get(p.systemNamespace, constants.ClusterDomainName)
	if err != nil {
		return err
	}

	if err := populateBuild(service, p.customRegistry, p.systemNamespace, clusterDomain.Status.ClusterDomain, os); err != nil {
		return err
	}

	webhook, err := p.serviceCache.Get(p.systemNamespace, "webhook")
	if errors.IsNotFound(err) {
		webhook = nil
	} else if err != nil {
		return err
	}

	populateWebhookAndSecrets(webhook, service, os)
	return nil
}

func populateBuild(service *riov1.Service, customRegistry, systemNamespace, domain string, os *objectset.ObjectSet) error {
	// we only support setting imageBuild for primary container
	rev := service.Spec.Build.Revision
	if rev == "" {
		return nil
	}

	build := constructors.NewBuild(service.Namespace, name.SafeConcatName(service.Name, name.Hex(service.Spec.Build.Repo, 5), rev), v1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"service-name":      service.Name,
				"service-namespace": service.Namespace,
			},
		},
		Spec: v1alpha1.BuildSpec{
			//ServiceAccountName: serviceAccountName,
			Source: &v1alpha1.SourceSpec{
				Git: &v1alpha1.GitSourceSpec{
					Url:      service.Spec.Build.Repo,
					Revision: rev,
				},
			},
			Template: &v1alpha1.TemplateInstantiationSpec{
				Kind: "ClusterBuildTemplate",
				Name: "buildkit",
				Arguments: []v1alpha1.ArgumentSpec{
					{
						Name:  "IMAGE",
						Value: ImageName(customRegistry, systemNamespace, rev, domain, service),
					},
					{
						Name:  "BUILDKIT_DAEMON_ADDRESS",
						Value: fmt.Sprintf("tcp://%s.%s:9001", buildkitAddr, systemNamespace),
					},
				},
			},
		},
	})
	os.Add(build)
	return nil
}

func populateWebhookAndSecrets(webhookService *riov1.Service, service *riov1.Service, os *objectset.ObjectSet) {
	webhookReceiver := webhookv1.NewGitWatcher(service.Namespace, service.Name, webhookv1.GitWatcher{
		Spec: webhookv1.GitWatcherSpec{
			RepositoryURL:                  service.Spec.Build.Repo,
			Enabled:                        true,
			Push:                           true,
			Tag:                            true,
			Branch:                         service.Spec.Build.Branch,
			RepositoryCredentialSecretName: service.Spec.Build.Secret,
		},
	})

	if webhookService != nil && len(webhookService.Status.Endpoints) > 0 {
		webhookReceiver.Spec.ReceiverURL = webhookService.Status.Endpoints[0]
	}

	os.Add(webhookReceiver)
}

func ImageName(customeRegistry, registryNamespace, rev, domain string, service *riov1.Service) string {
	var registryAddr string
	if customeRegistry == "" {
		registryAddr = domains.GetExternalDomain("registry", registryNamespace, domain)
		return fmt.Sprintf("%s/%s:%s", registryAddr, service.Namespace+"/"+service.Name, rev)
	}
	return fmt.Sprintf("%s/%s:%s", customeRegistry, service.Namespace+"-"+service.Name, rev)
}
