package install

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

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
	Lite            bool     `desc:"Disable service-mesh, autoscaler and build feature"`
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
		yamlOutput, err := controllerStack.Yaml(answers)
		if err != nil {
			return err
		}
		fmt.Println(yamlOutput)
		return nil
	}

	if err := i.preConfigure(ctx); err != nil {
		return err
	}

	if err := i.configure(ctx, controllerStack); err != nil {
		return err
	}

	if !i.Check {
		fmt.Println("Deploying Rio control plane....")
		if err := controllerStack.Deploy(answers); err != nil {
			return err
		}
	}

	var disabledFeatures []string
	for _, dfs := range i.DisableFeatures {
		parts := strings.Split(dfs, ",")
		for _, p := range parts {
			disabledFeatures = append(disabledFeatures, strings.Trim(p, " "))
		}
	}

	if i.Lite {
		disabledFeatures = append(disabledFeatures, "linkerd", "autoscaling", "build")
	}
	i.DisableFeatures = disabledFeatures

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
		if !info.Status.Ready {
			progress.Display("Waiting for rio controller to be ready", 2)
			continue
		}

		fmt.Printf("\rrio controller version %s (%s) installed into namespace %s\n", version.Version, version.GitCommit, info.Status.SystemNamespace)

		// Checking if clusterDomain is available
		fmt.Println("Detecting if clusterDomain is accessible...")
		clusterDomain, err := ctx.Domain()
		if err != nil {
			return err
		}
		if clusterDomain == nil {
			fmt.Println("Warning: Detected that Rio cluster domain is not generated for this cluster right now")
		} else {
			_, err = http.Get(fmt.Sprintf("http://%s:%d", clusterDomain.Name, clusterDomain.Spec.HTTPPort))
			if err != nil {
				fmt.Printf("Warning: ClusterDomain is not accessible. Error: %v\n", err)
			} else {
				fmt.Println("ClusterDomain is reachable. Run `rio info` to get more info.")
			}
		}

		fmt.Println("Controller logs are available from `rio systemlogs`")
		fmt.Println("")
		fmt.Println("Welcome to Rio!")
		fmt.Println("")
		fmt.Println("Run `rio run https://github.com/rancher/rio-demo` as an example")
		break
	}
	return nil
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

func (i *Install) preConfigure(ctx *clicontext.CLIContext) error {
	var disabledFeatures []string
	for _, dfs := range i.DisableFeatures {
		parts := strings.Split(dfs, ",")
		for _, p := range parts {
			disabledFeatures = append(disabledFeatures, strings.Trim(p, " "))
		}
	}

	if i.Lite {
		disabledFeatures = append(disabledFeatures, "linkerd", "autoscaling", "build")
	}
	i.DisableFeatures = disabledFeatures

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

func (i *Install) configure(ctx *clicontext.CLIContext, systemStack *stack.SystemStack) error {
	ns, err := ctx.Core.Namespaces().Get(ctx.GetSystemNamespace(), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		ns, err = ctx.Core.Namespaces().Create(constructors.NewNamespace(ctx.GetSystemNamespace(), v1.Namespace{}))
		if err != nil {
			return err
		}
	}

	systemStack.WithApply(ctx.Apply.WithOwner(ns).WithSetOwnerReference(true, true).WithDynamicLookup())

	cfg := config2.Config{
		Features:    map[string]config2.FeatureConfig{},
		LetsEncrypt: config2.LetsEncrypt{},
		Gateway:     config2.Gateway{},
	}

	disabled := false
	for _, f := range i.DisableFeatures {
		cfg.Features[f] = config2.FeatureConfig{
			Enabled: &disabled,
		}
	}

	ips := strings.Join(i.IPAddress, ",")
	for _, ip := range strings.Split(ips, ",") {
		if ip != "" {
			cfg.Gateway.StaticAddresses = append(cfg.Gateway.StaticAddresses, adminv1.Address{
				IP: ip,
			})
		}
	}

	if config, err := ctx.Core.ConfigMaps(ctx.SystemNamespace).Get(config2.ConfigName, metav1.GetOptions{}); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		config := constructors.NewConfigMap(ctx.SystemNamespace, config2.ConfigName, v1.ConfigMap{})

		config, err := config2.SetConfig(config, cfg)
		if err != nil {
			return err
		}
		if _, err := ctx.Core.ConfigMaps(ctx.SystemNamespace).Create(config); err != nil {
			return err
		}
	} else {
		config, err := config2.SetConfig(config, cfg)
		if err != nil {
			return err
		}
		if _, err := ctx.Core.ConfigMaps(ctx.SystemNamespace).Update(config); err != nil {
			return err
		}
	}
	return nil
}

func isReady(status appsv1.DeploymentStatus) bool {
	for _, con := range status.Conditions {
		if con.Type == appsv1.DeploymentAvailable && con.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
