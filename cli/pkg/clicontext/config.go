package clicontext

import (
	"flag"
	"os"
	"strings"

	"github.com/pkg/errors"
	projectv1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/rio.cattle.io/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubernetes/staging/src/k8s.io/client-go/tools/clientcmd"
)

var ErrNoConfig = errors.New("no config found")

type Config struct {
	Namespace        string
	Kubeconfig       string
	DefaultStackName string
	Debug            bool
	Wait             bool
	WaitTimeout      int
	WaitState        string
	ServerURL        string
	Token            string

	Core    corev1.CoreV1Interface
	Rio     riov1.RioV1Interface
	Project projectv1.ProjectV1Interface
}

func (c *Config) Validate() error {
	if c.Debug {
		flag.Set("v", "9")
		logrus.SetLevel(logrus.DebugLevel)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	c.Kubeconfig = strings.Replace(c.Kubeconfig, "${HOME}", homeDir, -1)
	c.Kubeconfig = strings.Replace(c.Kubeconfig, "$HOME", homeDir, -1)

	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: c.Kubeconfig},
		&clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{Server: ""},
			Context: clientcmdapi.Context{
				Namespace: c.Namespace,
			},
		})

	namespace, _, err := loader.Namespace()
	if err != nil {
		return err
	}

	restConfig, err := loader.ClientConfig()
	if err != nil {
		return err
	}

	project, err := projectv1.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	rio, err := riov1.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	core, err := corev1.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	c.Rio = rio
	c.Project = project
	c.Core = core
	c.Namespace = namespace
	c.DefaultStackName = "default"
	return nil
}

func (c *Config) Domain() (string, error) {
	return "fixme.domain", nil
}
