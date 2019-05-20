package edit

import (
	"encoding/json"
	"fmt"

	"github.com/rancher/rio/cli/cmd/inspect"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

func (edit *Edit) rawEdit(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("exactly one ID (not name) arguement is required for raw edit")
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

	str, err := yaml.JSONToYAML(data)
	if err != nil {
		return err
	}

	updated, err := Loop(nil, str, func(content []byte) error {
		return ctx.UpdateResource(r, func(obj runtime.Object) error {
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
