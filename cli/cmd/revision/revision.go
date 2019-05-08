package revision

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Revision struct {
	N_Namespace string `desc:"specify namespace"`
	System      bool   `desc:"whether to show system resources"`
}

func (r *Revision) Customize(cmd *cli.Command) {
	cmd.Flags = append(table.WriterFlags(), cmd.Flags...)
}

func (r *Revision) Run(ctx *clicontext.CLIContext) error {
	return r.revisions(ctx)
}

func (r *Revision) revisions(ctx *clicontext.CLIContext) error {
	var output []runtime.Object

	// list services for specific app
	if len(ctx.CLI.Args()) > 0 {
		for _, app := range ctx.CLI.Args() {
			namespace, appName := stack.NamespaceAndName(ctx, app)
			appObj, err := ctx.Rio.Apps(namespace).Get(appName, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return err
			}
			for _, rev := range appObj.Spec.Revisions {
				service, err := ctx.Rio.Services(namespace).Get(rev.ServiceName, metav1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) {
						continue
					}
					return err
				}
				service.Spec.Weight = appObj.Status.RevisionWeight[rev.Version].Weight
				output = append(output, service)
			}
		}
	} else {
		objs, err := ctx.List(types.AppType)
		if err != nil {
			return err
		}
		for _, obj := range objs {
			app := obj.(*riov1.App)
			for _, rev := range app.Spec.Revisions {
				service, err := ctx.Rio.Services(app.Namespace).Get(rev.ServiceName, metav1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) {
						continue
					}
					return err
				}
				service.Spec.Weight = app.Status.RevisionWeight[rev.Version].Weight
				output = append(output, service)
			}
		}
	}

	writer := tables.NewService(ctx)
	return writer.Write(output)
}
