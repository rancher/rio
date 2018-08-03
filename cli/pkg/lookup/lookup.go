package lookup

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	client2 "github.com/rancher/rio/types/client/space/v1beta1"
)

func Lookup(c clientbase.APIBaseClientInterface, name string, typeNames ...string) (*types.Resource, error) {
	var result []*namedResource
	for _, schemaType := range typeNames {
		if strings.Contains(name, ":") && !strings.Contains(name, "/") {
			resourceByID, err := byID(c, name, schemaType)
			if err == nil {
				return &resourceByID.Resource, nil
			}
		}

		byName, err := byName(c, name, schemaType)
		if err != nil {
			return nil, err
		}

		if byName != nil {
			result = append(result, byName)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("not found: %s", name)
	}

	if len(result) == 1 {
		return &result[0].Resource, nil
	}

	for {
		questions.PrintfToTerm("Choose resource for %s:\n", name)
		for i, r := range result {
			msg := fmt.Sprintf("[%d] type=%s %s(%s)", i+1, r.Type, r.Name, r.ID)
			if len(r.Description) > 0 {
				msg += ": " + r.Description
			}
			questions.PrintlnToTerm(msg)
		}

		ans, err := questions.Prompt("Select Number [] ", "")
		if err != nil {
			return nil, err
		}
		num, err := strconv.Atoi(ans)
		if err != nil {
			questions.PrintfToTerm("invalid number: %s\n", ans)
			continue
		}

		num--
		if num < 0 || num >= len(result) {
			questions.PrintlnToTerm("Select a number between 1 and", +len(result))
			continue
		}

		return &result[num].Resource, nil
	}
}

type namedResourceCollection struct {
	types.Collection
	Data []namedResource `json:"data,omitempty"`
}

type namedResource struct {
	types.Resource
	Name        string `json:"name"`
	Description string `json:"description"`
}

func byID(c clientbase.APIBaseClientInterface, id, schemaType string) (*namedResource, error) {
	var resource namedResource

	err := c.ByID(schemaType, id, &resource)
	return &resource, err
}

func parseStackServiceName(name string) (string, string) {
	var (
		stackName   string
		serviceName string
	)

	// logic for routes
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, "://", 2)
		if len(parts) > 1 {
			parts[0] = parts[1]
		}
		parts = strings.Split(parts[0], ".")
		if len(parts) == 1 {
			stackName = "default"
			serviceName = parts[0]
		} else {
			stackName = parts[1]
			serviceName = parts[0]
		}
	} else {
		parsedService := ParseServiceName(name)
		stackName, serviceName = parsedService.StackName, parsedService.ServiceName
	}

	return stackName, serviceName
}

func setupFilters(c clientbase.APIBaseClientInterface, name, schemaType string) (map[string]interface{}, error) {
	filters := map[string]interface{}{
		"name":         name,
		"removed_null": "1",
	}

	if schemaType == client.StackType {
		return filters, nil
	}

	stackName, serviceName := parseStackServiceName(name)

	stack, err := byName(c, stackName, client.StackType)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Unknown schema type [") {
			stack = nil
		} else {
			return nil, err
		}
	}

	if stack != nil {
		filters["stackId"] = stack.ID
	}
	filters["name"] = serviceName

	return filters, nil
}

func byName(c clientbase.APIBaseClientInterface, name, schemaType string) (*namedResource, error) {
	var collection namedResourceCollection

	if schemaType == client.StackType && strings.Contains(name, "/") {
		// stacks can't be foo/bar
		return nil, nil
	}

	if schemaType == client2.PodType {
		container, ok := ParseContainerName(name)
		if ok {
			name = container.PodName
		}
	}

	filters, err := setupFilters(c, name, schemaType)
	if err != nil {
		return nil, err
	}

	if err := c.List(schemaType, &types.ListOpts{
		Filters: filters,
	}, &collection); err != nil {
		return nil, err
	}

	if len(collection.Data) > 1 {
		var ids []string
		for _, data := range collection.Data {
			switch schemaType {
			default:
				ids = append(ids, fmt.Sprintf("%s (%s)", data.ID, name))
			}

		}
		index := selectFromList("Resources: ", ids)
		return &collection.Data[index], nil
	}

	if len(collection.Data) == 0 {
		return nil, nil
	}

	return &collection.Data[0], nil
}

func selectFromList(header string, choices []string) int {
	if header != "" {
		fmt.Println(header)
	}

	reader := bufio.NewReader(os.Stdin)
	selected := -1
	for selected <= 0 || selected > len(choices) {
		for i, choice := range choices {
			fmt.Printf("[%d] %s\n", i+1, choice)
		}
		fmt.Print("Select: ")

		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		num, err := strconv.Atoi(text)
		if err == nil {
			selected = num
		}
	}
	return selected - 1
}
