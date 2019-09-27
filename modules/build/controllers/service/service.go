package service

import (
	"context"
	"fmt"
	"strconv"

	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/randomtoken"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/rancher/wrangler/pkg/relatedresource"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	DefaultGitCrendential    = "gitcredential"
	DefaultDockerCrendential = "dockerconfig"
	DefaultGithubCrendential = "githubtoken"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "stack-service-build", rContext.Rio.Rio().V1().Service())
	c.Apply = c.Apply.WithCacheTypes(
		rContext.Build.Tekton().V1alpha1().TaskRun(),
		rContext.Webhook.Gitwatcher().V1().GitWatcher(),
		rContext.Core.Core().V1().ServiceAccount(),
		rContext.Core.Core().V1().Secret(),
	).WithStrictCaching()

	p := populator{
		systemNamespace:    rContext.Namespace,
		appCache:           rContext.Rio.Rio().V1().App().Cache(),
		secretsCache:       rContext.Core.Core().V1().Secret().Cache(),
		clusterDomainCache: rContext.Global.Admin().V1().ClusterDomain().Cache(),
		serviceCache:       rContext.Rio.Rio().V1().Service().Cache(),
		services:           rContext.Rio.Rio().V1().Service(),
	}

	c.Populator = p.populate

	relatedresource.Watch(ctx, "webhook-service", p.resolve,
		rContext.Rio.Rio().V1().Service(), rContext.Rio.Rio().V1().Service())

	return nil
}

type populator struct {
	systemNamespace    string
	customRegistry     string
	appCache           v1.AppCache
	secretsCache       corev1controller.SecretCache
	serviceCache       v1.ServiceCache
	services           v1.ServiceClient
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

	if service.Status.BuildLogToken == "" {
		token, err := randomtoken.Generate()
		if err != nil {
			return err
		}
		service.Status.BuildLogToken = token
		newService, err := p.services.Update(service)
		if err != nil {
			return err
		}
		service = newService
	}

	if err := p.populateBuild(service, p.systemNamespace, os); err != nil {
		return err
	}

	webhook, err := p.appCache.Get(p.systemNamespace, "webhook")
	if errors.IsNotFound(err) {
		webhook = nil
	} else if err != nil {
		return err
	}

	populateWebhookAndSecrets(webhook, service, os)
	return nil
}

func (p populator) populateBuild(service *riov1.Service, systemNamespace string, os *objectset.ObjectSet) error {
	// we only support setting imageBuild for primary container
	rev := service.Spec.Build.Revision
	if rev == "" {
		rev = service.Status.FirstRevision
	}
	if rev == "" {
		return nil
	}

	trName := name.SafeConcatName(service.Namespace, service.Name, name.Hex(service.Spec.Build.Repo, 5), name.Hex(rev, 5))

	p.setDefaults(service)

	sa := constructors.NewServiceAccount(service.Namespace, trName, corev1.ServiceAccount{})
	if service.Spec.Build.GitSecretName != "" {
		sa.Secrets = append(sa.Secrets, corev1.ObjectReference{
			Name: service.Spec.Build.GitSecretName,
		})
	}
	if service.Spec.Build.PushRegistrySecretName != "" {
		sa.Secrets = append(sa.Secrets, corev1.ObjectReference{
			Name: service.Spec.Build.PushRegistrySecretName,
		})
	}

	build := constructors.NewTaskRun(service.Namespace, trName, tektonv1alpha1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"service-name":      service.Name,
				"service-namespace": service.Namespace,
				"gitcommit-name":    service.Status.GitCommitName,
				"log-token":         service.Status.BuildLogToken,
			},
		},
		Spec: tektonv1alpha1.TaskRunSpec{
			ServiceAccount: sa.Name,
			Timeout:        service.Spec.Build.BuildTimeout,
			TaskSpec: &tektonv1alpha1.TaskSpec{
				Inputs: &tektonv1alpha1.Inputs{
					Params: []tektonv1alpha1.TaskParam{
						{
							Name:        "image",
							Description: "Where to publish the resulting image",
						},
						{
							Name:        "dockerfile",
							Description: "The name of the Dockerfile",
						},
						{
							Name:        "dockerfile-path",
							Description: "The path of the Dockerfile",
						},
						{
							Name:        "docker-context",
							Description: "The context of the build",
						},
						{
							Name:        "push",
							Description: "Whether push or not",
							Default:     "true",
						},
						{
							Name:        "workingdir",
							Description: " The directory containing the app",
							Default:     "/workspace",
						},
						{
							Name:        "buildkit-image",
							Description: "The name of the BuildKit client (buildctl) image",
							Default:     "moby/buildkit:v0.6.1",
						},
						{
							Name:        "insecure-registry",
							Description: "Whether to use insecure registry",
						},
						{
							Name:        "buildkit-daemon-address",
							Description: "The address of the BuildKit daemon (buildkitd) service",
							Default:     fmt.Sprintf("tcp://%s.%s:8080", constants.BuildkitdService, systemNamespace),
						},
					},
					Resources: []tektonv1alpha1.TaskResource{
						{
							Name: "source",
							Type: tektonv1alpha1.PipelineResourceTypeGit,
						},
					},
				},
				Steps: []corev1.Container{
					{
						Name:       "build-and-push",
						Image:      "${inputs.params.buildkit-image}",
						Command:    []string{"buildctl"},
						WorkingDir: "/workspace/source",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &[]bool{true}[0],
						},
						Args: []string{
							"--addr=${inputs.params.buildkit-daemon-address}",
							"build",
							"--progress=plain",
							"--frontend=dockerfile.v0",
							"--frontend-opt", "filename=${inputs.params.dockerfile}",
							"--local", "context=${inputs.params.docker-context}",
							"--local", "dockerfile=${inputs.params.dockerfile-path}",
							"--output", "type=image,name=${inputs.params.image},push=true,registry.insecure=${inputs.params.insecure-registry}",
						},
					},
				},
			},
			Inputs: tektonv1alpha1.TaskRunInputs{
				Params: []tektonv1alpha1.Param{
					{
						Name:  "image",
						Value: ImageName(rev, service),
					},
					{
						Name:  "insecure-registry",
						Value: strconv.FormatBool(service.Spec.Build.PushRegistry == ""),
					},
					{
						Name:  "dockerfile",
						Value: service.Spec.Build.DockerFile,
					},
					{
						Name:  "docker-context",
						Value: service.Spec.Build.BuildContext,
					},
					{
						Name:  "dockerfile-path",
						Value: service.Spec.Build.DockerFilePath,
					},
				},
				Resources: []tektonv1alpha1.TaskResourceBinding{
					{
						Name: "source",
						ResourceSpec: &tektonv1alpha1.PipelineResourceSpec{
							Type: tektonv1alpha1.PipelineResourceTypeGit,
							Params: []tektonv1alpha1.Param{
								{
									Name:  "url",
									Value: service.Spec.Build.Repo,
								},
								{
									Name:  "revision",
									Value: rev,
								},
							},
						},
					},
				},
			},
		},
	})
	os.Add(sa)
	os.Add(build)
	return nil
}

