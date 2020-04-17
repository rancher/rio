package export

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/riofile"
	"github.com/rancher/wrangler/pkg/gvk"
	name "github.com/rancher/wrangler/pkg/name"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Export struct {
	Format  string `desc:"Specify output format, yaml/json. Defaults to yaml" default:"yaml"`
	Riofile bool   `desc:"Export riofile format. (example: rio export --riofile namespace/default)"`
}

func (e *Export) Run(ctx *clicontext.CLIContext) error {
	output := &strings.Builder{}

	var (
		svcs  []*riov1.Service
		other []runtime.Object
	)

	for _, arg := range ctx.CLI.Args() {
		r, err := ctx.ByID(arg)
		if err != nil {
			return err
		}

		switch r.Type {
		case clitypes.ServiceType:
			svc := r.Object.(*riov1.Service)
			svcs = append(svcs, svc)
		case clitypes.NamespaceType:
			ns := r.Object.(*corev1.Namespace)

			objects, err := collectObjectsFromNs(ctx, ns.Name)
			if err != nil {
				return err
			}
			if err := exportObjects(objects, output, e.Riofile, e.Format); err != nil {
				return err
			}
		case clitypes.StackType:
			stack := r.Object.(*riov1.Stack)
			objects, err := collectObjectsFromStack(ctx, stack)
			if err != nil {
				return err
			}
			if err := exportObjects(objects, output, e.Riofile, e.Format); err != nil {
				return err
			}
		default:
			other = append(other, r.Object)
		}
	}

	if len(other) > 0 {
		if err := exportObjects(other, output, e.Riofile, e.Format); err != nil {
			return err
		}
	}

	if len(svcs) > 0 {
		var objects []runtime.Object
		for _, svc := range svcs {
			cms, err := configmapsFromService(ctx, *svc)
			if err != nil {
				return err
			}
			objects = append(objects, cms...)
		}
		for _, svc := range svcs {
			objects = append(objects, svc)
		}
		if err := exportObjects(objects, output, e.Riofile, e.Format); err != nil {
			return err
		}
	}

	fmt.Println(output.String())
	return nil
}

func collectObjectsFromStack(ctx *clicontext.CLIContext, stack *riov1.Stack) ([]runtime.Object, error) {
	var objects []runtime.Object
	client, err := dynamic.NewForConfig(ctx.Config.RestConfig)
	if err != nil {
		return nil, err
	}
	for _, gvk := range stack.Spec.AdditionalGroupVersionKinds {
		gvr := schema.GroupVersionResource{
			Group:    gvk.Group,
			Version:  gvk.Version,
			Resource: strings.ToLower(name.GuessPluralName(gvk.Kind)),
		}
		resourceClient := client.Resource(gvr)
		ul, err := resourceClient.Namespace(stack.Namespace).List(metav1.ListOptions{
			LabelSelector: "rio.cattle.io/stack=" + stack.Name,
		})
		if err != nil {
			return nil, err
		}
		for _, item := range ul.Items {
			if item.GroupVersionKind().Group == "rio.cattle.io" {
				data, err := item.MarshalJSON()
				if err != nil {
					return nil, err
				}
				if item.GetKind() == "Service" {
					svc := riov1.Service{}
					if err := json.Unmarshal(data, &svc); err != nil {
						return nil, err
					}
					objects = append(objects, &svc)
				} else if item.GetKind() == "ExternalService" {
					es := &riov1.ExternalService{}
					if err := json.Unmarshal(data, es); err != nil {
						return nil, err
					}
					objects = append(objects, es)
				} else if item.GetKind() == "ConfigMap" {
					cm := &corev1.ConfigMap{}
					if err := json.Unmarshal(data, cm); err != nil {
						return nil, err
					}
					objects = append(objects, cm)
				} else if item.GetKind() == "Router" {
					rt := &riov1.Router{}
					if err := json.Unmarshal(data, rt); err != nil {
						return nil, err
					}
					objects = append(objects, rt)
				}
			} else {
				// From here down we only want to export non-rio created objects which would already be exported above
				stackObj := true
				for k, v := range item.GetAnnotations() {
					if stackObj && k == "objectset.rio.cattle.io/owner-gvk" && v != "rio.cattle.io/v1, Kind=Stack" {
						stackObj = false
						break
					}
				}
				if stackObj {
					objects = append(objects, item.DeepCopyObject())
				}
			}
		}
	}
	return objects, nil
}

func collectObjectsFromNs(ctx *clicontext.CLIContext, ns string) ([]runtime.Object, error) {
	var objects []runtime.Object

	services, err := ctx.Rio.Services(ns).List(metav1.ListOptions{})
	if err != nil {
		return objects, err
	}
	for i := range services.Items {
		objects = append(objects, &services.Items[i])
	}

	for i := range services.Items {
		configs, err := configmapsFromService(ctx, services.Items[i])
		if err != nil {
			return objects, err
		}
		objects = append(objects, configs...)
	}

	externalServices, err := ctx.Rio.ExternalServices(ns).List(metav1.ListOptions{})
	if err != nil {
		return objects, err
	}
	for i := range externalServices.Items {
		objects = append(objects, &externalServices.Items[i])
	}

	routers, err := ctx.Rio.Routers(ns).List(metav1.ListOptions{})
	if err != nil {
		return objects, err
	}
	for i := range routers.Items {
		objects = append(objects, &routers.Items[i])
	}

	return objects, nil
}

func configmapsFromService(ctx *clicontext.CLIContext, svc riov1.Service) ([]runtime.Object, error) {
	var objects []runtime.Object

	for _, cm := range svc.Spec.Configs {
		configMap, err := ctx.Core.ConfigMaps(svc.Namespace).Get(cm.Name, metav1.GetOptions{})
		if err != nil {
			return objects, err
		}
		if err := gvk.Set(configMap); err != nil {
			return objects, err
		}
		objects = append(objects, configMap)
	}
	return objects, nil
}

func exportObjects(objects []runtime.Object, output *strings.Builder, riofileFormat bool, format string) error {
	if !riofileFormat {
		return exportObjectsNative(objects, output, format)
	}

	return exportObjectStack(objects, output)
}

func exportObjectStack(objects []runtime.Object, output *strings.Builder) error {
	content, err := riofile.Render(objects)
	if err != nil {
		return err
	}
	output.Write(content)
	output.Write([]byte("\n"))
	return nil
}

func exportObjectsNative(objects []runtime.Object, output *strings.Builder, format string) error {
	for _, obj := range objects {
		result, err := riofile.ObjToYaml(obj, format)
		if err != nil {
			return err
		}
		output.WriteString(result)
		output.WriteString("\n")
		if format != "json" {
			output.WriteString("---\n")
		}
	}
	return nil
}
