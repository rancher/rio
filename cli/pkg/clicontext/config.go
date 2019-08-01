package clicontext

import (
	"flag"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/rancher/rio/pkg/constants"
	projectv1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/admin.rio.cattle.io/v1"
	autoscalev1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/autoscale.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/sirupsen/logrus"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

var ErrNoConfig = errors.New("Can not find rio info resource inside your cluster. Have you installed Rio?(run `rio install --help`)")

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
	Build     tektonv1alpha1.TektonV1alpha1Interface
	Rio       riov1.RioV1Interface
	Project   projectv1.AdminV1Interface
	Autoscale autoscalev1.AutoscaleV1Interface

	NoPrompt bool
	Writer   io.Writer

	DebugLevel string
}

func (c *Config) Validate() error {
	if c.Debug {
		klog.InitFlags(flag.CommandLine)
		flag.CommandLine.Lookup("v").Value.Set(c.DebugLevel)
		flag.CommandLine.Lookup("alsologtostderr").Value.Set("true")

		logrus.SetLevel(logrus.DebugLevel)
	}

	loader := kubeconfig.GetInteractiveClientConfig(c.Kubeconfig)

	restConfig, err := loader.ClientConfig()
	if err != nil {
		logrus.Error(err)
		return ErrNoConfig
	}
	restConfig.QPS = 500
	restConfig.Burst = 100

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

	build, err := tektonv1alpha1.NewForConfig(restConfig)
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
		logrus.Debug(err)
		return ErrNoConfig
	} else if c.SystemNamespace == "" {
		c.SystemNamespace = info.Status.SystemNamespace
	}

	if c.DefaultNamespace == c.SystemNamespace {
		c.ShowSystem = true
	}
	c.Writer = os.Stdout

	return nil
}

func (c *Config) Domain() (string, error) {
	clusterDomain, err := c.Project.ClusterDomains(c.SystemNamespace).Get(constants.ClusterDomainName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return clusterDomain.Status.ClusterDomain, nil
}
