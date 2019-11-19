package stack

import (
	"context"
	"fmt"

	"github.com/rancher/rio/modules/build/pkg"

	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/modules/service/controllers/service/populate/rbac"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/rancher/wrangler/pkg/objectset"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
)

func Register(ctx context.Context, rContext *types.Context) error {
	p := populator{
		systemNamespace: rContext.Namespace,
		info:            rContext.Admin.Admin().V1().RioInfo().Cache(),
		pods:            rContext.Core.Core().V1().Pod().Cache(),
		services:        rContext.Rio.Rio().V1().Service().Cache(),
	}

	riov1controller.RegisterStackGeneratingHandler(ctx,
		rContext.Rio.Rio().V1().Stack(),
		rContext.Apply.WithCacheTypes(
			rContext.Build.Tekton().V1alpha1().TaskRun(),
			rContext.Webhook.Gitwatcher().V1().GitWatcher(),
			rContext.Core.Core().V1().ServiceAccount(),
			rContext.Core.Core().V1().Secret(),
			rContext.RBAC.Rbac().V1().Role(),
			rContext.RBAC.Rbac().V1().RoleBinding(),
		),
		"BuildDeployed",
		"stack-service-build",
		p.populate,
		nil)

	return nil
}

type populator struct {
	apply           apply.Apply
	systemNamespace string
	info            adminv1controller.RioInfoCache
	pods            corev1controller.PodCache
	services        riov1controller.ServiceCache
}

func (p populator) populate(stack *riov1.Stack, status riov1.StackStatus) ([]runtime.Object, riov1.StackStatus, error) {
	if stack == nil || stack.Spec.Build == nil || stack.Spec.Build.Repo == "" {
		return nil, status, nil
	}

	os := objectset.NewObjectSet()

	if err := p.populateBuild(stack, p.systemNamespace, os); err != nil {
		return nil, status, err
	}

	webhook, err := p.services.Get(p.systemNamespace, "webhook")
	if errors.IsNotFound(err) {
		webhook = nil
	} else if err != nil {
		return nil, status, err
	}

	populateWebhookAndSecrets(webhook, stack, os)
	return os.All(), status, nil
}

