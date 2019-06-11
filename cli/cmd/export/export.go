package export

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

var (
	exportTypes = []string{
		clitypes.ConfigType,
		clitypes.ServiceType,
	}
)

type Export struct {
	T_Type   string `desc:"Export specific type. Supported types: namespace or service"`
	Raw      bool   `desc:"Export the raw API object, not the pretty formatted one"`
	O_Output string `desc:"Output format (yaml/json)"`
}

func (e *Export) Run(ctx *clicontext.CLIContext) error {
	output := &strings.Builder{}
	for _, arg := range ctx.CLI.Args() {
		r, err := lookup.Lookup(ctx, arg, clitypes.ServiceType, clitypes.NamespaceType)
		if err != nil {
			return err
		}

		switch r.Type {
		case clitypes.ServiceType:
			svc := r.Object.(*riov1.Service)
			if err := exportService(ctx, svc, output, !e.Raw); err != nil {
				return err
			}
		case clitypes.NamespaceType:
			ns := r.Object.(*corev1.Namespace)
			services, err := ctx.Rio.Services(ns.Name).List(metav1.ListOptions{})
			if err != nil {
				return err
			}
			for _, svc := range services.Items {
				if err := exportService(ctx, &svc, output, !e.Raw); err != nil {
					return err
				}
			}
		}
	}

	fmt.Println(string(output.String()))
	return nil
}

func exportService(ctx *clicontext.CLIContext, svc *riov1.Service, output *strings.Builder, pretty bool) error {
	result, err := objToYaml(svc, pretty)
	if err != nil {
		return err
	}
	output.WriteString(result)
	output.WriteString("\n---\n")

	for _, cm := range svc.Spec.Configs {
		configMap, err := ctx.Core.ConfigMaps(svc.Namespace).Get(cm.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		result, err := objToYaml(configMap, pretty)
		if err != nil {
			return err
		}
		output.WriteString(result)
		output.WriteString("\n---\n")
	}
	return nil
}

func objToYaml(obj runtime.Object, pretty bool) (string, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	if pretty {
		m := make(map[string]interface{})
		if err := json.Unmarshal(data, &m); err != nil {
			return "", err
		}

		modifiedMap := make(map[string]interface{})
		newMeta := map[string]interface{}{}
		if meta, ok := m["metadata"].(map[string]interface{}); ok {
			if meta["labels"] != nil {
				newMeta["labels"] = meta["labels"]
			}
			if meta["annotations"] != nil {
				newMeta["annotations"] = meta["annotations"]
			}
			newMeta["name"] = meta["name"]
			newMeta["namespace"] = meta["namespace"]
		}
		modifiedMap["apiVersion"] = m["apiVersion"]
		modifiedMap["kind"] = m["kind"]
		modifiedMap["metadata"] = newMeta
		modifiedMap["spec"] = m["spec"]
		if m["data"] != nil {
			modifiedMap["data"] = m["data"]
		}

		data, err = json.Marshal(modifiedMap)
		if err != nil {
			return "", err
		}
	}

	r, err := yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}

	return string(r), nil
}
