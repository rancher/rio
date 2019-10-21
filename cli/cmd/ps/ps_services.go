package ps

import (
	"sort"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/tables"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

type ServiceData struct {
	ID        string
	Service   *riov1.Service
	Namespace string
}

func (p *Ps) services(ctx *clicontext.CLIContext) error {
	namespace := ctx.GetSetNamespace()
	services, err := ctx.List(clitypes.ServiceType)
	if err != nil {
		return err
	}

	var output []ServiceData

	for _, service := range services {
		allNamespace := namespace == ""
		id, err := util.GetID(service, allNamespace)
		if err != nil {
			return err
		}
		output = append(output, ServiceData{
			ID:        id,
			Service:   service.(*riov1.Service),
			Namespace: service.(*riov1.Service).Namespace,
		})
	}

	sort.Slice(output, func(i, j int) bool {
		return output[i].Service.CreationTimestamp.After(output[j].Service.CreationTimestamp.Time)
	})

	writer := tables.NewService(ctx)
	defer writer.TableWriter().Close()
	for _, obj := range output {
		writer.TableWriter().Write(obj)
	}
	return writer.TableWriter().Err()
}
