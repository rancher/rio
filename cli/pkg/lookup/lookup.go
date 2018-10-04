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
)

type ClientLookup interface {
	ClientLookup(typeName string) (clientbase.APIBaseClientInterface, error)
	LookupFilters(name, typeName string) (map[string]interface{}, bool, error)
	ByID(id, typeName string) (*types.NamedResource, error)
}

func Lookup(c ClientLookup, name string, typeNames ...string) (*types.NamedResource, error) {
	var result []*types.NamedResource
	for _, schemaType := range typeNames {
		resourceByID, err := c.ByID(name, schemaType)
		if err == nil && resourceByID != nil {
			return resourceByID, nil
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
		return nil, fmt.Errorf("not found (types=%v): %s", typeNames, name)
	}

	if len(result) == 1 {
		return result[0], nil
	}

	msg := fmt.Sprintf("Choose resource for %s:\n", name)
	var options []string

	for i, r := range result {
		msg := fmt.Sprintf("[%d] type=%s %s(%s)\n", i+1, r.Type, r.Name, r.ID)
		if len(r.Description) > 0 {
			msg += ": " + r.Description
		}
		options = append(options, msg)
	}

	num, err := questions.PromptOptions(msg, -1, options...)
	if err != nil {
		return nil, err
	}
	return result[num], nil
}

func byName(c ClientLookup, name, schemaType string) (*types.NamedResource, error) {
	var collection types.NamedResourceCollection

	filters, ok, err := c.LookupFilters(name, schemaType)
	if err != nil || !ok {
		return nil, err
	}

	client, err := c.ClientLookup(schemaType)
	if err != nil {
		return nil, err
	}

	if err := client.List(schemaType, &types.ListOpts{
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
