package edit

import (
	"fmt"

	"encoding/json"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/server"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func (edit *Edit) rawEdit(app *cli.Context, ctx *server.Context) error {
	if edit.T_Type == "" {
		return fmt.Errorf("when using raw edit you must specify a specific type")
	}

	if len(app.Args()) != 1 {
		return fmt.Errorf("exactly one ID (not name) arguement is required for raw edit")
	}

	var (
		c   clientbase.APIBaseClientInterface
		err error
	)

	if _, ok := ctx.Client.Types[edit.T_Type]; ok {
		c = &ctx.Client.APIBaseClient
	} else {
		spaceClient, err := ctx.SpaceClient()
		if err != nil {
			return err
		}
		c = &spaceClient.APIBaseClient
	}

	obj := &types.Resource{}
	jsonObj := map[string]interface{}{}

	err = c.ByID(edit.T_Type, app.Args()[0], obj)
	if err != nil {
		return err
	}

	err = c.ByID(edit.T_Type, app.Args()[0], &jsonObj)
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
