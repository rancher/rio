package volume

import (
	"fmt"
	"strconv"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

type Create struct {
	L_Label    map[string]string `desc:"Set meta data on a container"`
	D_Driver   string            `desc:"Volume driver to use"`
	T_Template bool              `desc:"Create volume template, not volume"`
	AccessMode string            `desc:"Volume access mode (readWriteOnce|readWriteMany|readOnlyMany)" default:"readWriteOnce"`
}

func (c *Create) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 2 {
		return fmt.Errorf("two arguments are required, name and size in gigabytes")
	}

	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}

	name := ctx.CLI.Args()[0]
	size := ctx.CLI.Args()[1]
	sizeGB, err := strconv.Atoi(size)
	if err != nil {
		return fmt.Errorf("invalid number: %s", size)
	}

	volume := &client.Volume{
		SizeInGB:   int64(sizeGB),
		Driver:     c.D_Driver,
		Labels:     c.L_Label,
		Template:   c.T_Template,
		AccessMode: c.AccessMode,
	}

	volume.SpaceID, volume.StackID, volume.Name, err = ctx.ResolveSpaceStackName(name)
	if err != nil {
		return err
	}

	volume, err = wc.Volume.Create(volume)
	if err != nil {
		return err
	}

	fmt.Println(volume.ID)
	return nil
}
