package ps

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

type ServiceData struct {
	ID       string
	Created  string
	Service  *client.Service
	Stack    *client.Stack
	Endpoint string
	External string
}

func FormatServiceName(cluster *clientcfg.Cluster) func(data, data2 interface{}) (string, error) {
	return func(data, data2 interface{}) (string, error) {
		stackName, ok := data.(string)
		if !ok {
			return "", nil
		}

		service, ok := data2.(*client.Service)
		if !ok {
			return "", nil
		}

		if service.ParentService == "" || service.Version == "" {
			return table.FormatStackScopedName(cluster)(stackName, service.Name)
		}

		return table.FormatStackScopedName(cluster)(stackName, service.ParentService+":"+service.Version)
	}
}

func FormatImage(data interface{}) (string, error) {
	s, ok := data.(*client.Service)
	if !ok {
		return fmt.Sprint(data), nil
	}
	if s.Image == "" || len(s.Sidekicks) > 0 {
		return s.Sidekicks[firstSortedKey(s.Sidekicks)].Image, nil
	}
	return s.Image, nil
}

func firstSortedKey(m map[string]client.SidekickConfig) string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	if len(keys) == 0 {
		return ""
	}
	sort.Strings(keys)
	return keys[0]
}

func FormatScale(data, data2 interface{}) (string, error) {
	scale, ok := data.(int64)
	if !ok {
		return fmt.Sprint(data), nil
	}
	scaleStr := strconv.FormatInt(scale, 10)

	scaleStatus, ok := data2.(*client.ScaleStatus)
	if !ok || scaleStatus == nil {
		return scaleStr, nil
	}

	if scaleStatus.Available == 0 && scaleStatus.Unavailable == 0 && scaleStatus.Ready == scale {
		return scaleStr, nil
	}

	percentage := ""
	if scale > 0 && scaleStatus.Updated > 0 && scale != scaleStatus.Updated {
		percentage = fmt.Sprintf(" %d%%", (scaleStatus.Updated*100)/scale)
	}

	return fmt.Sprintf("(%d/%d/%d)/%d%s", scaleStatus.Unavailable, scaleStatus.Available, scaleStatus.Ready, scale, percentage), nil
}

func (p *Ps) services(ctx *clicontext.CLIContext, stacks map[string]bool) error {
	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}

	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}

	services, err := wc.Service.List(util.DefaultListOpts())
	if err != nil {
		return err
	}

	writer := table.NewWriter([][]string{
		{"NAME", "{{serviceName .Stack.Name .Service}}"},
		{"IMAGE", "{{.Service | image}}"},
		{"CREATED", "{{.Created | ago}}"},
		{"SCALE", "{{scale .Service.Scale .Service.ScaleStatus}}"},
		{"STATE", "Service.State"},
		{"ENDPOINT", "Endpoint"},
		{"DETAIL", "{{first .Service.TransitioningMessage .Stack.TransitioningMessage}}"},
		{"EXTERNAL", "External"},
	}, ctx)
	defer writer.Close()

	writer.AddFormatFunc("serviceName", FormatServiceName(cluster))
	writer.AddFormatFunc("image", FormatImage)
	writer.AddFormatFunc("scale", FormatScale)

	stackByID, err := util.StacksByID(wc)
	if err != nil {
		return err
	}

	for i, service := range services.Data {
		stack := stackByID[service.StackID]
		if stack == nil {
			continue
		}

		if len(stacks) > 0 && !stacks[service.StackID] {
			continue
		}

		writer.Write(&ServiceData{
			ID:       service.ID,
			Created:  services.Data[i].Created,
			Service:  &services.Data[i],
			Stack:    stack,
			Endpoint: endpoint(&service),
		})
	}
	if p.E_External {
		externalServices, err := wc.ExternalService.List(util.DefaultListOpts())
		if err != nil {
			return err
		}

		for _, e := range externalServices.Data {
			stack := stackByID[e.StackID]
			if stack == nil {
				continue
			}

			if len(stacks) > 0 && !stacks[e.StackID] {
				continue
			}
			fakeService := &client.Service{}
			fakeService.Name = e.Name
			fakeService.State = "active"
			writer.Write(&ServiceData{
				ID:       e.ID,
				Created:  e.Created,
				Service:  fakeService,
				Stack:    stack,
				Endpoint: e.Target,
				External: "*",
			})
		}
	}

	return writer.Err()
}

func endpoint(service *client.Service) string {
	if len(service.Endpoints) > 0 {
		return service.Endpoints[0].URL
	}

	return ""
}
