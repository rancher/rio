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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

type Export struct {
	Format  string `desc:"Specify output format, yaml/json. Defaults to yaml" default:"yaml"`
	Riofile bool   `desc:"Export riofile format. Only works for namespace (example: rio export --riofile namespace/default)"`
}

func (e *Export) Run(ctx *clicontext.CLIContext) error {
	output := &strings.Builder{}

	var svcs []*riov1.Service
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
			output.Write([]byte(stack.Spec.Template))
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

func exportObjects(objects []runtime.Object, output *strings.Builder, riofile bool, format string) error {
	if !riofile {
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
		result, err := objToYaml(obj, format)
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

func objToYaml(obj runtime.Object, format string) (string, error) {
	data, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		return "", err
	}

	if format == "json" {
		return string(data), nil
	}

	r, err := yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}

	return string(r), nil
}
