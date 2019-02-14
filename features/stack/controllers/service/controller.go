package service

import (
	"context"
	"fmt"
	os1 "os"

	buildapis "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/features/routing/pkg/istio/config"
	"github.com/rancher/rio/features/stack/controllers/service/populate"
	name1 "github.com/rancher/rio/pkg/name"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	buildv1alpha1 "github.com/rancher/rio/types/apis/build.knative.dev/v1alpha1"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	webhookv1 "github.com/rancher/rio/types/apis/webhookinator.rio.cattle.io/v1"
	corev1 "github.com/rancher/types/apis/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	buildkitAddr = "10.43.0.100"
	registryAddr = "10.43.0.101"
)

func Register(ctx context.Context, rContext *types.Context) error {
	cf := config.NewConfigFactory(ctx, rContext.Core.ConfigMap.Interface(),
		settings.IstioExternalLBNamespace,
		settings.IstionConfigMapName,
		settings.IstionConfigMapKey)
	injector := config.NewIstioInjector(cf)

	c := stackobject.NewGeneratingController(ctx, rContext, "stack-service", rContext.Rio.Service, *injector)
	c.Processor.Client(
		rContext.Build.Build,
		rContext.RBAC.Role,
		rContext.RBAC.RoleBinding,
		rContext.RBAC.ClusterRole,
		rContext.RBAC.ClusterRoleBinding,
		rContext.Apps.DaemonSet,
		rContext.Apps.Deployment,
		rContext.Apps.StatefulSet,
		rContext.Policy.PodDisruptionBudget,
		rContext.Core.ServiceAccount,
		rContext.Core.Service,
		rContext.Core.Secret,
		rContext.Core.ServiceAccount,
		rContext.AutoScale.ServiceScaleRecommendation,
		rContext.Webhook.GitWebHookReceiver)

	sh := &serviceHandler{
		serviceClient: rContext.Rio.Service,
		serviceCache:  rContext.Rio.Service.Cache(),
		configCache:   rContext.Rio.Config.Cache(),
		volumeCache:   rContext.Rio.Volume.Cache(),
		buildCache:    rContext.Build.Build.Cache(),
	}

	c.Populator = sh.populate
	rContext.Rio.Service.OnChange(ctx, "stack-service-change-controller", sh.onChange)

	return nil
}

type serviceHandler struct {
	serviceClient riov1.ServiceClient
	serviceCache  riov1.ServiceClientCache
	configCache   riov1.ConfigClientCache
	volumeCache   riov1.VolumeClientCache
	buildCache    buildv1alpha1.BuildClientCache
}

func (s *serviceHandler) onChange(service *riov1.Service) (runtime.Object, error) {
	if service.Spec.Revision.ParentService != "" {
		// enqueue parent so that we re-evaluate the destionationRules
		s.serviceClient.Enqueue(service.Namespace, service.Spec.Revision.ParentService)
	}

	return service, nil
}

func (s *serviceHandler) populate(obj runtime.Object, stack *riov1.Stack, os *objectset.ObjectSet) error {
	service := obj.(*riov1.Service)
	services, err := s.serviceCache.List(service.Namespace, labels.Everything())
	if err != nil {
		return err
	}

	configsByName := map[string]*riov1.Config{}
	configs, err := s.configCache.List(service.Namespace, labels.Everything())
	if err != nil {
		return err
	}
	for _, config := range configs {
		configsByName[config.Name] = config
	}

	volumesByName := map[string]*riov1.Volume{}
	volumes, err := s.volumeCache.List(service.Namespace, labels.Everything())
	if err != nil {
		return err
	}
	for _, volume := range volumes {
		volumesByName[volume.Name] = volume
	}

	if service.Spec.ImageBuild != nil {
		populateBuild(service.Namespace, service.Name, stack.Name, service.Spec.ImageBuild, os)
		populateWebhookAndSecrets(service.Namespace, service.Name, service.Spec.ImageBuild, os)
		service.Spec.Image = imageName(service.Name, stack.Name, service.Spec.ImageBuild)
	}

	return populate.Service(stack, configsByName, volumesByName, services, service, os)
}

func populateBuild(ns, name, stackName string, imageBuild *riov1.ImageBuild, os *objectset.ObjectSet) {
	build := buildv1alpha1.NewBuild(ns, fmt.Sprintf("%s-%s", name, hexName(imageBuild)), buildv1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"build-owner": name,
			},
		},
		Spec: buildapis.BuildSpec{
			Source: &buildapis.SourceSpec{
				Git: &buildapis.GitSourceSpec{
					Url:      imageBuild.Url,
					Revision: revision(imageBuild),
				},
			},
			Template: &buildapis.TemplateInstantiationSpec{
				Kind: "ClusterBuildTemplate",
				Name: imageBuild.Template,
				Arguments: []buildapis.ArgumentSpec{
					{
						Name:  "IMAGE",
						Value: imageName(name, stackName, imageBuild),
					},
					{
						Name:  "BUILDKIT_DAEMON_ADDRESS",
						Value: fmt.Sprintf("tcp://%s:9001", buildkitAddr),
					},
				},
			},
		},
	})
	os.Add(build)
}

func populateWebhookAndSecrets(ns, name string, imageBuild *riov1.ImageBuild, os *objectset.ObjectSet) {
	if imageBuild.Branch == "" || !imageBuild.TagOnly {
		return
	}
	webhookReceiver := webhookv1.NewGitWebHookReceiver(ns, name, webhookv1.GitWebHookReceiver{
		Spec: webhookv1.GitWebHookReceiverSpec{
			RepositoryURL:                  imageBuild.Url,
			Enabled:                        true,
			Push:                           true,
			Tag:                            true,
			RepositoryCredentialSecretName: fmt.Sprintf("webhook-secrets-%s", name1.Hex(imageBuild.Url, 5)),
		},
	})
	accessToken := imageBuild.WebhookAccessToken
	if accessToken == "" {
		accessToken = os1.Getenv("GITHUB_ACCESS_TOKEN")
	}
	secret := corev1.NewSecret(ns, fmt.Sprintf("webhook-secrets-%s", name1.Hex(imageBuild.Url, 5)), v1.Secret{
		Data: map[string][]byte{
			"accessToken": []byte(imageBuild.WebhookAccessToken),
		},
	})
	os.Add(webhookReceiver)
	os.Add(secret)
}

func imageName(name, stackName string, build *riov1.ImageBuild) string {
	tag := "latest"
	if build.Commit != "" {
		tag = build.Commit
	} else if build.Tag != "" {
		tag = build.Tag
	} else if build.Branch != "" {
		tag = build.Branch
	}
	return fmt.Sprintf("%s:5000/%s:%s", registryAddr, stackName+"/"+name, tag)
}

func revision(build *riov1.ImageBuild) string {
	if build.Commit != "" {
		return build.Commit
	} else if build.Tag != "" {
		return build.Tag
	} else if build.Branch != "" {
		return build.Branch
	}
	return ""
}

func hexName(build *riov1.ImageBuild) string {
	return name1.Hex(build.Url+revision(build), 5)
}
