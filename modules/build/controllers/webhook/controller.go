package webhook

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/github"
	"github.com/google/uuid"
	"github.com/rancher/rio/modules/build/controllers/service"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	webhookv1 "github.com/rancher/rio/pkg/apis/webhookinator.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	gitv1controller "github.com/rancher/rio/pkg/generated/controllers/git.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	webhookcontrollerv1 "github.com/rancher/rio/pkg/generated/controllers/webhookinator.rio.cattle.io/v1"
	webhookv1controller "github.com/rancher/rio/pkg/generated/controllers/webhookinator.rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/relatedresource"
	"github.com/rancher/wrangler/pkg/trigger"
	"golang.org/x/oauth2"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	hookTrigger trigger.Trigger
)

const (
	hookEnqueue         = "hook-enqueue"
	githubURL           = "https://api.github.com"
	HooksEndpointPrefix = "hooks?gitwebhookId="
	GitWebHookParam     = "gitwebhookId"
)

func Register(ctx context.Context, rContext *types.Context) error {
	wh := webhookHandler{
		namespace:       rContext.Namespace,
		secretsLister:   rContext.Core.Core().V1().Secret().Cache(),
		services:        rContext.Rio.Rio().V1().Service(),
		webhookReceiver: rContext.Webhook.Webhookinator().V1().GitWebHookReceiver(),
	}

	hookTrigger = trigger.New(rContext.Rio.Rio().V1().Service())
	hookTrigger.OnTrigger(ctx, hookEnqueue, wh.syncAll)

	relatedresource.Watch(ctx, hookEnqueue,
		wh.resolve,
		rContext.Networking.Networking().V1alpha3().VirtualService(),
		rContext.Networking.Networking().V1alpha3().VirtualService(),
		rContext.Global.Project().V1().ClusterDomain())

	rContext.Webhook.Webhookinator().V1().GitWebHookReceiver().OnChange(ctx, "webhook-receiver",
		webhookv1controller.UpdateGitWebHookReceiverOnChange(rContext.Webhook.Webhookinator().V1().GitWebHookReceiver().Updater(), wh.onChange))

	return nil
}

type webhookHandler struct {
	namespace       string
	gitmodule       gitv1controller.GitModuleClient
	secretsLister   v1.SecretCache
	services        riov1controller.ServiceController
	webhookReceiver webhookcontrollerv1.GitWebHookReceiverController
}

func (w webhookHandler) resolve(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	if namespace == w.namespace && name == "webhook" {
		return []relatedresource.Key{hookTrigger.Key()}, nil
	}
	return nil, nil
}

func (w webhookHandler) syncAll() error {
	receivers, err := w.webhookReceiver.Cache().List(w.namespace, labels.NewSelector())
	if err != nil {
		return err
	}
	for _, r := range receivers {
		w.webhookReceiver.Enqueue(r.Namespace, r.Name)
	}
	return nil
}

func (w webhookHandler) onChange(key string, obj *webhookv1.GitWebHookReceiver) (*webhookv1.GitWebHookReceiver, error) {
	if obj == nil {
		return nil, nil
	}

	if obj.Status.HookID != "" {
		return obj, nil
	}

	svc, err := w.services.Cache().Get(w.namespace, "webhook")
	if err != nil {
		return obj, err
	}

	if len(svc.Status.Endpoints) == 0 {
		return obj, err
	}

	return obj, webhookv1.GitWebHookReceiverConditionRegistered.DoUntilTrue(obj, func() (runtime.Object, error) {
		obj, err := w.createGithubWebhook(svc, obj)
		if err != nil {
			return obj, err
		}
		svc, err := w.services.Cache().Get(obj.Namespace, obj.Name)
		if err != nil {
			return obj, err
		}
		deepcopy := svc.DeepCopy()
		firstCommit, err := service.FirstCommit(deepcopy.Spec.Build.Repo, deepcopy.Spec.Build.Branch)
		if err != nil {
			return obj, err
		}
		deepcopy.Spec.Build.Revision = firstCommit
		if _, err := w.services.Update(deepcopy); err != nil {
			return obj, err
		}
		return obj, nil
	})
}

func (w webhookHandler) createGithubWebhook(svc *riov1.Service, obj *webhookv1.GitWebHookReceiver) (*webhookv1.GitWebHookReceiver, error) {
	obj.Status.Token = uuid.New().String()
	secret, err := w.secretsLister.Get(obj.Namespace, obj.Spec.RepositoryCredentialSecretName)
	if err != nil {
		return obj, err
	}

	client, err := newGithubClient(string(secret.Data["accessToken"]))
	if err != nil {
		return obj, err
	}

	repoName, err := getRepoNameFromURL(obj.Spec.RepositoryURL)
	if err != nil {
		return obj, err
	}

	in := &scm.HookInput{
		Name:   "rio-webhookinator",
		Target: getHookEndpoint(obj, svc.Status.Endpoints[0]),
		Secret: obj.Status.Token,
		Events: scm.HookEvents{
			Push: true,
			Tag:  true,
		},
	}

	hook, _, err := client.Repositories.CreateHook(context.Background(), repoName, in)
	if err != nil {
		return obj, err
	}

	obj.Status.HookID = hook.ID
	return obj, nil
}

func newGithubClient(token string) (*scm.Client, error) {
	c, err := github.New(githubURL)
	if err != nil {
		return nil, err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	c.Client = tc
	return c, nil
}

func getRepoNameFromURL(repoURL string) (string, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", err
	}
	repo := strings.TrimPrefix(u.Path, "/")
	repo = strings.TrimSuffix(repo, ".git")
	return repo, nil
}

func getHookEndpoint(receiver *webhookv1.GitWebHookReceiver, endpoint string) string {
	if os.Getenv("RIO_WEBHOOK_URL") != "" {
		return hookURL(os.Getenv("RIO_WEBHOOK_URL"), receiver)
	}
	return hookURL(endpoint, receiver)
}

func hookURL(base string, receiver *webhookv1.GitWebHookReceiver) string {
	return fmt.Sprintf("%s/%s%s:%s", base, HooksEndpointPrefix, receiver.Namespace, receiver.Name)
}
