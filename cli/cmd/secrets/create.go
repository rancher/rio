package secrets

import (
	"fmt"
	"io/ioutil"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/kv"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	generateversioned "k8s.io/kubectl/pkg/generate/versioned"
)

type Create struct {
	N_Name        string   `desc:"Assign a name to the secret. Use format [namespace:]name"`
	T_Type        string   `desc:"Create type" default:"Opaque"`
	F_FromFile    []string `desc:"Creating secrets from files"`
	D_Data        []string `desc:"Creating secrets from key-pair data"`
	GithubWebhook bool     `desc:"Configure github token"`
	Docker        bool     `desc:"Configure docker registry secret"`
	GitBasicAuth  bool     `desc:"Configure git basic credential"`
	GitSSHKeyAuth bool     `desc:"Configure git ssh key auth"`
}

const (
	defaultDockerRegistry = "https://index.docker.io/v1/"
)

func (s *Create) Run(ctx *clicontext.CLIContext) error {
	name := s.N_Name
	if name != "" {
		r := ctx.ParseID(fmt.Sprintf("secret/%s", name))
		if _, err := ctx.Core.Secrets(r.Namespace).Get(r.Name, metav1.GetOptions{}); err == nil {
			fmt.Printf("Warning: %s already exists\n", r)
		}
	}
	if s.GithubWebhook {
		var err error
		var accessToken, ns string
		ns, err = questions.Prompt("Select namespace[default]: ", "default")
		if err != nil {
			return err
		}

		if name == "" {
			name, err = questions.Prompt(fmt.Sprintf("Name[%s]: ", constants.DefaultGithubCrendential), constants.DefaultGithubCrendential)
			if err != nil {
				return err
			}
		}

		secret, err := ctx.Core.Secrets(ns).Get(name, metav1.GetOptions{})
		if err == nil {
			accessToken = string(secret.Data["accessToken"])
		} else {
			secret = constructors.NewSecret(ns, name, v1.Secret{})
		}
		setDefaults(secret, v1.SecretTypeOpaque)

		accessToken, err = questions.PromptPassword("Github AccessToken[******]: ", accessToken)
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

		if name == "" {
			name, err = questions.Prompt(fmt.Sprintf("Name[%s]: ", constants.DefaultDockerCrendential), constants.DefaultDockerCrendential)
			if err != nil {
				return err
			}
		}

		url, err = questions.Prompt(fmt.Sprintf("Registry URL[%s]: ", defaultDockerRegistry), defaultDockerRegistry)
		if err != nil {
			return err
		}

		username, err = questions.Prompt(fmt.Sprintf("Username[%s]: ", username), username)
		if err != nil {
			return err
		}

		password, err = questions.PromptPassword("Password[******]: ", password)
		if err != nil {
			return err
		}

		generator := generateversioned.SecretForDockerRegistryGeneratorV1{
			Name:     name,
			Username: username,
			Password: password,
			Server:   url,
		}
		pullSecret, err := generator.StructuredGenerate()
		if err != nil {
			return err
		}
		pullSecret.(*v1.Secret).Namespace = ns
		if pullSecret.(*v1.Secret).Annotations == nil {
			pullSecret.(*v1.Secret).Annotations = map[string]string{}
		}
		pullSecret.(*v1.Secret).Annotations["tekton.dev/docker-0"] = url

		return createOrUpdate(pullSecret.(*v1.Secret), ctx)
	}

	if s.GitBasicAuth {
		var err error
		var url, username, password, ns string
		ns, err = questions.Prompt("Select namespace[default]: ", "default")
		if err != nil {
			return err
		}

		if name == "" {
			name, err = questions.Prompt(fmt.Sprintf("Name[%s]: ", constants.DefaultGitCrendential), constants.DefaultGitCrendential)
			if err != nil {
				return err
			}
		}

		secret, err := ctx.Core.Secrets(ns).Get(name, metav1.GetOptions{})
		if err == nil {
			url = secret.Annotations["tekton.dev/git-0"]
			username = string(secret.Data["username"])
			password = string(secret.Data["password"])
		} else {
			secret = constructors.NewSecret(ns, name, v1.Secret{})
		}
		setDefaults(secret, v1.SecretTypeBasicAuth)

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

	if s.GitSSHKeyAuth {
		var err error
		var url, sshPrivateKeyPath, ns string

		ns, err = questions.Prompt("Select namespace[default]: ", "default")
		if err != nil {
			return err
		}

		if name == "" {
			name, err = questions.Prompt(fmt.Sprintf("Name[%s]: ", constants.DefaultGitCrendentialSSH), constants.DefaultGitCrendentialSSH)
			if err != nil {
				return err
			}
		}

		secret, err := ctx.Core.Secrets(ns).Get(name, metav1.GetOptions{})
		if err == nil {
			url = secret.Annotations["tekton.dev/git-0"]
		} else {
			secret = constructors.NewSecret(ns, name, v1.Secret{})
			secret.Type = v1.SecretTypeSSHAuth
		}
		setDefaults(secret, v1.SecretTypeSSHAuth)

		url, err = questions.Prompt(fmt.Sprintf("git url[%s]: ", url), url)
		if err != nil {
			return err
		}
		secret.Annotations["tekton.dev/git-0"] = url

		sshPrivateKeyPath, err = questions.Prompt(fmt.Sprintf("ssh_key_path[%s]: ", sshPrivateKeyPath), sshPrivateKeyPath)
		if err != nil {
			return err
		}
		sshKey, err := ioutil.ReadFile(sshPrivateKeyPath)
		if err != nil {
			return err
		}
		secret.Data["ssh-privatekey"] = sshKey

		return createOrUpdate(secret, ctx)
	}

	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("exact one argument is required")
	}

	r := ctx.ParseID(ctx.CLI.Args()[0])
	secret := constructors.NewSecret(r.Namespace, r.Name, v1.Secret{
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

	if _, err := ctx.Core.Secrets(r.Namespace).Create(secret); err != nil {
		return err
	}
	fmt.Printf("%s:%s\n", secret.Namespace, secret.Name)

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
	fmt.Printf("%s:%s/%s\n", secret.Namespace, "secret", secret.Name)
	return nil
}

func setDefaults(secret *v1.Secret, secretType v1.SecretType) {
	secret.Type = secretType
	if secret.Annotations == nil {
		secret.Annotations = make(map[string]string)
	}
	if secret.StringData == nil {
		secret.StringData = make(map[string]string)
	}
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
}
