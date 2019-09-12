package revision

import (
	"fmt"
	"sort"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	"github.com/urfave/cli"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type Revision struct {
	N_Namespace string `desc:"specify namespace"`
	System      bool   `desc:"whether to show system resources"`
}

func (r *Revision) Customize(cmd *cli.Command) {
	cmd.Flags = append(table.WriterFlags(), cmd.Flags...)
}

func (r *Revision) Run(ctx *clicontext.CLIContext) error {
	return Revisions(ctx)
}

type ServiceData struct {
	Name    string
	Service *riov1.Service
	Pods    []v1.Pod
}

func Revisions(ctx *clicontext.CLIContext) error {
	var output []ServiceData

	// list services for specific app
	if len(ctx.CLI.Args()) > 0 {
		namespaces := sets.NewString()
		for _, app := range ctx.CLI.Args() {
			namespace, _ := stack.NamespaceAndName(ctx, app)
			namespaces.Insert(namespace)
		}

		m, err := PodsMap(ctx, namespaces.List())
		if err != nil {
			return err
		}
		for _, app := range ctx.CLI.Args() {
			namespace, appName := stack.NamespaceAndName(ctx, app)
			appObj, err := ctx.Rio.Apps(namespace).Get(appName, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return err
			}
			for _, rev := range appObj.Spec.Revisions {
				service, err := ctx.Rio.Services(namespace).Get(rev.ServiceName, metav1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) {
						continue
					}
					return err
				}
				service.Spec.Weight = appObj.Status.RevisionWeight[rev.Version].Weight
				app, version := services.AppAndVersion(service)
				output = append(output, ServiceData{
					Name:    fmt.Sprintf("%s/%s", service.Namespace, app),
					Service: service,
					Pods:    m[fmt.Sprintf("%s/%s/%s", service.Namespace, app, version)],
				})
			}
		}
	} else {
		objs, err := ctx.List(types.AppType)
		if err != nil {
			return err
		}
		namespaces := sets.NewString()
		for _, obj := range objs {
			app := obj.(*riov1.App)
			namespaces.Insert(app.Namespace)
		}
		m, err := PodsMap(ctx, namespaces.List())
		if err != nil {
			return err
		}

		for _, obj := range objs {
			app := obj.(*riov1.App)
			for _, rev := range app.Spec.Revisions {
				service, err := ctx.Rio.Services(app.Namespace).Get(rev.ServiceName, metav1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) {
						continue
					}
					return err
				}
				service.Spec.Weight = app.Status.RevisionWeight[rev.Version].Weight
				appName, version := services.AppAndVersion(service)
				output = append(output, ServiceData{
					Name:    fmt.Sprintf("%s/%s", service.Namespace, app),
					Service: service,
					Pods:    m[fmt.Sprintf("%s/%s/%s", app.Namespace, appName, version)],
				})
			}
		}
	}

	sort.Slice(output, func(i, j int) bool {
		return output[i].Service.CreationTimestamp.After(output[j].Service.CreationTimestamp.Time)
	})

	writer := tables.NewService(ctx)
	defer writer.TableWriter().Close()
	for _, obj := range output {
		writer.TableWriter().Write(obj)
	}
	return writer.TableWriter().Err()
}

func PodsMap(ctx *clicontext.CLIContext, namespaces []string) (map[string][]v1.Pod, error) {
	podMap := map[string][]v1.Pod{}
	for _, ns := range namespaces {
		pods, err := ctx.Core.Pods(ns).List(metav1.ListOptions{})
		if err != nil {
			return podMap, err
		}
		for _, p := range pods.Items {
			app := p.Labels["app"]
			version := p.Labels["version"]
			key := fmt.Sprintf("%s/%s/%s", ns, app, version)
			list := podMap[key]
			list = append(list, p)
			podMap[key] = list
		}
	}
	return podMap, nil
}
