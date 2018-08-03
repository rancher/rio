package volume

import (
	"fmt"
	"strconv"

	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type Create struct {
	L_Label    map[string]string `desc:"Set meta data on a container"`
	D_Driver   string            `desc:"Volume driver to use"`
	T_Template bool              `desc:"Create volume template, not volume"`
	AccessMode string            `desc:"Volume access mode (readWriteOnce|readWriteMany|readOnlyMany)" default:"readWriteOnce"`
}

func (c *Create) Run(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	if len(app.Args()) != 2 {
		return fmt.Errorf("two arguments are required, name and size in gigabytes")
	}

	name := app.Args()[0]
	size := app.Args()[1]
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

	volume, err = ctx.Client.Volume.Create(volume)
	if err != nil {
		return err
	}

	fmt.Println(volume.ID)
	return nil
}
