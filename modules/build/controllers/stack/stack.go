package stack

import (
	"context"
	"fmt"

	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/modules/service/controllers/service/populate/rbac"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/rancher/wrangler/pkg/objectset"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "stack-service-build", rContext.Rio.Rio().V1().Stack())
	c.Apply = c.Apply.WithCacheTypes(
		rContext.Build.Tekton().V1alpha1().TaskRun(),
		rContext.Webhook.Gitwatcher().V1().GitWatcher(),
		rContext.Core.Core().V1().ServiceAccount(),
		rContext.Core.Core().V1().Secret(),
		rContext.RBAC.Rbac().V1().Role(),
		rContext.RBAC.Rbac().V1().RoleBinding(),
		rContext.RBAC.Rbac().V1().ClusterRole(),
		rContext.RBAC.Rbac().V1().ClusterRoleBinding(),
	).WithStrictCaching()

	p := populator{
		apply:           c.Apply,
		systemNamespace: rContext.Namespace,
		appCache:        rContext.Rio.Rio().V1().App().Cache(),
		info:            rContext.Global.Admin().V1().RioInfo().Cache(),
	}

	c.Populator = p.populate

	return nil
}

type populator struct {
	apply           apply.Apply
	systemNamespace string
	appCache        riov1controller.AppCache
	info            adminv1controller.RioInfoCache
}

func (p populator) populate(obj runtime.Object, ns *corev1.Namespace, os *objectset.ObjectSet) error {
	stack := obj.(*riov1.Stack)

	if stack == nil || stack.Spec.Build == nil || stack.Spec.Build.Repo == "" {
		return nil
	}

	if err := p.populateBuild(stack, p.systemNamespace, os); err != nil {
		return err
	}

	webhook, err := p.appCache.Get(p.systemNamespace, "webhook")
	if errors.IsNotFound(err) {
		webhook = nil
	} else if err != nil {
		return err
	}

	populateWebhookAndSecrets(webhook, stack, os)
	return nil
}

