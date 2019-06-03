package edit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/rancher/rio/cli/cmd/inspect"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
	"sigs.k8s.io/yaml"
)

const (
	format = "application/yaml"
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

func (edit *Edit) edit(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("exactly one Name (not name) arguement is required for raw edit")
	}

	var types []string
	for _, t := range inspect.InspectTypes {
		if t == clitypes.AppType {
			continue
		}
		types = append(types, t)
	}
	r, err := lookup.Lookup(ctx, ctx.CLI.Args()[0], types...)
	if err != nil {
		return err
	}

	data, err := json.Marshal(r.Object)
	if err != nil {
		return err
	}

	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return err
	}
	modifiedMap := make(map[string]interface{})
	if !edit.Raw {
		newMeta := map[string]interface{}{}
		if meta, ok := dataMap["metadata"].(map[string]interface{}); ok {
			newMeta["labels"] = meta["labels"]
			newMeta["annotations"] = meta["annotations"]
		}

		modifiedMap["metadata"] = newMeta
		modifiedMap["spec"] = dataMap["spec"]
		if dataMap["data"] != nil {
			modifiedMap["data"] = dataMap["data"]
		}
	} else {
		modifiedMap = dataMap
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
		return ctx.UpdateResource(r, func(obj runtime.Object) error {
			if !edit.Raw {
				m := make(map[string]interface{})
				if err := yaml.Unmarshal(content, &m); err != nil {
					return err
				}
				newMeta := dataMap["metadata"].(map[string]interface{})
				if meta, ok := m["metadata"].(map[string]interface{}); ok {
					newMeta["labels"] = meta["labels"]
					newMeta["annotations"] = meta["annotations"]
				}
				dataMap["spec"] = m["spec"]
				if m["data"] != nil {
					dataMap["data"] = m["data"]
				}

				content, err = json.Marshal(dataMap)
				if err != nil {
					return err
				}
			}

			if err := yaml.Unmarshal(content, &obj); err != nil {
				return err
			}
			return nil
		})
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

//
//func (edit *Edit) update(ctx *clicontext.CLIContext, format string, obj *types.Resource, self string, content []byte) error {
//	if obj.Type == clitypes.NamespaceType {
//		return up.Run(ctx, content, obj.Name, true, edit.Prompt, nil, "")
//	}
//
//	if obj.Type == clitypes.ConfigType {
//		return config.RunUpdate(ctx, obj.Name, content, nil)
//	}
//
//	parsed, err := url.Parse(self)
//	if err != nil {
//		return err
//	}
//
//	q := parsed.Query()
//	q.Set("_edited", "true")
//	q.Set("_replace", "true")
//	parsed.RawQuery = q.Encode()
//
//	req, err := http.NewRequest(http.MethodPut, parsed.String(), bytes.NewReader(content))
//	if err != nil {
//		return err
//	}
//
//	wc, err := ctx.ProjectClient()
//	if err != nil {
//		return err
//	}
//
//	wc.Ops.SetupRequest(req)
//	req.Header.Set("Content-Type", format)
//
//	resp, err := wc.Ops.Client.Do(req)
//	if err != nil {
//		return err
//	}
//	defer func() {
//		io.Copy(ioutil.Discard, resp.Body)
//		resp.Body.Close()
//	}()
//
//	if resp.StatusCode >= 300 {
//		return clientbase.NewAPIError(resp, parsed.String())
//	}
//
//	return nil
//}
