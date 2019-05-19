package stage

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/aokoli/goutils"
	"github.com/rancher/rio/cli/cmd/edit"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type Stage struct {
	Image  string `desc:"Runtime image (Docker image/OCI image)"`
	E_Edit bool   `desc:"Edit the config to change the spec in new revision"`
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
		if r.E_Edit {
			rev := app.Spec.Revisions[0]
			r, err := ctx.ByID(app.Namespace, rev.ServiceName, types.ServiceType)
			if err != nil {
				return err
			}

			bytes, err := json.Marshal(r.Object)
			if err != nil {
				return err
			}
			yamlBytes, err := yaml.JSONToYAML(bytes)
			if err != nil {
				return err
			}
			yamlBytes = append(yamlBytes, []byte("/n")...)

			update, err := edit.Loop(nil, yamlBytes, func(content []byte) error {
				var obj *riov1.Service
				if err := json.Unmarshal(content, &obj); err != nil {
					return err
				}
				svc := riov1.NewService(namespace, name, riov1.Service{
					Spec: obj.Spec,
				})
				svc.Name = app.Name + "-" + version
				svc.Spec.Version = version
				svc.Spec.App = app.Name
				svc.Spec.Weight = 0
				return ctx.Create(svc)
			})
			if err != nil {
				return err
			}
			if update {
				fmt.Printf("%s/%s:%s\n", r.Namespace, app.Name, version)
			}
		} else {
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
			return ctx.Create(stagedService)
		}
	}

	return nil
}
