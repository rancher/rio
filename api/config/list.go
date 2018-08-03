package config

import (
	"encoding/base64"

	"github.com/rancher/norman/api/access"
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func ListHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	if apiContext.ID != "" &&
		apiContext.ResponseFormat == "yaml" &&
		(apiContext.Option("edit") == "true" || apiContext.Option("export") == "true") {
		config := &client.Config{}
		err := access.ByID(apiContext, apiContext.Version, client.ConfigType, apiContext.ID, config)
		if err != nil {
			return err
		}

		bytes := []byte(config.Content)
		if config.Encoded {
			bytes, err = base64.StdEncoding.DecodeString(config.Content)
			if err != nil {
				return err
			}
		}

		apiContext.Response.Header().Set("Content-Type", "text/plain")
		_, err = apiContext.Response.Write(bytes)
		return err
	}

	return next(apiContext, nil)
}
