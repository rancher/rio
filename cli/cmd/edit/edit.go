package edit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rancher/rio/cli/cmd/inspect"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	"github.com/rancher/wrangler/pkg/gvk"
	name2 "github.com/rancher/wrangler/pkg/name"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
	"sigs.k8s.io/yaml"
)

var (
	editTypes = []string{
		clitypes.NamespaceType,
		clitypes.ServiceType,
		clitypes.ConfigType,
		clitypes.RouterType,
		clitypes.ExternalServiceType,
		clitypes.FeatureType,
	}
)

type Edit struct {
	Raw    bool   `desc:"Edit the raw API object, not the pretty formatted one"`
	T_Type string `desc:"Specific type to edit"`
}

func (edit *Edit) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("at least one parameter is required")
	}

	return edit.edit(ctx)
}

func (edit *Edit) updateObject(ctx *clicontext.CLIContext, data map[string]interface{}, name string, types ...string) (map[string]interface{}, error) {
	_, origObject, err := edit.getObject(ctx, true, name, types...)
	if err != nil {
		return nil, err
	}

	origMdMap := getMdMap(origObject)
	newMdMap := getMdMap(data)

	for _, key := range []string{"labels", "annotations"} {
		origMdMap[key] = newMdMap[key]
	}

	delete(data, "metadata")
	for k, v := range data {
		origObject[k] = v
	}

	return origObject, nil
}

func getMdMap(obj map[string]interface{}) map[string]interface{} {
	md, ok := obj["metadata"]
	if !ok {
		return map[string]interface{}{}
	}

	mdMap, ok := md.(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}

	return mdMap
}

func (edit *Edit) getObject(ctx *clicontext.CLIContext, raw bool, name string, types ...string) (clitypes.Resource, map[string]interface{}, error) {
	r, err := lookup.Lookup(ctx, ctx.CLI.Args()[0], types...)
	if err != nil {
		return clitypes.Resource{}, nil, err
	}

	data, err := json.Marshal(r.Object)
	if err != nil {
		return clitypes.Resource{}, nil, err
	}

	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return clitypes.Resource{}, nil, err
	}

	if raw {
		return r, dataMap, nil
	}

	newMdMap := map[string]interface{}{}
	mdMap := getMdMap(dataMap)
	for _, key := range []string{"labels", "annotations"} {
		if val, ok := mdMap[key]; ok {
			newMdMap[key] = val
		}
	}
	if len(newMdMap) == 0 {
		delete(dataMap, "metadata")
	} else {
		dataMap["metadata"] = newMdMap
	}

	delete(dataMap, "kind")
	delete(dataMap, "apiVersion")
	delete(dataMap, "status")
	return r, dataMap, nil
}

func (edit *Edit) edit(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("exactly one Name (not name) arguement is required for raw edit")
	}

	var (
		name  = ctx.CLI.Args()[0]
		types []string
	)

	for _, t := range inspect.InspectTypes {
		if t == clitypes.AppType {
			continue
		}
		types = append(types, t)
	}

	r, modifiedMap, err := edit.getObject(ctx, edit.Raw, name, types...)
	if err != nil {
		return err
	}

	g, err := gvk.Get(r.Object)
	if err != nil {
		return err
	}
	gvr := schema.GroupVersionResource{
		Group:    g.Group,
		Version:  g.Version,
		Resource: strings.ToLower(name2.GuessPluralName(g.Kind)),
	}

	m, err := json.Marshal(modifiedMap)
	if err != nil {
		return err
	}
	str, err := yaml.JSONToYAML(m)
	if err != nil {
		return err
	}

	updated, err := Loop(nil, str, func(content []byte) error {
		m := make(map[string]interface{})
		if err := yaml.Unmarshal(content, &m); err != nil {
			return err
		}

		if !edit.Raw {
			m, err = edit.updateObject(ctx, m, name, types...)
			if err != nil {
				return err
			}
		}

		obj := &unstructured.Unstructured{
			Object: m,
		}

		c, err := dynamic.NewForConfig(ctx.RestConfig)
		if err != nil {
			return err
		}

		_, err = c.Resource(gvr).Namespace(obj.GetNamespace()).Update(obj, v1.UpdateOptions{})
		return err
	})
	if err != nil {
		return err
	}

	if !updated {
		logrus.Infof("No change for %s/%s", r.Namespace, r.Name)
	}

	return nil
}

type updateFunc func(content []byte) error

func Loop(prefix, input []byte, update updateFunc) (bool, error) {
	for {
		buf := &bytes.Buffer{}
		buf.Write(prefix)
		buf.Write(input)
		rawInput := buf.Bytes()

		editors := []string{
			"KUBE_EDITOR",
			"EDITOR",
		}
		e := editor.NewDefaultEditor(editors)
		content, path, err := e.LaunchTempFile("rio-", "-edit.yaml", buf)
		if path != "" {
			defer os.Remove(path)
		}
		if err != nil {
			return false, err
		}

		if bytes.Compare(content, rawInput) != 0 {
			content = bytes.TrimPrefix(content, prefix)
			input = content
			if err := update(content); err != nil {
				prefix = []byte(fmt.Sprintf("#\n# Error updating content:\n#    %v\n#\n", err.Error()))
				continue
			}
		} else {
			return false, nil
		}

		break
	}

	return true, nil
}
