package webhook

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/github"
	"github.com/google/uuid"
	webhookv1 "github.com/rancher/rio/pkg/apis/webhookinator.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	v12 "github.com/rancher/rio/pkg/generated/controllers/webhookinator.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"golang.org/x/oauth2"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	githubUrl           = "https://api.github.com"
	HooksEndpointPrefix = "hooks?gitwebhookId="
	GitWebHookParam     = "gitwebhookId"
)

func Register(ctx context.Context, rContext *types.Context) error {
	wh := webhookHandler{
		secretsLister: rContext.Core.Core().V1().Secret().Cache(),
	}

	rContext.Webhook.Webhookinator().V1().GitWebHookReceiver().OnChange(ctx, "webhook-receiver",
		v12.UpdateGitWebHookReceiverOnChange(rContext.Webhook.Webhookinator().V1().GitWebHookReceiver().Updater(), wh.onChange))

	return nil
}

type webhookHandler struct {
	secretsLister v1.SecretCache
}

func (w webhookHandler) onChange(key string, obj *webhookv1.GitWebHookReceiver) (*webhookv1.GitWebHookReceiver, error) {
	if obj == nil {
		return nil, nil
	}

	if obj.Status.HookID != "" {
		return obj, nil
	}

	return obj, webhookv1.GitWebHookReceiverConditionRegistered.Do(func() (runtime.Object, error) {
		obj.Status.Token = uuid.New().String()
		secret, err := w.secretsLister.Get(obj.Name, obj.Spec.RepositoryCredentialSecretName)
		if err != nil {
			return obj, err
		}

		token := base64.StdEncoding.EncodeToString(secret.Data["accessToken"])
		client, err := newGithubClient(token)
		if err != nil {
			return obj, err
		}

		repoName, err := getRepoNameFromURL(obj.Spec.RepositoryURL)
		if err != nil {
			return obj, err
		}

		in := &scm.HookInput{
			Name:   "rio-webhookinator",
			Target: getHookEndpoint(obj),
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
	})
}

func newGithubClient(token string) (*scm.Client, error) {
	c, err := github.New(githubUrl)
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

func getHookEndpoint(receiver *webhookv1.GitWebHookReceiver) string {
	if os.Getenv("RIO_WEBHOOK_URL") != "" {
		return hookUrl(os.Getenv("RIO_WEBHOOK_URL"), receiver)
	}
	return hookUrl(fmt.Sprintf("http://%s.%s", namespace.HashIfNeed("webhook", "webhook", "rio-system"), settings.ClusterDomain), receiver)
}

func hookUrl(base string, receiver *webhookv1.GitWebHookReceiver) string {
	return fmt.Sprintf("%s/%s%s:%s", base, HooksEndpointPrefix, receiver.Namespace, receiver.Name)
}
