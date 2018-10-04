package edit

import (
	"encoding/json"
	"fmt"

	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/sirupsen/logrus"
)

func (edit *Edit) rawEdit(ctx *clicontext.CLIContext) error {
	if edit.T_Type == "" {
		return fmt.Errorf("when using raw edit you must specify a specific type")
	}

	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("exactly one ID (not name) arguement is required for raw edit")
	}

	c, err := ctx.ClientLookup(edit.T_Type)
	if err != nil {
		return err
	}

	obj := &types.Resource{}
	jsonObj := map[string]interface{}{}

	err = c.ByID(edit.T_Type, ctx.CLI.Args()[0], obj)
	if err != nil {
		return err
	}

	err = c.ByID(edit.T_Type, ctx.CLI.Args()[0], &jsonObj)
	if err != nil {
		return err
	}

	str, err := table.FormatJSON(jsonObj)
	if err != nil {
		return err
	}

	updated, err := editLoop(nil, []byte(str), func(content []byte) error {
		updates := map[string]interface{}{}
		err := json.Unmarshal(content, &updates)
		if err != nil {
			return err
		}
		resp := map[string]interface{}{}
		return c.Replace(edit.T_Type, obj, updates, &resp)
	})
	if err != nil {
		return err
	}

	if !updated {
		logrus.Infof("No change for %s", obj.ID)
	}

	return nil
}
