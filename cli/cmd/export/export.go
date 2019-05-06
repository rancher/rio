package export

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/riofile"
	corev1 "k8s.io/api/core/v1"
)

var (
	exportTypes = []string{
		clitypes.ConfigType,
		clitypes.ServiceType,
	}
)

type Export struct {
	T_Type   string `desc:"Export specific type"`
	O_Output string `desc:"Output format (yaml/json)"`
}

func (e *Export) Run(ctx *clicontext.CLIContext) error {
	services := make(map[string]riov1.Service)
	configs := make(map[string]corev1.ConfigMap)
	for _, arg := range ctx.CLI.Args() {
		r, err := lookup.Lookup(ctx, arg, clitypes.ServiceType, clitypes.ConfigType)
		if err != nil {
			return err
		}

		r, err = ctx.ByID(r.Namespace, r.Name, clitypes.ServiceType)
		if err != nil {
			return err
		}

		obj := r.Object
		switch obj.(type) {
		case *riov1.Service:
			newSvc := riov1.Service{}
			newSvc.Spec = obj.(*riov1.Service).Spec
			services[r.Name] = newSvc
		case *corev1.ConfigMap:
			configs[r.Name] = *obj.(*corev1.ConfigMap)
		}
	}

	content, err := riofile.ParseFrom(services, configs)
	if err != nil {
		return err
	}
	fmt.Println(string(content))
	return nil
}
