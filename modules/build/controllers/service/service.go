package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/randomtoken"
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

func Register(ctx context.Context, rContext *types.Context) error {
	p := populator{
		systemNamespace:    rContext.Namespace,
		secretsCache:       rContext.Core.Core().V1().Secret().Cache(),
		clusterDomainCache: rContext.Admin.Admin().V1().ClusterDomain().Cache(),
		serviceCache:       rContext.Rio.Rio().V1().Service().Cache(),
	}

	riov1controller.RegisterServiceGeneratingHandler(
		ctx,
		rContext.Rio.Rio().V1().Service(),
		rContext.Apply.WithCacheTypes(
			rContext.Build.Tekton().V1alpha1().TaskRun(),
			rContext.Webhook.Gitwatcher().V1().GitWatcher(),
			rContext.Core.Core().V1().ServiceAccount(),
			rContext.Core.Core().V1().Secret(),
		),
		"BuildDeployed",
		"service-build",
		p.populate,
		nil)

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

func (p *populator) populate(service *riov1.Service, status riov1.ServiceStatus) ([]runtime.Object, riov1.ServiceStatus, error) {
	imageBuilds := map[string]*riov1.ImageBuildSpec{}
	if service.Spec.ImageBuild != nil {
		imageBuilds[service.Name] = service.Spec.ImageBuild
	}
	for _, con := range service.Spec.Sidecars {
		if con.ImageBuild != nil {
			imageBuilds[con.Name] = con.ImageBuild
		}
	}

	if status.BuildLogToken == "" {
		token, err := randomtoken.Generate()
		if err != nil {
			return nil, status, err
		}
		status.BuildLogToken = token
	}

	os := objectset.NewObjectSet()
	for conName, build := range imageBuilds {
		if build.Repo == "" {
			continue
		}

		if build.Revision == "" {
			status.Watch = true
		}

		if err := p.populateBuild(conName, service.Namespace, build, service, status, p.systemNamespace, os); err != nil {
			return nil, status, err
		}

		if err := p.populateWebhookAndSecrets(build, status, conName, service.Name, service.Namespace, service.Spec.Template, os); err != nil {
			return nil, status, err
		}
	}

	return os.All(), status, nil
}

func (p populator) populateBuild(buildKey, namespace string, build *riov1.ImageBuildSpec, svc *riov1.Service, status riov1.ServiceStatus, systemNamespace string, os *objectset.ObjectSet) error {
	if svc.Spec.Template {
		return nil
	}

	rev := build.Revision
	if rev == "" {
		return nil
	}

	trName := name.SafeConcatName(buildKey, name.Hex(build.Repo, 5), name.Hex(rev, 5))

	p.setDefaults(build, namespace)

	sa := constructors.NewServiceAccount(namespace, trName, corev1.ServiceAccount{})
	if build.CloneSecretName != "" {
		sa.Secrets = append(sa.Secrets, corev1.ObjectReference{
			Name: build.CloneSecretName,
		})
	}
	if build.PushRegistrySecretName != "" {
		sa.Secrets = append(sa.Secrets, corev1.ObjectReference{
			Name: build.PushRegistrySecretName,
		})
	}

	var timeout *metav1.Duration
	if build.TimeoutSeconds != nil {
		timeout = &metav1.Duration{
			Duration: time.Duration(*build.TimeoutSeconds) * time.Second,
		}
	}

	dir, fileName := filepath.Join(build.Context, filepath.Dir(build.Dockerfile)), filepath.Base(build.Dockerfile)
	taskrun := constructors.NewTaskRun(namespace, trName, tektonv1alpha1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				constants.ContainerLabel: buildKey,
				constants.ServiceLabel:   svc.Name,
				constants.GitCommitLabel: svc.Annotations[constants.GitCommitLabel],
				constants.LogTokenLabel:  status.BuildLogToken,
			},
		},
		Spec: tektonv1alpha1.TaskRunSpec{
			ServiceAccountName: sa.Name,
			Timeout:            timeout,
			TaskSpec: &tektonv1alpha1.TaskSpec{
				Inputs: &tektonv1alpha1.Inputs{
					Params: []tektonv1alpha1.ParamSpec{
						{
							Name:        "image",
							Type:        tektonv1alpha1.ParamTypeString,
							Description: "Where to publish the resulting image",
						},
						{
							Name:        "docker-context",
							Type:        tektonv1alpha1.ParamTypeString,
							Description: "The context of the build",
						},
						{
							Name:        "insecure-registry",
							Type:        tektonv1alpha1.ParamTypeString,
							Description: "Whether to use insecure registry",
						},
					},
					Resources: []tektonv1alpha1.TaskResource{
						{
							ResourceDeclaration: tektonv1alpha1.ResourceDeclaration{
								Name: "source",
								Type: tektonv1alpha1.PipelineResourceTypeGit,
							},
						},
					},
				},
				Steps: []tektonv1alpha1.Step{
					{
						Container: corev1.Container{
							Name:       "build-and-push",
							Image:      constants.BuildkitdImage,
							Command:    []string{"buildctl"},
							WorkingDir: "/workspace/source",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &[]bool{true}[0],
							},
							Args: []string{
								fmt.Sprintf("--addr=tcp://%s.%s:8080", constants.BuildkitdService, p.systemNamespace),
								"build",
								"--progress=plain",
								"--frontend=dockerfile.v0",
								"--opt", fmt.Sprintf("filename=%s", fileName),
								"--local", "context=$(inputs.params.docker-context)",
								"--local", fmt.Sprintf("dockerfile=%s", dir),
								"--output", "type=image,name=$(inputs.params.image),push=true,registry.insecure=$(inputs.params.insecure-registry)",
							},
						},
					},
				},
			},
			Inputs: tektonv1alpha1.TaskRunInputs{
				Params: []tektonv1alpha1.Param{
					{
						Name: "image",
						Value: tektonv1alpha1.ArrayOrString{
							Type:      tektonv1alpha1.ParamTypeString,
							StringVal: ImageName(rev, namespace, buildKey, build),
						},
					},
					{
						Name: "insecure-registry",
						Value: tektonv1alpha1.ArrayOrString{
							Type:      tektonv1alpha1.ParamTypeString,
							StringVal: strconv.FormatBool(build.PushRegistry == ""),
						},
					},
					{
						Name: "docker-context",
						Value: tektonv1alpha1.ArrayOrString{
							Type:      tektonv1alpha1.ParamTypeString,
							StringVal: build.Context,
						},
					},
				},
				Resources: []tektonv1alpha1.TaskResourceBinding{
					{
						PipelineResourceBinding: tektonv1alpha1.PipelineResourceBinding{
							Name: "source",
							ResourceSpec: &tektonv1alpha1.PipelineResourceSpec{
								Type: tektonv1alpha1.PipelineResourceTypeGit,
								Params: []tektonv1alpha1.ResourceParam{
									{
										Name:  "url",
										Value: build.Repo,
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
		},
	})
	os.Add(sa)
	os.Add(taskrun)
	return nil
}

func (p populator) populateWebhookAndSecrets(build *riov1.ImageBuildSpec, status riov1.ServiceStatus, containerName, svcName, namespace string, template bool, os *objectset.ObjectSet) error {
	if !status.Watch {
		return nil
	}

	webhook, err := p.serviceCache.Get(p.systemNamespace, "webhook")
	if errors.IsNotFound(err) {
		webhook = nil
	} else if err != nil {
		return err
	}

	webhookReceiver := webhookv1.NewGitWatcher(namespace, fmt.Sprintf("%s-%s", svcName, containerName), webhookv1.GitWatcher{
		Spec: webhookv1.GitWatcherSpec{
			RepositoryURL:                  build.Repo,
			Enabled:                        true,
			Push:                           true,
			Tag:                            true,
			PR:                             build.PR,
			Branch:                         build.Branch,
			RepositoryCredentialSecretName: build.CloneSecretName,
			GithubWebhookToken:             build.WebhookSecretName,
			GithubDeployment:               true,
		},
	})

	webhookReceiver.Annotations = map[string]string{
		constants.ServiceLabel:   svcName,
		constants.ContainerLabel: containerName,
	}

	if webhook != nil && len(webhook.Status.Endpoints) > 0 {
		webhookReceiver.Spec.ReceiverURL = webhook.Status.Endpoints[0]
	}

	os.Add(webhookReceiver)
	return nil
}

func (p populator) setDefaults(build *riov1.ImageBuildSpec, namespace string) {
	if build.Dockerfile == "" {
		build.Dockerfile = "Dockerfile"
	}
	if build.Template == "" {
		build.Template = "buildkit"
	}
	if build.Context == "" {
		build.Context = "."
	}
	if build.CloneSecretName == "" {
		if _, err := p.secretsCache.Get(namespace, constants.DefaultGitCrendential); err == nil {
			build.CloneSecretName = constants.DefaultGitCrendential
		}
	}

	if build.PushRegistry != "" && build.PushRegistrySecretName == "" {
		if _, err := p.secretsCache.Get(namespace, constants.DefaultDockerCrendential); err == nil {
			build.PushRegistrySecretName = constants.DefaultDockerCrendential
		}
	}
}

func ImageName(rev string, namespace, name string, build *riov1.ImageBuildSpec) string {
	registry := constants.RegistryService
	if build.PushRegistry != "" {
		registry = build.PushRegistry
	}
	imageName := namespace + "-" + name
	if build.ImageName != "" {
		imageName = build.ImageName
	}

	suffix := rev
	if len(rev) > 5 {
		suffix = rev[:5]
	}

	return fmt.Sprintf("%s/%s:%s", registry, imageName, suffix)
}

func PullImageName(rev string, namespace, name string, build *riov1.ImageBuildSpec) string {
	registry := constants.LocalRegistry
	if build.PushRegistry != "" {
		registry = build.PushRegistry
	}
	imageName := namespace + "-" + name
	if build.ImageName != "" {
		imageName = build.ImageName
	}

	suffix := rev
	if len(rev) > 5 {
		suffix = rev[:5]
	}

	return fmt.Sprintf("%s/%s:%s", registry, imageName, suffix)
}