func populateWebhookAndSecrets(webhookService *riov1.App, service *riov1.Service, os *objectset.ObjectSet) {
	if service.Spec.Build.Revision != "" {
		return
	}
	webhookReceiver := webhookv1.NewGitWatcher(service.Namespace, service.Name, webhookv1.GitWatcher{
		Spec: webhookv1.GitWatcherSpec{
			RepositoryURL:                  service.Spec.Build.Repo,
			Enabled:                        true,
			Push:                           true,
			Tag:                            true,
			PR:                             service.Spec.Build.EnablePR,
			Branch:                         service.Spec.Build.Branch,
			RepositoryCredentialSecretName: service.Spec.Build.GitSecretName,
			GithubWebhookToken:             service.Spec.Build.GithubSecretName,
			GithubDeployment:               true,
		},
	})

	if webhookService != nil && len(webhookService.Status.Endpoints) > 0 {
		webhookReceiver.Spec.ReceiverURL = webhookService.Status.Endpoints[0]
	}

	os.Add(webhookReceiver)
}

func (p populator) setDefaults(service *riov1.Service) {
	if service.Spec.Build.DockerFile == "" {
		service.Spec.Build.DockerFile = "Dockerfile"
	}
	if service.Spec.Build.Template == "" {
		service.Spec.Build.Template = "buildkit"
	}
	if service.Spec.Build.BuildContext == "" {
		service.Spec.Build.BuildContext = "."
	}
	if service.Spec.Build.DockerFilePath == "" {
		service.Spec.Build.DockerFilePath = service.Spec.Build.BuildContext
	}
	if service.Spec.Build.GitSecretName == "" {
		if _, err := p.secretsCache.Get(service.Namespace, DefaultGitCrendential); err == nil {
			service.Spec.Build.GitSecretName = DefaultGitCrendential
		}
	}

	if service.Spec.Build.PushRegistry != "" && service.Spec.Build.PushRegistrySecretName == "" {
		if _, err := p.secretsCache.Get(service.Namespace, DefaultDockerCrendential); err == nil {
			service.Spec.Build.PushRegistrySecretName = DefaultDockerCrendential
			service.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
				{
					Name: DefaultDockerCrendential + "-" + "pull",
				},
			}
		}
	}
}

func ImageName(rev string, service *riov1.Service) string {
	registry := constants.RegistryService
	if service.Spec.Build.PushRegistry != "" {
		registry = service.Spec.Build.PushRegistry
	}
	imageName := service.Namespace + "-" + service.Name
	if service.Spec.Build.BuildImageName != "" {
		imageName = service.Spec.Build.BuildImageName
	}

	return fmt.Sprintf("%s/%s:%s", registry, imageName, rev)
}

func PullImageName(rev string, service *riov1.Service) string {
	registry := "localhost:5442"
	if service.Spec.Build.PushRegistry != "" {
		registry = service.Spec.Build.PushRegistry
	}
	imageName := service.Namespace + "-" + service.Name
	if service.Spec.Build.BuildImageName != "" {
		imageName = service.Spec.Build.BuildImageName
	}

	return fmt.Sprintf("%s/%s:%s", registry, imageName, rev)
}
