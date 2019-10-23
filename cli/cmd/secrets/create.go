package secrets

import (
	"fmt"
	"io/ioutil"

	"github.com/rancher/rio/pkg/constants"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/kv"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	generateversioned "k8s.io/kubernetes/pkg/kubectl/generate/versioned"
)

type Create struct {
	T_Type        string   `desc:"Create type" default:"Opaque"`
	F_FromFile    []string `desc:"Creating secrets from files"`
	D_Data        []string `desc:"Creating secrets from key-pair data"`
	GithubWebhook bool     `desc:"Configure github token"`
	Docker        bool     `desc:"Configure docker registry secret"`
	GitBasicAuth  bool     `desc:"Configure git basic credential"`
}

func (s *Create) Run(ctx *clicontext.CLIContext) error {
	if s.GithubWebhook {
		var err error
		var accessToken, ns string
		ns, err = questions.Prompt("Select namespace[default]: ", "default")
		if err != nil {
			return err
		}

		secret, err := ctx.Core.Secrets(ns).Get(constants.DefaultGithubCrendential, metav1.GetOptions{})
		if err == nil {
			accessToken = string(secret.Data["accessToken"])
		} else {
			secret = constructors.NewSecret(ns, constants.DefaultGithubCrendential, v1.Secret{})
		}
		setDefaults(secret)
		secret.Type = v1.SecretTypeOpaque

		accessToken, err = questions.PromptPassword("accessToken[******]: ", accessToken)
		if err != nil {
			return err
		}
		secret.StringData["accessToken"] = accessToken
		return createOrUpdate(secret, ctx)
	}

	if s.Docker {
		var err error
		var url, username, password, ns string
		ns, err = questions.Prompt("Select namespace[default]: ", "default")
		if err != nil {
			return err
		}

		secret, err := ctx.Core.Secrets(ns).Get(constants.DefaultDockerCrendential, metav1.GetOptions{})
		if err == nil {
			url = secret.Annotations["tekton.dev/docker-0"]
			username = string(secret.Data["username"])
			password = string(secret.Data["password"])
		} else {
			secret = constructors.NewSecret(ns, constants.DefaultDockerCrendential, v1.Secret{})
		}
		setDefaults(secret)

		url, err = questions.Prompt(fmt.Sprintf("Registry url[%s]: ", url), url)
		if err != nil {
			return err
		}
		secret.Annotations["tekton.dev/docker-0"] = url

		username, err = questions.Prompt(fmt.Sprintf("username[%s]: ", username), username)
		if err != nil {
			return err
		}
		secret.StringData["username"] = username

		password, err = questions.PromptPassword("password[******]: ", password)
		if err != nil {
			return err
		}
		secret.StringData["password"] = password
		if err := createOrUpdate(secret, ctx); err != nil {
			return err
		}

		generator := generateversioned.SecretForDockerRegistryGeneratorV1{
			Name:     constants.DefaultDockerCrendential + "-" + "pull",
			Username: username,
			Password: password,
			Server:   url,
		}
		pullSecret, err := generator.StructuredGenerate()
		if err != nil {
			return err
		}
		pullSecret.(*v1.Secret).Namespace = ns

		return createOrUpdate(pullSecret.(*v1.Secret), ctx)
	}

	if s.GitBasicAuth {
		var err error
		var url, username, password, ns string
		ns, err = questions.Prompt("Select namespace[default]: ", "default")
		if err != nil {
			return err
		}

		secret, err := ctx.Core.Secrets(ns).Get(constants.DefaultGitCrendential, metav1.GetOptions{})
		if err == nil {
			url = secret.Annotations["tekton.dev/git-0"]
			username = string(secret.Data["username"])
			password = string(secret.Data["password"])
		} else {
			secret = constructors.NewSecret(ns, constants.DefaultGitCrendential, v1.Secret{})
		}
		setDefaults(secret)

		url, err = questions.Prompt(fmt.Sprintf("git url[%s]: ", url), url)
		if err != nil {
			return err
		}
		secret.Annotations["tekton.dev/git-0"] = url

		username, err = questions.Prompt(fmt.Sprintf("username[%s]: ", username), username)
		if err != nil {
			return err
		}
		secret.StringData["username"] = username

		password, err = questions.PromptPassword("password[******]: ", password)
		if err != nil {
			return err
		}
		secret.StringData["password"] = password
		return createOrUpdate(secret, ctx)
	}

	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("exact one argument is required")
	}

	namespace, name := stack.NamespaceAndName(ctx, ctx.CLI.Args()[0])
	secret := constructors.NewSecret(namespace, name, v1.Secret{
		Type:       v1.SecretType(s.T_Type),
		Data:       make(map[string][]byte),
		StringData: make(map[string]string),
	})
	for _, f := range s.F_FromFile {
		k, v := kv.Split(f, "=")
		content, err := ioutil.ReadFile(v)
		if err != nil {
			return err
		}
		secret.Data[k] = content
	}

	for _, d := range s.D_Data {
		k, v := kv.Split(d, "=")
		secret.StringData[k] = v
	}
	if _, err := ctx.Core.Secrets(namespace).Create(secret); err != nil {
		return err
	}
	fmt.Printf("%s/%s\n", secret.Namespace, secret.Name)

	return nil
}

func createOrUpdate(secret *v1.Secret, ctx *clicontext.CLIContext) error {
	if _, err := ctx.Core.Secrets(secret.Namespace).Get(secret.Name, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			if _, err := ctx.Core.Secrets(secret.Namespace).Create(secret); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if _, err := ctx.Core.Secrets(secret.Namespace).Update(secret); err != nil {
			return err
		}
	}
	fmt.Printf("%s/%s\n", secret.Namespace, secret.Name)
	return nil
}

func setDefaults(secret *v1.Secret) {
	secret.Type = v1.SecretTypeBasicAuth
	if secret.Annotations == nil {
		secret.Annotations = make(map[string]string)
	}
	if secret.StringData == nil {
		secret.StringData = make(map[string]string)
	}
}
