package promote

import (
	"fmt"

	"github.com/rancher/mapper"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/stack"
	"github.com/rancher/rio/cli/pkg/types"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Promote struct {
	RolloutIncrement int  `desc:"Rollout increment value" default:"5"`
	RolloutInterval  int  `desc:"Rollout interval value" default:"5"`
	NoRollout        bool `desc:"Don't rollout"`
}

func (p *Promote) Run(ctx *clicontext.CLIContext) error {
	var errors []error
	for _, arg := range ctx.CLI.Args() {
		app, version := kv.Split(arg, ":")
		if app == "" || version == "" {
			continue
		}
		namespace, name := stack.NamespaceAndName(ctx, app)
		appObj, err := ctx.Rio.Apps(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		var errors []error
		for _, rev := range appObj.Spec.Revisions {
			resource, err := lookup.Lookup(ctx, fmt.Sprintf("%s/%s", appObj.Namespace, rev.ServiceName), types.ServiceType)
			if err != nil {
				return err
			}
			err = ctx.UpdateResource(resource, func(obj runtime.Object) error {
				service := obj.(*v1.Service)
				if rev.Version == version {
					service.Spec.ServiceRevision.Weight = 100
				} else {
					service.Spec.ServiceRevision.Weight = 0
				}
				if p.NoRollout {
					service.Spec.Rollout = false
				} else {
					service.Spec.Rollout = true
					service.Spec.RolloutInterval = p.RolloutInterval
					service.Spec.RolloutIncrement = p.RolloutIncrement
				}
				return nil
			})
			errors = append(errors, err)
		}
	}

	return mapper.NewErrors(errors...)
}
