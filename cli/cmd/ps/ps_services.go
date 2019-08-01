package ps

import (
	"fmt"

	"github.com/rancher/rio/cli/cmd/revision"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
)

type ServiceData struct {
	ID       string
	Created  string
	Service  *riov1.Service
	Endpoint string
	External string
}

func (p *Ps) apps(ctx *clicontext.CLIContext) error {
	var appDataList []runtime.Object
	appDatas := map[string]tables.AppData{}

	appObjs, err := ctx.List(types.AppType)
	if err != nil {
		return err
	}

	namespaces := sets.NewString()
	for _, app := range appObjs {
		namespaces.Insert(app.(*riov1.App).Namespace)
	}
	m, err := revision.PodsMap(ctx, namespaces.List())
	if err != nil {
		return err
	}
	for _, v := range appObjs {
		app := v.(*riov1.App)
		appDatas[app.Namespace+"/"+app.Name] = tables.AppData{
			ObjectMeta: app.ObjectMeta,
			App:        app,
			Revisions: map[string]struct {
				Revision *riov1.Service
				Pods     []corev1.Pod
			}{},
		}
	}

	svcObjs, err := ctx.List(types.ServiceType)
	if err != nil {
		return err
	}

	for _, v := range svcObjs {
		svc := v.(*riov1.Service)
		appName, version := services.AppAndVersion(svc)
		key := svc.Namespace + "/" + appName
		app, ok := appDatas[key]
		if !ok {
			app = tables.AppData{
				App: riov1.NewApp(svc.Namespace, appName, riov1.App{
					ObjectMeta: v1.ObjectMeta{
						CreationTimestamp: svc.CreationTimestamp,
					},
					Spec: riov1.AppSpec{
						Revisions: []riov1.Revision{
							{
								Scale:   *svc.Spec.Scale,
								Version: version,
							},
						},
					},
				}),
				Revisions: map[string]struct {
					Revision *riov1.Service
					Pods     []corev1.Pod
				}{},
			}
			appDatas[key] = app
		}
		app.Revisions[version] = struct {
			Revision *riov1.Service
			Pods     []corev1.Pod
		}{Revision: svc, Pods: m[fmt.Sprintf("%s/%s/%s", svc.Namespace, appName, version)]}
	}

	for _, v := range appDatas {
		copy := v
		copy.Namespace = v.App.Namespace
		copy.Name = v.App.Name
		appDataList = append(appDataList, &copy)
	}

	writer := tables.NewApp(ctx)
	return writer.Write(appDataList)
}
