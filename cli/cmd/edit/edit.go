package edit

import (
	"bytes"
	"fmt"
	"os"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/riofile"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
)

const (
	format = "application/yaml"
)

var (
	editTypes = []string{
		clitypes.NamespaceType,
		clitypes.ServiceType,
		clitypes.ConfigType,
		clitypes.RouterType,
		clitypes.ExternalServiceType,
		clitypes.FeatureType,
	}
)

type Edit struct {
	Prompt bool   `desc:"When editing a stack re-ask all questions if not found in environment variables"`
	Raw    bool   `desc:"Edit the raw API object, not the pretty formatted one"`
	T_Type string `desc:"Specific type to edit"`
}

func (edit *Edit) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("at least one parameter is required")
	}

	if edit.Raw {
		return edit.rawEdit(ctx)
	}

	return edit.Edit(ctx, ctx.CLI.Args()[0])
}

func (edit *Edit) Edit(ctx *clicontext.CLIContext, arg string) error {
	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("exactly one ID (not name) arguement is required for raw edit")
	}

	r, err := lookup.Lookup(ctx, arg, clitypes.ServiceType)
	if err != nil {
		return err
	}

	r, err = ctx.ByID(r.Namespace, r.Name, clitypes.ServiceType)
	if err != nil {
		return err
	}

	services := make(map[string]riov1.Service)
	configs := make(map[string]corev1.ConfigMap)
	obj := r.Object
	switch obj.(type) {
	case *riov1.Service:
		newSvc := riov1.Service{}
		newSvc.Spec = obj.(*riov1.Service).Spec
		services[r.Name] = newSvc
	case *corev1.ConfigMap:
		configs[r.Name] = *obj.(*corev1.ConfigMap)
	}

	content, err := riofile.ParseFrom(services, configs)
	if err != nil {
		return err
	}

	updated, err := EditLoop(nil, content, func(content []byte) error {
		stack := systemstack.NewStack(ctx.Apply, r.Namespace, "edit-"+r.Name, true)
		stack.WithContent(content)
		return stack.Deploy(nil)
	})
	if err != nil {
		return err
	}

	if !updated {
		logrus.Infof("No change for %s/%s", r.Namespace, r.Name)
	}

	return nil
}

type updateFunc func(content []byte) error

func EditLoop(prefix, input []byte, update updateFunc) (bool, error) {
	for {
		buf := &bytes.Buffer{}
		buf.Write(prefix)
		buf.Write(input)
		rawInput := buf.Bytes()

		editors := []string{
			"KUBE_EDITOR",
			"EDITOR",
		}
		e := editor.NewDefaultEditor(editors)
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

//
//func (edit *Edit) update(ctx *clicontext.CLIContext, format string, obj *types.Resource, self string, content []byte) error {
//	if obj.Type == clitypes.NamespaceType {
//		return up.Run(ctx, content, obj.ID, true, edit.Prompt, nil, "")
//	}
//
//	if obj.Type == clitypes.ConfigType {
//		return config.RunUpdate(ctx, obj.ID, content, nil)
//	}
//
//	parsed, err := url.Parse(self)
//	if err != nil {
//		return err
//	}
//
//	q := parsed.Query()
//	q.Set("_edited", "true")
//	q.Set("_replace", "true")
//	parsed.RawQuery = q.Encode()
//
//	req, err := http.NewRequest(http.MethodPut, parsed.String(), bytes.NewReader(content))
//	if err != nil {
//		return err
//	}
//
//	wc, err := ctx.ProjectClient()
//	if err != nil {
//		return err
//	}
//
//	wc.Ops.SetupRequest(req)
//	req.Header.Set("Content-Type", format)
//
//	resp, err := wc.Ops.Client.Do(req)
//	if err != nil {
//		return err
//	}
//	defer func() {
//		io.Copy(ioutil.Discard, resp.Body)
//		resp.Body.Close()
//	}()
//
//	if resp.StatusCode >= 300 {
//		return clientbase.NewAPIError(resp, parsed.String())
//	}
//
//	return nil
//}
