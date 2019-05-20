package edit

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/rancher/rio/cli/pkg/types"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	projectv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (edit *Edit) rawEdit(ctx *clicontext.CLIContext) error {
	if edit.T_Type == "" {
		return fmt.Errorf("when using raw edit you must specify a specific type")
	}

	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("exactly one ID (not name) arguement is required for raw edit")
	}

	r, err := lookup.Lookup(ctx, ctx.CLI.Args()[0], edit.T_Type)
	if err != nil {
		return err
	}

	r, err = ctx.ByID(r.Namespace, r.Name, edit.T_Type)
	if err != nil {
		return err
	}

	data, err := json.Marshal(r.Object)
	if err != nil {
		return err
	}

	str, err := yaml.JSONToYAML(data)
	if err != nil {
		return err
	}

	updated, err := Loop(nil, str, func(content []byte) error {
		return ctx.UpdateResource(r, func(obj runtime.Object) error {
			if err := yaml.Unmarshal(content, &obj); err != nil {
				return err
			}
			return nil
		})
	})
	if err != nil {
		return err
	}

	if !updated {
		logrus.Infof("No change for %s/%s", r.Namespace, r.Name)
	}

	return nil
}

func convertRuntime(t string) runtime.Object {
	switch t {
	case types.AppType:
		return &riov1.App{}
	case types.ServiceType:
		return &riov1.Service{}
	case types.ConfigType:
		return &corev1.ConfigMap{}
	case types.PublicDomainType:
		return &projectv1.PublicDomain{}
	case types.RouterType:
		return &riov1.Router{}
	case types.FeatureType:
		return &projectv1.Feature{}
	case types.ExternalServiceType:
		return &riov1.ExternalService{}
	}
	return &unstructured.Unstructured{}
}
