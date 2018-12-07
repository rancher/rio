package volume

import (
	"fmt"
	"strconv"

	"github.com/rancher/rio/cli/pkg/stack"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/types/client/rio/v1"
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

	wc, err := ctx.ProjectClient()
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

	volume.ProjectID, volume.StackID, volume.Name, err = stack.ResolveSpaceStackForName(ctx, name)
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
