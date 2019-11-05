package clicontext

import (
	"flag"
	"io"
	"os"

	v3 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/management.cattle.io/v3"

	"github.com/pkg/errors"
	webhookv1 "github.com/rancher/gitwatcher/pkg/generated/clientset/versioned/typed/gitwatcher.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	projectv1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/sirupsen/logrus"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

var ErrNoConfig = errors.New("Can not find rio info resource inside your cluster. Have you installed Rio?(run `rio install --help`)")

type Config struct {
	AllNamespace        bool
	ShowSystemNamespace bool
	SystemNamespace     string
	DefaultNamespace    string
	Kubeconfig          string
	Debug               bool
	Wait                bool
	WaitTimeout         int
	WaitState           string

	Apply      apply.Apply
	RestConfig *rest.Config
	K8s        *kubernetes.Clientset

	Core       corev1.CoreV1Interface
	Apps       appsv1.AppsV1Interface
	Build      tektonv1alpha1.TektonV1alpha1Interface
	Rio        riov1.RioV1Interface
	Project    projectv1.AdminV1Interface
	Mgmt       v3.ManagementV3Interface
	Gitwatcher webhookv1.GitwatcherV1Interface

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

	if c.ShowSystemNamespace {
		c.DefaultNamespace = c.SystemNamespace
	}

	loader := kubeconfig.GetInteractiveClientConfig(c.Kubeconfig)

	defaultNs, _, err := loader.Namespace()
	if err != nil {
		return err
	}

	restConfig, err := loader.ClientConfig()
	if err != nil {
		return err
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

	apps, err := appsv1.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	gitwatcher, err := webhookv1.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	mgmt, err := v3.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	k8s := kubernetes.NewForConfigOrDie(restConfig)

	c.Apply = apply.New(k8s.Discovery(), apply.NewClientFactory(restConfig))
	c.Apps = apps
	c.K8s = k8s
	c.RestConfig = restConfig
	c.Rio = rio
	c.Project = project
	c.Core = core
	c.Build = build
	c.Gitwatcher = gitwatcher
	c.Mgmt = mgmt

	if c.DefaultNamespace == "" {
		c.DefaultNamespace = defaultNs
	}

	if info, err := project.RioInfos().Get("rio", metav1.GetOptions{}); err == nil {
		c.SystemNamespace = info.Status.SystemNamespace
	}

	c.Writer = os.Stdout

	return nil
}

func (c *Config) Domain() (*v1.ClusterDomain, error) {
	clusterDomain, err := c.Project.ClusterDomains().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(clusterDomain.Items) > 0 {
		return &clusterDomain.Items[0], nil
	}
	return nil, nil
}
