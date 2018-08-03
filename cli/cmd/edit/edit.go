package edit

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/cmd/config"
	"github.com/rancher/rio/cli/pkg/up"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/cli/pkg/yamldownload"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
)

const (
	format = "application/yaml"
)

var (
	editTypes = []string{
		client.StackType,
		client.ServiceType,
		client.ConfigType,
		client.RouteSetType,
	}
)

type Edit struct {
	Prompt bool   `desc:"When editing a stack re-ask all questions if not found in environment variables"`
	Raw    bool   `desc:"Edit the raw API object, not the pretty formatted one"`
	T_Type string `desc:"Specific type to edit"`
}

func (edit *Edit) Run(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	waiter, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	if edit.Raw {
		return edit.rawEdit(app, ctx)
	}

	args := app.Args()
	if len(args) == 0 {
		args = []string{ctx.DefaultStackName}
	}

	for _, arg := range args {
		obj, body, url, err := yamldownload.DownloadYAML(ctx, format, "edit", arg, editTypes...)
		if err != nil {
			return err
		}
		defer body.Close()

		var prefix []byte
		input, err := ioutil.ReadAll(body)
		if err != nil {
			return err
		}

		updated, err := editLoop(prefix, input, func(content []byte) error {
			if err := edit.update(ctx, format, obj, url, content); err != nil {
				return err
			}
			waiter.Add(obj.ID)
			return nil
		})

		if err != nil {
			return err
		}

		if !updated {
			logrus.Infof("No change for %s(%s)", arg, obj.ID)
		}
	}

	return waiter.Wait()
}

type updateFunc func(content []byte) error

func editLoop(prefix, input []byte, update updateFunc) (bool, error) {
	for {
		buf := &bytes.Buffer{}
		buf.Write(prefix)
		buf.Write(input)
		rawInput := buf.Bytes()

		e := editor.NewDefaultEditor(os.Environ())
		content, path, err := e.LaunchTempFile("rio-", "-edit.yaml", buf)
		if path != "" {
			defer os.Remove(path)
		}
		if err != nil {
			return false, err
		}

		if bytes.Compare(content, rawInput) != 0 {
			content = bytes.TrimPrefix(content, prefix)
			input = content
			if err := update(content); err != nil {
				prefix = []byte(fmt.Sprintf("#\n# Error updating content:\n#    %v\n#\n", err.Error()))
				continue
			}
		} else {
			return false, nil
		}

		break
	}

	return true, nil
}

func (edit *Edit) update(ctx *server.Context, format string, obj *types.Resource, self string, content []byte) error {
	if obj.Type == client.StackType {
		return up.Run(ctx, content, obj.ID, true, edit.Prompt, nil)
	}

	if obj.Type == client.ConfigType {
		return config.RunUpdate(ctx, obj.ID, content, nil)
	}

	parsed, err := url.Parse(self)
	if err != nil {
		return err
	}

	q := parsed.Query()
	q.Set("_edited", "true")
	q.Set("_replace", "true")
	parsed.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodPut, parsed.String(), bytes.NewReader(content))
	if err != nil {
		return err
	}

	ctx.Client.Ops.SetupRequest(req)
	req.Header.Set("Content-Type", format)

	resp, err := ctx.Client.Ops.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode >= 300 {
		return clientbase.NewAPIError(resp, parsed.String())
	}

	return nil
}
