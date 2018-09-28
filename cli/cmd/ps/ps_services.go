package ps

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/kv"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type ServiceData struct {
	ID       string
	Service  *client.Service
	Stack    *client.Stack
	Endpoint string
}

func FormatImage(data interface{}) (string, error) {
	s, ok := data.(*client.Service)
	if !ok {
		return fmt.Sprint(data), nil
	}
	if s.Image == "" && len(s.Sidekicks) > 0 {
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

func (p *Ps) services(app *cli.Context, ctx *server.Context) error {
	services, err := ctx.Client.Service.List(util.DefaultListOpts())
	if err != nil {
		return err
	}

	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Stack.Name .Service.Name}}"},
		{"IMAGE", "{{.Service | image}}"},
		{"CREATED", "{{.Service.Created | ago}}"},
		{"SCALE", "{{scale .Service.Scale .Service.ScaleStatus}}"},
		{"STATE", "Service.State"},
		{"ENDPOINT", "Endpoint"},
		{"DETAIL", "{{first .Service.TransitioningMessage .Stack.TransitioningMessage}}"},
	}, app)
	defer writer.Close()

	writer.AddFormatFunc("image", FormatImage)
	writer.AddFormatFunc("scale", FormatScale)

	stackByID, err := util.StacksByID(ctx)
	if err != nil {
		return err
	}

	for i, service := range services.Data {
		stack := stackByID[service.StackID]
		if stack == nil {
			continue
		}

		writer.Write(&ServiceData{
			ID:       service.ID,
			Service:  &services.Data[i],
			Stack:    stack,
			Endpoint: endpoint(ctx, stack, service.PortBindings, &service),
		})

		for revName, revision := range service.Revisions {
			newService := &client.Service{}
			if err := convert.ToObj(&revision, newService); err != nil {
				return err
			}
			newService.Name += service.Name + ":" + revName
			newService.Created = service.Created
			if newService.Image == "" {
				newService.Image = service.Image
			}

			writer.Write(&ServiceData{
				ID:      service.ID,
				Service: newService,
				Stack:   stack,
				// use parent service ports
				Endpoint: endpoint(ctx, stack, service.PortBindings, newService),
			})
		}
	}

	return writer.Err()
}

func endpoint(ctx *server.Context, stack *client.Stack, ports []client.PortBinding, service *client.Service) string {
	if ctx.Domain == "" || stack == nil {
		return ""
	}

	for _, port := range ports {
		if port.Protocol == "http" {
			name, rev := kv.Split(service.Name, ":")
			if rev != "" && rev != "latest" {
				name = name + "-" + rev
			}
			domain := fmt.Sprintf("%s.%s.%s", name, stack.Name, ctx.Domain)

			return "http://" + domain
		}
	}

	return ""
}
