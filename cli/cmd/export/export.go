package export

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
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
	T_Type string `desc:"Export specific type. Supported types: namespace or service"`
	Raw    bool   `desc:"Export the raw API object, not the pretty formatted one"`
	Format string `desc:"Specify output format, yaml/json. Defaults to yaml" default:"yaml"`
	Stack  bool   `desc:"Export riofile format. Only works for namespace"`
}

func (e *Export) Run(ctx *clicontext.CLIContext) error {
	output := &strings.Builder{}
	types := []string{
		clitypes.ServiceType,
		clitypes.NamespaceType,
		clitypes.StackType,
	}
	if e.T_Type != "" {
		types = []string{e.T_Type}
	}

	var svcs []*riov1.Service
	for _, arg := range ctx.CLI.Args() {
		r, err := lookup.Lookup(ctx, arg, types...)
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

			if err := exportObjects(objects, output, e.Raw, e.Stack, e.Format); err != nil {
				return err
			}
		case clitypes.StackType:
			stack := r.Object.(*riov1.Stack)
			if e.Raw {
				if err := exportObjects([]runtime.Object{stack}, output, e.Raw, e.Stack, e.Format); err != nil {
					return err
				}
			} else {
				output.Write([]byte(stack.Spec.Template))
			}
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
		if err := exportObjects(objects, output, e.Raw, e.Stack, e.Format); err != nil {
			return err
		}
	}

	fmt.Println(string(output.String()))
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

func exportObjects(objects []runtime.Object, output *strings.Builder, raw, stack bool, format string) error {
	if !stack {
		return exportObjectsNative(objects, output, raw, format)
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

func exportObjectsNative(objects []runtime.Object, output *strings.Builder, raw bool, format string) error {
	for _, obj := range objects {
		result, err := objToYaml(obj, raw, format)
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

func objToYaml(obj runtime.Object, raw bool, format string) (string, error) {
	data, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		return "", err
	}

	if !raw {
		m := make(map[string]interface{})
		if err := json.Unmarshal(data, &m); err != nil {
			return "", err
		}

		modifiedMap := make(map[string]interface{})
		newMeta := map[string]interface{}{}
		if meta, ok := m["metadata"].(map[string]interface{}); ok {
			if meta["labels"] != nil {
				labels := cleanLabels(meta["labels"].(map[string]interface{}))
				if len(labels) != 0 {
					newMeta["labels"] = labels
				}
			}
			if meta["annotations"] != nil {
				anno := cleanLabels(meta["annotations"].(map[string]interface{}))
				if len(anno) != 0 {
					newMeta["annotations"] = anno
				}
			}
			newMeta["name"] = meta["name"]
			newMeta["namespace"] = meta["namespace"]
		}
		modifiedMap["apiVersion"] = m["apiVersion"]
		modifiedMap["kind"] = m["kind"]
		modifiedMap["metadata"] = newMeta
		if _, ok := m["spec"]; ok {
			modifiedMap["spec"] = m["spec"]
		}
		if m["data"] != nil {
			modifiedMap["data"] = m["data"]
		}

		data, err = json.MarshalIndent(modifiedMap, "", " ")
		if err != nil {
			return "", err
		}
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

func cleanLabels(m map[string]interface{}) map[string]interface{} {
	r := make(map[string]interface{})
	for k, v := range m {
		if !strings.Contains(k, "rio.cattle.io") {
			r[k] = v
		}
	}
	return r
}