func (p populator) populateBuild(stack *riov1.Stack, systemNamespace string, os *objectset.ObjectSet) error {
	rev := stack.Spec.Build.Revision
	if rev == "" {
		rev = stack.Status.Revision
	}
	if rev == "" {
		return nil
	}
	rioUpArgs := []string{
		"-n",
		stack.Namespace,
		"up",
		"--name",
		stack.Name,
	}
	if stack.Spec.Build.Riofile != "" {

		rioUpArgs = append(rioUpArgs, "--file", stack.Spec.Build.Riofile)

	}
	if stack.Spec.Build.RiofileAnswers != "" {

		rioUpArgs = append(rioUpArgs, "--answers", stack.Spec.Build.RiofileAnswers)

	}

	trName := name.SafeConcatName(stack.Namespace, stack.Name+"-stack", name.Hex(stack.Spec.Build.Repo, 5), name.Hex(rev, 5))
	sa := constructors.NewServiceAccount(stack.Namespace, trName+"-stack", corev1.ServiceAccount{})
	if stack.Spec.Build.CloneSecretName != "" {
		sa.Secrets = append(sa.Secrets, corev1.ObjectReference{
			Name: stack.Spec.Build.CloneSecretName,
		})
	}
	if stack.Spec.Build.PushRegistrySecretName != "" {
		sa.Secrets = append(sa.Secrets, corev1.ObjectReference{
			Name: stack.Spec.Build.PushRegistrySecretName,
		})
	}
	os.Add(sa)

	r, err := labels.NewRequirement("app", selection.Equals, []string{constants.BuildkitdService})
	if err != nil {
		return err
	}
	selector := labels.NewSelector().Add(*r)
	pods, err := p.pods.List(p.systemNamespace, selector)
	if err != nil {
		return err
	}
	var pod corev1.Pod
	for _, p := range pods {
		if p.Status.Phase == corev1.PodRunning {
			pod = *p
			break
		}
	}

	rbacs := populateRbac(stack, sa.Name, p.systemNamespace, pod.Name)
	os.Add(rbacs...)

	build := constructors.NewTaskRun(stack.Namespace, trName, tektonv1alpha1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"stack-name":      stack.Name,
				"stack-namespace": stack.Namespace,
			},
		},
		Spec: tektonv1alpha1.TaskRunSpec{
			ServiceAccountName: sa.Name,
			TaskSpec: &tektonv1alpha1.TaskSpec{
				Inputs: &tektonv1alpha1.Inputs{
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
							Name:            "rio-up",
							Image:           fmt.Sprintf("%s:%s", constants.ControllerImage, constants.ControllerImageTag),
							ImagePullPolicy: corev1.PullAlways,
							WorkingDir:      "/workspace/source",
							Command: []string{
								"rio",
							},
							Args: rioUpArgs,
						},
					},
				},
			},
			Inputs: tektonv1alpha1.TaskRunInputs{
				Resources: []tektonv1alpha1.TaskResourceBinding{
					{
						PipelineResourceBinding: tektonv1alpha1.PipelineResourceBinding{
							Name: "source",
							ResourceSpec: &tektonv1alpha1.PipelineResourceSpec{
								Type: tektonv1alpha1.PipelineResourceTypeGit,
								Params: []tektonv1alpha1.ResourceParam{
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
		},
	})
	os.Add(build)
	return nil
}

func populateRbac(stack *riov1.Stack, saName, systemNamespace, buildKitPodName string) []runtime.Object {
	role1 := rbac.NewRole(systemNamespace, fmt.Sprintf("%s-%s-stack", stack.Namespace, stack.Name), nil)
	role1.Rules = []v1.PolicyRule{
		{
			APIGroups:     []string{""},
			Resources:     []string{"pods/portforward"},
			ResourceNames: []string{buildKitPodName},
			Verbs:         []string{"create", "get"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"pods"},
			Verbs:     []string{"list", "get"},
		},
	}

	roleBinding1 := rbac.NewBinding(systemNamespace, fmt.Sprintf("%s-%s-stack", stack.Namespace, stack.Name), nil)
	roleBinding1.RoleRef = v1.RoleRef{
		Kind:     "Role",
		Name:     role1.Name,
		APIGroup: "rbac.authorization.k8s.io",
	}
	roleBinding1.Subjects = []v1.Subject{
		{
			Kind:      "ServiceAccount",
			Namespace: stack.Namespace,
			Name:      saName,
		},
	}

	roleBinding2 := rbac.NewBinding(stack.Namespace, fmt.Sprintf("%s-stack", stack.Name), nil)
	roleBinding2.RoleRef = v1.RoleRef{
		Kind:     "ClusterRole",
		Name:     "rio-standard",
		APIGroup: "rbac.authorization.k8s.io",
	}
	roleBinding2.Subjects = []v1.Subject{
		{
			Kind:      "ServiceAccount",
			Namespace: stack.Namespace,
			Name:      saName,
		},
	}

	// extra permission from stack
	roleExtraPermission := rbac.NewRole(stack.Namespace, fmt.Sprintf("%s-stack-extra-permissions", stack.Name), nil)
	for _, perm := range stack.Spec.Permissions {
		if perm.Role != "" {
			continue
		}
		policyRule, ok := rbac.PermToPolicyRule(perm)
		if ok {
			roleExtraPermission.Rules = append(roleExtraPermission.Rules, policyRule)
		}
	}
	roleBinding3 := rbac.NewBinding(stack.Namespace, fmt.Sprintf("%s-stack-extra-permissions", stack.Name), nil)
	roleBinding3.RoleRef = v1.RoleRef{
		Kind:     "Role",
		Name:     roleExtraPermission.Name,
		APIGroup: "rbac.authorization.k8s.io",
	}
	roleBinding3.Subjects = []v1.Subject{
		{
			Kind:      "ServiceAccount",
			Namespace: stack.Namespace,
			Name:      saName,
		},
	}
	objects := []runtime.Object{
		role1,
		roleExtraPermission,
		roleBinding1,
		roleBinding2,
		roleBinding3,
	}

	for _, role := range stack.Spec.Permissions {
		if role.Role == "" {
			continue
		}
		roleBinding := rbac.NewBinding(stack.Namespace, name.SafeConcatName("rio-stack", stack.Name, role.Role), nil)
		roleBinding.Subjects = []v1.Subject{
			{
				Kind:      "ServiceAccount",
				Namespace: stack.Namespace,
				Name:      saName,
			},
		}
		roleBinding.RoleRef = v1.RoleRef{
			Name:     role.Role,
			Kind:     "Role",
			APIGroup: "rbac.authorization.k8s.io",
		}
		objects = append(objects, roleBinding)
	}

	return objects
}

func populateWebhookAndSecrets(webhookService *riov1.Service, stack *riov1.Stack, os *objectset.ObjectSet) {
	webhookReceiver := webhookv1.NewGitWatcher(stack.Namespace, stack.Name+"-stack", webhookv1.GitWatcher{
		Spec: webhookv1.GitWatcherSpec{
			RepositoryURL:                  stack.Spec.Build.Repo,
			Enabled:                        true,
			Push:                           true,
			Tag:                            true,
			Branch:                         stack.Spec.Build.Branch,
			RepositoryCredentialSecretName: stack.Spec.Build.CloneSecretName,
			GithubWebhookToken:             stack.Spec.Build.WebhookSecretName,
		},
	})
	webhookReceiver.Annotations = map[string]string{
		pkg.StackLabel: stack.Name,
	}

	if webhookService != nil && len(webhookService.Status.Endpoints) > 0 {
		webhookReceiver.Spec.ReceiverURL = webhookService.Status.Endpoints[0]
	}

	os.Add(webhookReceiver)
}
