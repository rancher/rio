package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/rancher/rio/features/routing/pkg/domains"

	"github.com/rancher/rio/pkg/namespace"

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	buildkitAddr = "buildkit.build"
)

func Register(ctx context.Context, rContext *types.Context) error {
	cf := config.NewConfigFactory(ctx, rContext.Core.ConfigMap.Interface(),
		settings.IstioStackName,
		settings.MeshConfigMapName,
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
	if len(service.Spec.ContainerConfig.PortBindings) > 0 {
		service.Status.Endpoints = []riov1.Endpoint{
			{
				URL: "https://" + domains.GetExternalDomain(service.Name, service.Spec.StackName, service.Spec.ProjectName),
			},
		}
	}
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
		if err := populateBuild(service.Name, stack, service.Spec.ImageBuild, os); err != nil {
			return err
		}
		populateWebhookAndSecrets(service.Namespace, service.Name, service.Spec.ImageBuild, os)
		service.Spec.Image = imageName(service.Name, stack, service.Spec.ImageBuild)

		// always pull images if tag or commit is missing
		if service.Spec.ImageBuild.Tag == "" && service.Spec.ImageBuild.Commit == "" {
			service.Spec.ImagePullPolicy = "always"
		}
	}

	return populate.Service(stack, configsByName, volumesByName, services, service, os)
}

func populateBuild(name string, stack *riov1.Stack, imageBuild *riov1.ImageBuild, os *objectset.ObjectSet) error {
	if settings.ClusterDomain.Get() == "" {
		return errors.New("image build need cluster domain to be set")
	}
	buildName := namespace.NameRef(name, stack)
	build := buildv1alpha1.NewBuild(stack.Name, fmt.Sprintf("%s-%s", buildName, hexName(imageBuild)), buildv1alpha1.Build{
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
				Name: "buildkit",
				Arguments: []buildapis.ArgumentSpec{
					{
						Name:  "IMAGE",
						Value: imageName(name, stack, imageBuild),
					},
					{
						Name:  "BUILDKIT_DAEMON_ADDRESS",
						Value: fmt.Sprintf("tcp://%s.%s.rio.local:9001", buildkitAddr, stack.Namespace),
					},
				},
			},
		},
	})
	os.Add(build)
	return nil
}

func populateWebhookAndSecrets(ns, name string, imageBuild *riov1.ImageBuild, os *objectset.ObjectSet) {
	if imageBuild.Branch == "" || !imageBuild.Hook {
		return
	}
	webhookReceiver := webhookv1.NewGitWebHookReceiver(ns, name, webhookv1.GitWebHookReceiver{
		Spec: webhookv1.GitWebHookReceiverSpec{
			RepositoryURL:                  imageBuild.Url,
			Enabled:                        true,
			Push:                           true,
			Tag:                            true,
			RepositoryCredentialSecretName: imageBuild.Secret,
		},
	})
	os.Add(webhookReceiver)
}

func imageName(name string, stack *riov1.Stack, build *riov1.ImageBuild) string {
	tag := "latest"
	if build.Commit != "" {
		tag = build.Commit
	} else if build.Tag != "" {
		tag = build.Tag
	} else if build.Branch != "" {
		tag = build.Branch
	}
	registryAddr := namespace.HashIfNeed("registry", "build", stack.Namespace) + "." + settings.ClusterDomain.Get()
	return fmt.Sprintf("%s/%s:%s", registryAddr, stack.Name+"/"+name, tag)
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
