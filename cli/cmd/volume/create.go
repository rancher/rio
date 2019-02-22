package volume

import (
	"fmt"
	"strconv"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	name := ctx.CLI.Args()[0]
	size := ctx.CLI.Args()[1]
	sizeGB, err := strconv.Atoi(size)
	if err != nil {
		return fmt.Errorf("invalid number: %s", size)
	}

	volume := &riov1.Volume{
		ObjectMeta: metav1.ObjectMeta{
			Labels: c.L_Label,
		},
		Spec: riov1.VolumeSpec{
			SizeInGB:   sizeGB,
			Driver:     c.D_Driver,
			Template:   c.T_Template,
			AccessMode: c.AccessMode,
		},
	}

	volume.Spec.ProjectName, volume.Spec.StackName, volume.Name, err = stack.ResolveSpaceStackForName(ctx, name)
	if err != nil {
		return err
	}

	volume, err = client.Rio.Volumes(volume.Spec.StackName).Create(volume)
	if err != nil {
		return err
	}

	fmt.Println(volume.Name)
	return nil
}
