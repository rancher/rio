package yamldownload

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/server"
)

func DownloadYAML(ctx *server.Context, contentType, option string, name string, resourceTypes ...string) (*types.Resource, io.ReadCloser, string, error) {
	obj, err := lookup.Lookup(ctx.Client, name, resourceTypes...)
	if err != nil {
		return nil, nil, "", err
	}

	body, err := download(ctx, contentType, option, obj.Links["self"])
	if err != nil {
		return nil, nil, "", err
	}

	return obj, body, obj.Links["self"], nil
}

func download(ctx *server.Context, contentType, option, self string) (io.ReadCloser, error) {
	parsed, err := url.Parse(self)
	if err != nil {
		return nil, err
	}
	q := parsed.Query()
	q.Set("_"+option, "true")
	parsed.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}

	ctx.Client.Ops.SetupRequest(req)
	req.Header.Set("Accept", contentType)

	resp, err := ctx.Client.Ops.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("invalid response code getting %s: %d", self, resp.StatusCode)
	}

	return resp.Body, nil
}
