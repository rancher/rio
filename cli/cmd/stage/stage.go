package stage

import (
	"fmt"
	"sort"

	"github.com/aokoli/goutils"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Stage struct {
	Image string `desc:"Runtime image (Docker image/OCI image)"`
}

func (r *Stage) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("must specify the service to update")
	}

	if len(ctx.CLI.Args()) > 1 {
		return fmt.Errorf("more than one argument found")
	}

	var err error
	appName, version := kv.Split(ctx.CLI.Args()[0], ":")
	if version == "" {
		version, err = goutils.RandomNumeric(5)
		if err != nil {
			return fmt.Errorf("failed to generate random version, err: %v", err)
		}
	}
	namespace, name := stack.NamespaceAndName(ctx, appName)
	app, err := ctx.Rio.Apps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	sort.Slice(app.Spec.Revisions, func(i, j int) bool {
		return app.Spec.Revisions[i].Weight > app.Spec.Revisions[j].Weight
	})
	if len(app.Spec.Revisions) > 0 {
		rev := app.Spec.Revisions[0]
		svc, err := ctx.Rio.Services(app.Namespace).Get(rev.ServiceName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		spec := svc.Spec.DeepCopy()
		spec.Version = version
		spec.App = app.Name
		spec.Weight = 0
		stagedService := riov1.NewService(app.Namespace, app.Name+"-"+version, riov1.Service{})

		if ctx.CLI.String("image") != "" {
			spec.Image = ctx.CLI.String("image")
		}

		stagedService.Spec = *spec
		if err := ctx.Create(stagedService); err != nil {
			return err
		}
		fmt.Printf("%s/%s:%s\n", stagedService.Namespace, app.Name, version)
	}

	return nil
}
