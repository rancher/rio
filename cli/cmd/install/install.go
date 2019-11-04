package install

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/rancher/wrangler/pkg/kv"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/progress"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	config2 "github.com/rancher/rio/pkg/config"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/pkg/version"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Install struct {
	IPAddress       []string `desc:"Manually specify IP addresses to generate rdns domain, supports comma separated values" name:"ip-address"`
	DisableFeatures []string `desc:"Manually specify features to disable, supports comma separated values"`
	EnableDebug     bool     `desc:"Enable debug logging in controller"`
	HTTPProxy       string   `desc:"Set HTTP_PROXY environment variable for control plane"`
	Yaml            bool     `desc:"Only print out k8s yaml manifest"`
	Check           bool     `desc:"Only check status, don't deploy controller"`
}

func (i *Install) Run(ctx *clicontext.CLIContext) error {
	if ctx.K8s == nil {
		return fmt.Errorf("can't contact Kubernetes cluster. Please make sure your cluster is accessible")
	}

	namespace := ctx.SystemNamespace
	controllerStack := stack.NewSystemStack(ctx.Apply, nil, namespace, "rio-controller")

	answers := map[string]string{
		"NAMESPACE":      namespace,
		"DEBUG":          strconv.FormatBool(i.EnableDebug),
		"IMAGE":          fmt.Sprintf("%s:%s", constants.ControllerImage, constants.ControllerImageTag),
		"HTTP_PROXY":     i.HTTPProxy,
		"RUN_CONTROLLER": "true",
	}

	if i.Yaml {
		cm, err := i.getConfigMap(ctx, true)
		if err != nil {
			return err
		}

		yamlOutput, err := controllerStack.Yaml(answers, cm)
		if err != nil {
			return err
		}

		fmt.Println(yamlOutput)
		return nil
	}

	cm, err := i.getConfigMap(ctx, false)
	if err != nil {
		return err
	}

	if err := i.preConfigure(ctx, false); err != nil {
		return err
	}

	if !i.Check {
		if err := i.configureNamespace(ctx, controllerStack); err != nil {
			return err
		}

		fmt.Println("Deploying Rio control plane....")
		if err := controllerStack.Deploy(answers, cm); err != nil {
			return err
		}
	}

	progress := progress.NewWriter()
	for {
		// Checking rio-controller deployment
		if !i.Check {
			dep, err := ctx.K8s.AppsV1().Deployments(namespace).Get("rio-controller", metav1.GetOptions{})
			if err != nil {
				return err
			}
			if !isReady(dep.Status) {
				progress.Display("Waiting for deployment %s/%s to become ready", 2, dep.Namespace, dep.Name)
				continue
			}
		}

		// Checking systemInfo and components
		info, err := ctx.Project.RioInfos().Get("rio", metav1.GetOptions{})
		if err != nil || info.Status.Version == "" {
			progress.Display("Waiting for rio controller to initialize", 2)
			continue
		}

		if ready, notReadyList, err := i.checkDeployment(ctx); err != nil {
			return err
		} else if ready {
			clusterDomain, err := ctx.Domain()
			if err != nil {
				return err
			}
			if clusterDomain == nil {
				fmt.Println("\rWarning: clusterDomain is not generated")
			} else {
				_, err = http.Get(fmt.Sprintf("http://%s:%d", clusterDomain.Name, clusterDomain.Spec.HTTPPort))
				if err != nil {
					fmt.Printf("\rWarning: trying to access clusterDomain(http://%s:%d): %v\n", clusterDomain.Name, clusterDomain.Spec.HTTPPort, err)
				} else {
					fmt.Printf("\rGenerating clusterDomain for this cluster: %s. Verified clusterDomain is reachable.\n", clusterDomain.Name)
				}
			}
		} else {
			progress.Display("Waiting for system components. Not ready deployments: %v", 2, notReadyList.String())
			continue
		}

		fmt.Printf("rio controller version %s (%s) installed into namespace %s\n", version.Version, version.GitCommit, info.Status.SystemNamespace)

		fmt.Println("Controller logs are available from `rio systemlogs`")
		fmt.Println("")
		fmt.Println("Welcome to Rio!")
		fmt.Println("")
		fmt.Println("Run `rio run https://github.com/rancher/rio-demo` as an example")
		break
	}
	return nil
}

func (i *Install) preConfigure(ctx *clicontext.CLIContext, ignoreCluster bool) error {
	if ignoreCluster {
		return nil
	}

	nodes, err := ctx.Core.Nodes().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	var totalMemory int64
	for _, node := range nodes.Items {
		totalMemory += node.Status.Capacity.Memory().Value()
	}
	if totalMemory < 2147000000 {
		if isMinikubeCluster(nodes) {
			fmt.Println("Warning: detecting that your minikube cluster doesn't have at least 3 GB of memory. Please try to increase memory by running `minikube start --memory 4098`")
		} else if isDockerForMac(nodes) {
			fmt.Println("Warning: detecting that your Docker For Mac cluster doesn't have at least 3 GB of memory. Please try to increase memory by following the doc https://docs.docker.com/v17.12/docker-for-mac.")
		} else {
			fmt.Println("Warning: detecting that your cluster doesn't have at least 3 GB of memory in total. Please try to increase memory for your nodes")
		}
	}

	return nil
}