func (p populator) populateBuild(stack *riov1.Stack, systemNamespace string, os *objectset.ObjectSet) error {
	rev := stack.Spec.Build.Revision
	if rev == "" {
		rev = stack.Status.Revision
	}
	if rev == "" {
		return nil
	}

	trName := name.SafeConcatName(stack.Namespace, stack.Name+"-stack", name.Hex(stack.Spec.Build.Repo, 5), name.Hex(rev, 5))
	sa := constructors.NewServiceAccount(stack.Namespace, trName+"-stack", corev1.ServiceAccount{})
	if stack.Spec.Build.GitSecretName != "" {
		sa.Secrets = append(sa.Secrets, corev1.ObjectReference{
			Name: stack.Spec.Build.GitSecretName,
		})
	}
	os.Add(sa)

	info, err := p.info.Get("rio")
	if err != nil {
		return err
	}
	rbacs := populateRbac(stack, sa.Name, p.systemNamespace, info.Status.BuildkitPodName, info.Status.SocatPodName)
	os.Add(rbacs...)

	build := constructors.NewTaskRun(stack.Namespace, trName, tektonv1alpha1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"stack-name":      stack.Name,
				"stack-namespace": stack.Namespace,
			},
		},
		Spec: tektonv1alpha1.TaskRunSpec{
			ServiceAccount: sa.Name,
			TaskSpec: &tektonv1alpha1.TaskSpec{
				Inputs: &tektonv1alpha1.Inputs{
					Resources: []tektonv1alpha1.TaskResource{
						{
							Name: "source",
							Type: tektonv1alpha1.PipelineResourceTypeGit,
						},
					},
				},
				Steps: []corev1.Container{
					{
						Name:            "rio-up",
						Image:           fmt.Sprintf("%s:%s", constants.ControllerImage, constants.ControllerImageTag),
						ImagePullPolicy: corev1.PullAlways,
						WorkingDir:      "/workspace/source",
						Command: []string{
							"rio",
						},
						Args: []string{
							"-n",
							stack.Namespace,
							"up",
							"--name",
							stack.Name,
						},
						Env: []corev1.EnvVar{
							{
								Name:  "PUSH_LOCAL",
								Value: "TRUE",
							},
						},
					},
				},
			},
			Inputs: tektonv1alpha1.TaskRunInputs{
				Resources: []tektonv1alpha1.TaskResourceBinding{
					{
						Name: "source",
						ResourceSpec: &tektonv1alpha1.PipelineResourceSpec{
							Type: tektonv1alpha1.PipelineResourceTypeGit,
							Params: []tektonv1alpha1.Param{
								{
									Name:  "url",
									Value: stack.Spec.Build.Repo,
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
	os.Add(build)
	return nil
}

func populateRbac(stack *riov1.Stack, saName, systemNamespace, buildKitPodName, socatPodName string) []runtime.Object {
	role1 := rbac.NewRole(systemNamespace, fmt.Sprintf("%s-%s-stack", stack.Namespace, stack.Name), nil)
	role1.Rules = []v1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"pods"},
			Verbs:     []string{"get"},
		},
		{
			APIGroups:     []string{""},
			Resources:     []string{"pods/portforward"},
			ResourceNames: []string{buildKitPodName, socatPodName},
			Verbs:         []string{"create", "get"},
		},
	}

	roleBinding1 := rbac.NewBinding(systemNamespace, fmt.Sprintf("%s-%s-stack", stack.Namespace, stack.Name), nil)
	roleBinding1.RoleRef = v1.RoleRef{
		Kind:     "Role",
		Name:     fmt.Sprintf("%s-%s-stack", stack.Namespace, stack.Name),
		APIGroup: "rbac.authorization.k8s.io",
	}
	roleBinding1.Subjects = []v1.Subject{
		{
			Kind:      "ServiceAccount",
			Namespace: stack.Namespace,
			Name:      saName,
		},
	}

	role2 := rbac.NewRole(stack.Namespace, fmt.Sprintf("%s-stack", stack.Name), nil)
	role2.Rules = []v1.PolicyRule{
		{
			APIGroups:     []string{"rio.cattle.io"},
			Resources:     []string{"stacks"},
			ResourceNames: []string{stack.Name},
			Verbs:         []string{"get", "update"},
		},
	}

	roleBinding2 := rbac.NewBinding(stack.Namespace, fmt.Sprintf("%s-stack", stack.Name), nil)
	roleBinding2.RoleRef = v1.RoleRef{
		Kind:     "Role",
		Name:     fmt.Sprintf("%s-stack", stack.Name),
		APIGroup: "rbac.authorization.k8s.io",
	}
	roleBinding2.Subjects = []v1.Subject{
		{
			Kind:      "ServiceAccount",
			Namespace: stack.Namespace,
			Name:      saName,
		},
	}

	clusterRole := rbac.NewClusterRole(stack.Name+"-stack", nil)
	clusterRole.Rules = []v1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"nodes"},
			Verbs:     []string{"list"},
		},
		{
			APIGroups: []string{"admin.rio.cattle.io"},
			Resources: []string{"rioinfos"},
			Verbs:     []string{"get"},
		},
	}

	clusterRolebinding := rbac.NewClusterBinding(stack.Name+"-stack", nil)
	clusterRolebinding.RoleRef = v1.RoleRef{
		Kind:     "ClusterRole",
		Name:     stack.Name + "-stack",
		APIGroup: "rbac.authorization.k8s.io",
	}
	clusterRolebinding.Subjects = []v1.Subject{
		{
			Kind:      "ServiceAccount",
			Namespace: stack.Namespace,
			Name:      saName,
		},
	}
	return []runtime.Object{
		role1,
		roleBinding1,
		role2,
		roleBinding2,
		clusterRole,
		clusterRolebinding,
	}
}

func populateWebhookAndSecrets(webhookService *riov1.App, stack *riov1.Stack, os *objectset.ObjectSet) {
	webhookReceiver := webhookv1.NewGitWatcher(stack.Namespace, stack.Name+"-stack", webhookv1.GitWatcher{
		Spec: webhookv1.GitWatcherSpec{
			RepositoryURL:                  stack.Spec.Build.Repo,
			Enabled:                        true,
			Push:                           true,
			Tag:                            true,
			Branch:                         stack.Spec.Build.Branch,
			RepositoryCredentialSecretName: stack.Spec.Build.GitSecretName,
		},
	})

	if webhookService != nil && len(webhookService.Status.Endpoints) > 0 {
		webhookReceiver.Spec.ReceiverURL = webhookService.Status.Endpoints[0]
	}

	os.Add(webhookReceiver)
}
