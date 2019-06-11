package ps

import (
	"fmt"
	"strings"

	"github.com/knative/build/pkg/apis/build/v1alpha1"

	services2 "github.com/rancher/rio/pkg/services"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/cli/pkg/types"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ContainerData struct {
	Name      string
	Pod       *v1.Pod
	PodData   *tables.PodData
	Container *v1.Container
}

func ListFirstPod(ctx *clicontext.CLIContext, all bool, podOrServices ...string) (*tables.PodData, error) {
	cds, err := ListPods(ctx, all, podOrServices...)
	if len(cds) == 0 {
		return nil, err
	}
	return &cds[0], err
}

func ListPods(ctx *clicontext.CLIContext, all bool, podOrServices ...string) ([]tables.PodData, error) {
	var result []tables.PodData

	var pods []clitypes.Resource
	var services []clitypes.Resource
	var apps []clitypes.Resource
	var builds []clitypes.Resource

	for _, name := range podOrServices {
		var types []string
		if strings.Contains(name, ":") {
			types = []string{clitypes.ServiceType}
		} else {
			types = []string{clitypes.PodType, clitypes.AppType, clitypes.BuildType}
		}
		r, err := lookup.Lookup(ctx, name, types...)
		if err != nil {
			return nil, err
		}
		switch r.Type {
		case clitypes.PodType:
			pods = append(pods, r)
		case clitypes.ServiceType:
			services = append(services, r)
		case clitypes.AppType:
			apps = append(apps, r)
		case clitypes.BuildType:
			builds = append(builds, r)
		}
	}

	for _, pod := range pods {
		containerName, _ := lookup.ParseContainer(ctx.GetDefaultNamespace(), pod.LookupName)
		pod := pod.Object.(*v1.Pod)
		podData, ok := toPodData(ctx, all, pod, containerName.ContainerName)
		if ok {
			result = append(result, podData)
		}
	}

	if len(pods) > 0 && len(services) == 0 && len(apps) == 0 && len(builds) == 0 {
		return result, nil
	}

	podList, err := ctx.Core.Pods("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for i := range podList.Items {
		podData, ok := toPodData(ctx, all, &podList.Items[i], "")
		if !ok {
			continue
		}

		if len(services) == 0 && len(apps) == 0 && len(builds) == 0 {
			result = append(result, podData)
			continue
		}

		for _, service := range services {
			appName, version := services2.AppAndVersion(service.Object.(*riov1.Service))
			if appName == podData.Service.ServiceName && service.Namespace == podData.Service.StackName && podData.Service.Version == version {
				result = append(result, podData)
				break
			}
		}

		for _, app := range apps {
			if app.Name == podData.Service.ServiceName && app.Namespace == podData.Service.StackName {
				result = append(result, podData)
				break
			}
		}

		for _, build := range builds {
			b := build.Object.(*v1alpha1.Build)
			if podData.Pod.Labels["build.knative.dev/buildName"] == b.Name {
				result = append(result, podData)
				break
			}
		}
	}

	return result, nil
}

func toPodData(ctx *clicontext.CLIContext, all bool, pod *v1.Pod, containerName string) (tables.PodData, bool) {
	stackScoped := lookup.StackScopedFromLabels(ctx.GetDefaultNamespace(), pod)

	podData := tables.PodData{
		Pod:     pod,
		Service: &stackScoped,
		Managed: pod.Namespace == ctx.SystemNamespace,
	}

	lookupName := stackScoped.ServiceName + "-"
	if strings.HasPrefix(pod.Name, lookupName) {
		podData.Name = strings.TrimPrefix(pod.Name, lookupName)
	} else {
		podData.Name = pod.Name
	}

	if !all && podData.Name == "" {
		return podData, false
	}

	if podData.Managed && !ctx.ShowSystem {
		return podData, false
	}

	containers := append(pod.Spec.Containers, pod.Spec.InitContainers...)
	for _, container := range containers {
		if containerName == "" || container.Name == containerName {
			podData.Containers = append(podData.Containers, container)
		}
	}

	return podData, true
}

func Containers(ctx *clicontext.CLIContext) error {
	appDatas, err := listApp(ctx)
	if err != nil {
		return err
	}

	pds, err := ListPods(ctx, true, ctx.CLI.Args()...)
	if err != nil {
		return err
	}

	writer := tables.NewContainer(ctx)
	defer writer.TableWriter().Close()

	for _, pd := range pds {
		if _, ok := appDatas[fmt.Sprintf("%s/%s", pd.Pod.Namespace, pd.Service.ServiceName)]; !ok {
			continue
		}
		for _, container := range pd.Containers {
			writer.TableWriter().Write(ContainerData{
				Name:      fmt.Sprintf("%s/%s", pd.Pod.Namespace, pd.Name),
				PodData:   &pd,
				Container: &container,
			})
		}
	}

	return writer.TableWriter().Err()
}

func Pods(ctx *clicontext.CLIContext) error {
	appDatas, err := listApp(ctx)
	if err != nil {
		return err
	}

	pds, err := ListPods(ctx, true, ctx.CLI.Args()...)
	if err != nil {
		return err
	}

	writer := tables.NewPods(ctx)
	defer writer.TableWriter().Close()

	for _, pd := range pds {
		if _, ok := appDatas[fmt.Sprintf("%s/%s", pd.Pod.Namespace, pd.Service.ServiceName)]; !ok {
			continue
		}
		writer.TableWriter().Write(ContainerData{
			Name:    fmt.Sprintf("%s/%s", pd.Pod.Namespace, pd.Name),
			PodData: &pd,
		})
	}

	return writer.TableWriter().Err()
}

func listApp(ctx *clicontext.CLIContext) (map[string]tables.AppData, error) {
	appDatas := map[string]tables.AppData{}
	appObjs, err := ctx.List(types.AppType)
	if err != nil {
		return appDatas, err
	}
	for _, v := range appObjs {
		app := v.(*riov1.App)
		appDatas[app.Namespace+"/"+app.Name] = tables.AppData{
			App:       app,
			Revisions: map[string]*riov1.Service{},
		}
	}
	return appDatas, nil
}
