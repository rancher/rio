package stack

import (
	"github.com/rancher/norman/api/access"
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func ListHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	if apiContext.ID != "" &&
		apiContext.ResponseFormat == "yaml" &&
		(apiContext.Option("edit") == "true" || apiContext.Option("export") == "true") {
		stack := &client.Stack{}
		err := access.ByID(apiContext, apiContext.Version, client.StackType, apiContext.ID, stack)
		if err != nil {
			return err
		}

		if stack.Template != "" {
			apiContext.Response.Header().Set("Content-Type", "application/yaml")
			_, err := apiContext.Response.Write(append([]byte(stack.Template), '\n'))
			return err
		}
	}

	return next(apiContext, nil)
}
