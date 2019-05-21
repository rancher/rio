package clicontext

import (
	"os"
	"path/filepath"
	"strings"

	buildv1alpha1 "github.com/knative/build/pkg/client/clientset/versioned/typed/build/v1alpha1"
	"github.com/pkg/errors"
	"github.com/rancher/rio/pkg/constants"
	projectv1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/admin.rio.cattle.io/v1"
	autoscalev1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/autoscale.rio.cattle.io/v1"
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
	ShowSystem       bool
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

	Core      corev1.CoreV1Interface
	Build     buildv1alpha1.BuildV1alpha1Interface
	Autoscale autoscalev1.AutoscaleV1Interface
	Rio       riov1.RioV1Interface
	Project   projectv1.AdminV1Interface
}

func (c *Config) findKubeConfig() {
	homeDir, err := os.UserHomeDir()
	if err == nil {
		c.Kubeconfig = strings.Replace(c.Kubeconfig, "${HOME}", homeDir, -1)
		c.Kubeconfig = strings.Replace(c.Kubeconfig, "$HOME", homeDir, -1)
	}

	if c.Kubeconfig != "" {
		return
	}

	homeConfig := filepath.Join(homeDir, ".kube", "config")
	if _, err := os.Stat(homeConfig); err == nil {
		c.Kubeconfig = homeConfig
		return
	}

	k3sConfig := "/etc/rancher/k3s/k3s.yaml"
	if _, err := os.Stat(k3sConfig); err == nil {
		c.Kubeconfig = k3sConfig
		return
	}

	k3sConfig = filepath.Join(homeDir, ".kube", "k3s.yaml")
	if _, err := os.Stat(k3sConfig); err == nil {
		c.Kubeconfig = k3sConfig
		return
	}
}

func (c *Config) Validate() error {
	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	c.findKubeConfig()

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
		return ErrNoConfig
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

	autoscale, err := autoscalev1.NewForConfig(restConfig)
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
	c.Autoscale = autoscale

	if info, err := project.RioInfos().Get("rio", metav1.GetOptions{}); err != nil {
		return ErrNoConfig
	} else if c.SystemNamespace == "" {
		c.SystemNamespace = info.Status.SystemNamespace
	}

	if c.DefaultNamespace == c.SystemNamespace {
		c.ShowSystem = true
	}

	return nil
}

func (c *Config) Domain() (string, error) {
	clusterDomain, err := c.Project.ClusterDomains(c.SystemNamespace).Get(constants.ClusterDomainName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return clusterDomain.Status.ClusterDomain, nil
}
