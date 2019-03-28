package ps

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	mapper3 "github.com/rancher/types/mapper"

	"github.com/rancher/rio/cli/pkg/mapper"
	mapper2 "github.com/rancher/rio/types/mapper"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/pkg/namespace"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceData struct {
	ID       string
	Created  string
	Service  *riov1.Service
	Stack    *riov1.Stack
	Endpoint string
	External string
}

func FormatServiceName(cluster *clientcfg.Cluster) func(data, data2 interface{}) (string, error) {
	return func(data, data2 interface{}) (string, error) {
		stackName, ok := data.(string)
		if !ok {
			return "", nil
		}

		service, ok := data2.(*riov1.Service)
		if !ok {
			return "", nil
		}

		if service.Spec.Revision.ParentService == "" || service.Spec.Revision.Version == "" {
			return table.FormatStackScopedName(cluster)(stackName, service.Name)
		}

		return table.FormatStackScopedName(cluster)(stackName, service.Spec.Revision.ParentService+":"+service.Spec.Revision.Version)
	}
}

func FormatImage(data interface{}) (string, error) {
	s, ok := data.(*riov1.Service)
	if !ok {
		return fmt.Sprint(data), nil
	}
	if s.Spec.Image == "" || len(s.Spec.Sidekicks) > 0 {
		return s.Spec.Sidekicks[firstSortedKey(s.Spec.Sidekicks)].Image, nil
	}
	return s.Spec.Image, nil
}

func firstSortedKey(m map[string]riov1.SidekickConfig) string {
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
	scale, ok := data.(int)
	if !ok {
		return fmt.Sprint(data), nil
	}
	scaleStr := strconv.Itoa(scale)

	scaleStatus, ok := data2.(*riov1.ScaleStatus)
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
	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	project, err := ctx.Project()
	if err != nil {
		return err
	}

	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}

	domain, err := cluster.Domain()
	if err != nil {
		return err
	}

	writer := NewWriter(ctx, os.Stdout, cluster)
	defer writer.Close()

	stackByID, err := util.StacksByID(client, project.Project.Name)
	if err != nil {
		return err
	}

	services, err := ListAllMesh(stacks, stackByID, client, project.Project.Name, domain)
	if err != nil {
		return err
	}

	for _, d := range services {
		writer.Write(d)
	}
	return writer.Err()
}

func endpoint(service *riov1.Service) string {
	if len(service.Status.Endpoints) > 0 {
		return service.Status.Endpoints[0].URL
	}
	return ""
}

func listService(stacks map[string]bool, stackByID map[string]*riov1.Stack, client *clientcfg.KubeClient, project string) ([]ServiceData, error) {
	services, err := client.Rio.Services("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var filteredService []riov1.Service
	for _, s := range services.Items {
		if s.Spec.ProjectName == project {
			filteredService = append(filteredService, s)
		}
	}

	var data []ServiceData
	for i, service := range filteredService {
		stack := stackByID[service.Spec.StackName]
		if stack == nil {
			continue
		}

		if len(stacks) > 0 && !stacks[service.Spec.StackName] {
			continue
		}
		data = append(data, ServiceData{
			ID:       service.Name,
			Created:  filteredService[i].CreationTimestamp.String(),
			Service:  &filteredService[i],
			Stack:    stack,
			Endpoint: endpoint(&service),
		})
	}
	return data, nil
}

func listExternalService(stacks map[string]bool, stackByID map[string]*riov1.Stack, client *clientcfg.KubeClient, project string) ([]ServiceData, error) {
	externalServices, err := client.Rio.ExternalServices("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var filteredExternalServices []riov1.ExternalService
	for _, es := range externalServices.Items {
		if es.Spec.ProjectName == project {
			filteredExternalServices = append(filteredExternalServices, es)
		}
	}

	var data []ServiceData
	for _, e := range filteredExternalServices {
		stack := stackByID[e.Spec.StackName]
		if stack == nil {
			continue
		}

		if len(stacks) > 0 && !stacks[e.Spec.StackName] {
			continue
		}
		fakeService := &riov1.Service{}
		fakeService.Name = e.Name
		endpoint := ""
		if len(e.Spec.IPAddresses) > 0 {
			endpoint = strings.Join(e.Spec.IPAddresses, ",")
		} else if e.Spec.FQDN != "" {
			endpoint = e.Spec.FQDN
		} else if e.Spec.Service != "" {
			endpoint = e.Spec.Service
		}
		data = append(data, ServiceData{
			ID:       e.Name,
			Created:  e.CreationTimestamp.String(),
			Service:  fakeService,
			Stack:    stack,
			Endpoint: endpoint,
			External: "*",
		})
	}
	return data, nil
}

func listRoute(stacks map[string]bool, stackByID map[string]*riov1.Stack, client *clientcfg.KubeClient, project, domain string) ([]ServiceData, error) {
	routes, err := client.Rio.RouteSets("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var filteredRoutes []riov1.RouteSet
	for _, r := range routes.Items {
		if r.Spec.ProjectName == project {
			filteredRoutes = append(filteredRoutes, r)
		}
	}

	var data []ServiceData
	for _, r := range filteredRoutes {
		stack := stackByID[r.Spec.StackName]
		if stack == nil {
			continue
		}

		if len(stacks) > 0 && !stacks[r.Spec.StackName] {
			continue
		}
		fakeService := &riov1.Service{}
		fakeService.Name = r.Name
		endpoint := fmt.Sprintf("https://%s.%s", namespace.HashIfNeed(r.Name, stack.Name, project), domain)
		data = append(data, ServiceData{
			ID:       r.Name,
			Created:  r.CreationTimestamp.String(),
			Service:  fakeService,
			Stack:    stack,
			Endpoint: endpoint,
			External: "",
		})
	}
	return data, nil
}

func ListAllMesh(stacks map[string]bool, stackByID map[string]*riov1.Stack, client *clientcfg.KubeClient, project, domain string) ([]ServiceData, error) {
	services, err := listService(stacks, stackByID, client, project)
	if err != nil {
		return nil, err
	}

	externalServices, err := listExternalService(stacks, stackByID, client, project)
	if err != nil {
		return nil, err
	}

	routes, err := listRoute(stacks, stackByID, client, project, domain)
	if err != nil {
		return nil, err
	}

	return append(services, append(externalServices, routes...)...), nil
}

func NewWriter(ctx *clicontext.CLIContext, w io.Writer, cluster *clientcfg.Cluster) *table.Writer {
	writer := table.NewWriter([][]string{
		{"NAME", "{{serviceName .Stack.Name .Service}}"},
		{"IMAGE", "{{.Service | image}}"},
		{"CREATED", "{{.Service.CreationTimestamp | ago}}"},
		{"STATE", "{{.Service | toJson | stateMapper}}"},
		{"SCALE", "{{scale .Service.Spec.Scale .Service.Status.ScaleStatus}}"},
		{"ENDPOINT", "Endpoint"},
		{"EXTERNAL", "External"},
		{"DETAIL", "{{first (.Service | toJson | transitioning) (.Stack | toJson | transitioning)}}"},
	}, ctx, w)
	wrapper := mapper.Wrapper{}
	wrapper.AddMapper(&mapper2.DeploymentStatus{}, mapper2.Status{}, mapper3.Status{})
	writer.AddFormatFunc("stateMapper", wrapper.FormatState)
	writer.AddFormatFunc("transitioning", wrapper.FormatTransitionMessage)
	writer.AddFormatFunc("serviceName", FormatServiceName(cluster))
	writer.AddFormatFunc("image", FormatImage)
	writer.AddFormatFunc("scale", FormatScale)
	return writer
}
