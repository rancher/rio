package service

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	"github.com/rancher/rio/modules/istio/pkg/domains"
	gitv1 "github.com/rancher/rio/pkg/apis/git.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	webhookv1 "github.com/rancher/rio/pkg/apis/webhookinator.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	corev1controller "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/name"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var registrySecretExample = `{
        "auths": {
                "https://index.docker.io/v1/": {
                        "auth": ""
                }
        }
}`

const (
	buildkitAddr       = "buildkit"
	serviceAccountName = "build-sa"
	dockerBuildSecret  = "build-secrets"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "stack-service-build", rContext.Rio.Rio().V1().Service())
	c.Apply = c.Apply.WithCacheTypes(
		rContext.Build.Build().V1alpha1().Build(),
		rContext.Webhook.Webhookinator().V1().GitWebHookReceiver(),
		rContext.Git.Git().V1().GitModule(),
		rContext.Core.Core().V1().ServiceAccount(),
		rContext.Core.Core().V1().Secret(),
	)

	p := populator{
		systemNamespace:    rContext.Namespace,
		secretsCache:       rContext.Core.Core().V1().Secret().Cache(),
		clusterDomainCache: rContext.Global.Project().V1().ClusterDomain().Cache(),
	}

	c.Populator = p.populate
	return nil
}

type populator struct {
	systemNamespace    string
	customRegistry     string
	secretsCache       corev1controller.SecretCache
	clusterDomainCache projectv1controller.ClusterDomainCache
}

func (p populator) populate(obj runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	service := obj.(*riov1.Service)

	if service == nil || service.Spec.Build == nil || service.Spec.Build.Repo == "" {
		return nil
	}

	clusterDomain, err := p.clusterDomainCache.Get(p.systemNamespace, constants.ClusterDomainName)
	if err != nil {
		return err
	}

	addSecrets := true
	if _, err := p.secretsCache.Get(service.Namespace, dockerBuildSecret); err == nil {
		addSecrets = false
	}

	if err := populateBuild(service, p.customRegistry, p.systemNamespace, clusterDomain.Status.ClusterDomain, addSecrets, os); err != nil {
		return err
	}
	populateWebhookAndSecrets(ns.Name, service, os)
	return nil
}

func populateBuild(service *riov1.Service, customRegistry, systemNamespace, domain string, addSecrets bool, os *objectset.ObjectSet) error {
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
	sa := constructors.NewServiceAccount(service.Namespace, serviceAccountName, corev1.ServiceAccount{
		Secrets: []corev1.ObjectReference{
			{
				Name: dockerBuildSecret,
			},
		},
	})
	if addSecrets {
		secrets := constructors.NewSecret(service.Namespace, dockerBuildSecret, corev1.Secret{
			Type: corev1.SecretTypeDockerConfigJson,
			Data: map[string][]byte{
				".dockerconfigjson": []byte(registrySecretExample),
			},
		})
		os.Add(secrets)
	}
	os.Add(sa)
	os.Add(build)
	return nil
}

func populateWebhookAndSecrets(ns string, service *riov1.Service, os *objectset.ObjectSet) {
	if service.Spec.Build.Branch == "" {
		return
	}

	if service.Spec.Build.Secret != "" {
		webhookReceiver := constructors.NewGitWebHookReceiver(service.Namespace, service.Name, webhookv1.GitWebHookReceiver{
			Spec: webhookv1.GitWebHookReceiverSpec{
				RepositoryURL:                  service.Spec.Build.Repo,
				Enabled:                        true,
				Push:                           true,
				Tag:                            true,
				RepositoryCredentialSecretName: service.Spec.Build.Secret,
			},
		})
		os.Add(webhookReceiver)
	} else {
		moduleName := name.SafeConcatName(service.Namespace, service.Name, name.Hex(service.Spec.Build.Repo, 5), service.Spec.Build.Branch)
		module := gitv1.NewGitModule(service.Namespace, moduleName, gitv1.GitModule{
			Spec: gitv1.GitModuleSpec{
				ServiceName:      service.Name,
				ServiceNamespace: service.Namespace,
				Repo:             service.Spec.Build.Repo,
				Branch:           service.Spec.Build.Branch,
			},
		})
		os.Add(module)
	}

	return
}

func ImageName(customeRegistry, registryNamespace, rev, domain string, service *riov1.Service) string {
	var registryAddr string
	if customeRegistry == "" {
		registryAddr = domains.GetExternalDomain("registry", registryNamespace, domain)
		return fmt.Sprintf("%s/%s:%s", registryAddr, service.Namespace+"/"+service.Name, rev)
	}
	return fmt.Sprintf("%s/%s:%s", customeRegistry, service.Namespace+"-"+service.Name, rev)
}

func revision(build *riov1.ImageBuild) (string, error) {
	if build.Revision != "" {
		return build.Revision, nil
	} else if build.Branch != "" {
		return FirstCommit(build.Repo, build.Branch)
	}
	return "", nil
}

// need git installed, also need auth
func FirstCommit(repo, branch string) (string, error) {
	args := []string{"ls-remote", repo, "refs/heads/" + branch}
	buffer := &bytes.Buffer{}
	errBuf := &strings.Builder{}
	cmd := exec.Command("git", args...)
	cmd.Stdout = buffer
	cmd.Stderr = errBuf
	if err := cmd.Run(); err != nil {
		return "", errors.New(errBuf.String())
	}
	scanner := bufio.NewScanner(buffer)
	for scanner.Scan() {
		return strings.Fields(scanner.Text())[0], nil
	}
	return "", errors.New("can't find first commit by git ls-remote")
}