func isMinikubeCluster(nodes *v1.NodeList) bool {
	return len(nodes.Items) == 1 && nodes.Items[0].Name == "minikube"
}

func isDockerForMac(nodes *v1.NodeList) bool {
	return len(nodes.Items) == 1 && nodes.Items[0].Name == "docker-for-desktop"
}

func (i *Install) getConfigMap(ctx *clicontext.CLIContext, ignoreCluster bool) (*v1.ConfigMap, error) {
	var (
		cm = constructors.NewConfigMap(ctx.SystemNamespace, config2.ConfigName, v1.ConfigMap{
			Data: map[string]string{},
		})
		cfg config2.Config
	)

	if !ignoreCluster {
		config, err := ctx.Core.ConfigMaps(ctx.SystemNamespace).Get(config2.ConfigName, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		} else if err == nil {
			cfg, err = config2.FromConfigMap(config)
			if err != nil {
				return nil, err
			}
		}
	}

	for _, f := range i.DisableFeatures {
		if cfg.Features == nil {
			cfg.Features = map[string]config2.FeatureConfig{}
		}
		cfg.Features[f] = config2.FeatureConfig{
			Enabled: new(bool),
		}
	}

	for _, ips := range i.IPAddress {
		for _, ip := range strings.Split(ips, ",") {
			found := false
			for _, addr := range cfg.Gateway.StaticAddresses {
				if addr.IP == ip {
					found = true
					break
				}
			}
			if !found {
				cfg.Gateway.StaticAddresses = append(cfg.Gateway.StaticAddresses, adminv1.Address{
					IP: ip,
				})
			}
		}
	}

	return config2.SetConfig(cm, cfg)
}

func (i *Install) configureNamespace(ctx *clicontext.CLIContext, systemStack *stack.SystemStack) error {
	ns, err := ctx.Core.Namespaces().Get(ctx.GetSystemNamespace(), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		ns, err = ctx.Core.Namespaces().Create(constructors.NewNamespace(ctx.GetSystemNamespace(), v1.Namespace{}))
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	systemStack.WithApply(ctx.Apply.WithOwner(ns).WithSetOwnerReference(true, true).WithDynamicLookup())
	return nil
}

var checkFeatures = map[string][]string{
	"gateway":     {"rio-system/gateway-v2", "rio-system/gateway-proxy-v2", "rio-system/gloo"},
	"build":       {"rio-system/buildkitd", "rio-system/webhook", "tekton-pipelines/tekton-pipelines-webhook", "tekton-pipelines/tekton-pipelines-controller"},
	"letsencrypt": {"rio-system/cert-manager"},
	"autoscaling": {"rio-system/autoscaler"},
	"linkerd":     {"linkerd/linkerd-identity", "linkerd/linkerd-tap", "linkerd/linkerd-sp-validator", "linkerd/linkerd-proxy-injector", "linkerd/linkerd-controller", "linkerd/linkerd-grafana", "linkerd/linkerd-web", "linkerd/linkerd-destination", "linkerd/linkerd-prometheus"},
}

type list struct {
	notReady []string
}

func (l list) String() string {
	sort.Strings(l.notReady)
	if len(l.notReady) > 3 {
		return fmt.Sprint(append(l.notReady[:3], "..."))
	}
	return fmt.Sprint(l.notReady)
}

func (i *Install) checkDeployment(ctx *clicontext.CLIContext) (bool, list, error) {
	notReadyList := list{}
	config, err := ctx.Core.ConfigMaps(ctx.SystemNamespace).Get(config2.ConfigName, metav1.GetOptions{})
	if err != nil {
		return false, notReadyList, err
	}

	conf, err := config2.FromConfigMap(config)
	if err != nil {
		return false, notReadyList, err
	}

	for feature, toChecks := range checkFeatures {
		if f, ok := conf.Features[feature]; ok && f.Enabled != nil && !*f.Enabled {
			continue
		}

		for _, name := range toChecks {
			ns, n := kv.Split(name, "/")
			deploy, err := ctx.K8s.AppsV1().Deployments(ns).Get(n, metav1.GetOptions{})
			if err != nil || !isReady(deploy.Status) {
				notReadyList.notReady = append(notReadyList.notReady, name)
			}
		}
	}

	if len(notReadyList.notReady) > 0 {
		return false, notReadyList, nil
	}

	return true, notReadyList, nil
}

func isReady(status appsv1.DeploymentStatus) bool {
	for _, con := range status.Conditions {
		if con.Type == appsv1.DeploymentAvailable && con.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
