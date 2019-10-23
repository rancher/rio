package stage

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/aokoli/goutils"
	"github.com/rancher/rio/cli/cmd/edit/edit"
	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/wrangler/pkg/kv"
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

	serviceName, version := kv.Split(ctx.CLI.Args()[0], ":")
	if version == "" {
		var err error
		version, err = goutils.RandomNumeric(5)
		if err != nil {
			return fmt.Errorf("failed to generate random version, err: %v", err)
		}
		version = "v" + version
	}

	service, err := ctx.ByID(serviceName)
	if err != nil {
		return err
	}

	if r.E_Edit {
		byteContent, err := json.Marshal(service.Object)
		if err != nil {
			return err
		}
		yamlBytes, err := yaml.JSONToYAML(byteContent)
		if err != nil {
			return err
		}
		yamlBytes = append(yamlBytes, []byte("/n")...)

		_, err = edit.Loop(nil, yamlBytes, func(content []byte) error {
			var obj *riov1.Service
			content = bytes.TrimSuffix(bytes.TrimSpace(content), []byte("/n"))
			if err := yaml.Unmarshal(content, &obj); err != nil {
				return err
			}
			svc := riov1.NewService(service.Namespace, service.Name, riov1.Service{
				Spec: obj.Spec,
			})
			app, _ := services.AppAndVersion(svc)
			svc.Name = app + "-" + version
			svc.Spec.Version = version
			svc.Spec.App = obj.Spec.App
			svc.Spec.Weight = &[]int{0}[0]
			return ctx.Create(svc)
		})
		if err != nil {
			return err
		}
	} else {
		svc := service.Object.(*riov1.Service)
		app, _ := services.AppAndVersion(svc)
		spec := svc.Spec.DeepCopy()
		spec.Version = version
		spec.App = app
		spec.Weight = &[]int{0}[0]
		stagedService := riov1.NewService(svc.Namespace, spec.App+"-"+version, riov1.Service{})

		if ctx.CLI.String("image") != "" {
			spec.Image = ctx.CLI.String("image")
		}

		stagedService.Spec = *spec
		return ctx.Create(stagedService)
	}

	return nil
}
