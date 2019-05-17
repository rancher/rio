package clicontext

import (
	"os"
	"strings"

	buildv1alpha1 "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	"github.com/pkg/errors"
	"github.com/rancher/rio/pkg/constants"
	projectv1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubernetes/staging/src/k8s.io/client-go/tools/clientcmd"
)

var ErrNoConfig = errors.New("no config found")

type Config struct {
	SystemNamespace  string
	DefaultNamespace string
	Kubeconfig       string
	Debug            bool
	Wait             bool
	WaitTimeout      int
	WaitState        string

	Apply      apply.Apply
	RestConfig *rest.Config
	K8s        *kubernetes.Clientset

	Core    corev1.CoreV1Interface
	Build   buildv1alpha1.BuildV1alpha1Interface
	Rio     riov1.RioV1Interface
	Project projectv1.AdminV1Interface
}

func (c *Config) Validate() error {
	if c.Debug {
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
				Namespace: c.SystemNamespace,
			},
		})

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

	build, err := buildv1alpha1.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	k8s := kubernetes.NewForConfigOrDie(restConfig)

	c.Apply = apply.New(k8s.Discovery(), apply.NewClientFactory(restConfig))
	c.K8s = k8s
	c.RestConfig = restConfig
	c.Rio = rio
	c.Project = project
	c.Core = core
	c.Build = build
	return nil
}

func (c *Config) Domain() (string, error) {
	clusterDomain, err := c.Project.ClusterDomains(c.SystemNamespace).Get(constants.ClusterDomainName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return clusterDomain.Status.ClusterDomain, nil
}
