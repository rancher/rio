package edit

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/cmd/edit/pretty"
	"github.com/rancher/rio/cli/cmd/edit/raw"
	"github.com/rancher/rio/cli/cmd/edit/stack"
	"github.com/rancher/rio/cli/cmd/inspect"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	"github.com/rancher/wrangler/pkg/gvk"
	name2 "github.com/rancher/wrangler/pkg/name"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
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
	Raw    bool   `desc:"Edit the raw API object, not the pretty formatted one"`
	T_Type string `desc:"Specific type to edit"`
}

func (edit *Edit) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("at least one parameter is required")
	}

	return edit.edit(ctx)
}

type Editor interface {
	Edit(obj runtime.Object) (updated bool, err error)
}

func (edit *Edit) edit(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("exactly one Name (not name) arguement is required for raw edit")
	}

	var (
		name  = ctx.CLI.Args()[0]
		types []string
	)

	for _, t := range inspect.InspectTypes {
		if t == clitypes.AppType {
			continue
		}
		types = append(types, t)
	}

	r, err := lookup.Lookup(ctx, name, types...)
	if err != nil {
		return err
	}

	g, err := gvk.Get(r.Object)
	if err != nil {
		return err
	}
	gvr := schema.GroupVersionResource{
		Group:    g.Group,
		Version:  g.Version,
		Resource: strings.ToLower(name2.GuessPluralName(g.Kind)),
	}

	c, err := dynamic.NewForConfig(ctx.RestConfig)
	if err != nil {
		return err
	}
	u := updater{
		client: c,
		gvr:    gvr,
		gvk: schema.GroupVersionKind{
			Group:   g.Group,
			Version: g.Version,
			Kind:    g.Kind,
		},
	}

	editor := edit.getEditor(r.Type, u)
	updated, err := editor.Edit(r.Object)
	if err != nil {
		return err
	}

	if !updated {
		logrus.Infof("No change for %s/%s", r.Namespace, r.Name)
	}

	return nil
}

type updater struct {
	gvr    schema.GroupVersionResource
	gvk    schema.GroupVersionKind
	client dynamic.Interface
}

func (u updater) Update(obj runtime.Object) error {
	toUpdate, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("object is not an unstructured object")
	}
	_, err := u.client.Resource(u.gvr).Namespace(toUpdate.GetNamespace()).Update(toUpdate, v1.UpdateOptions{})
	return err
}

func (u updater) GetGvk() schema.GroupVersionKind {
	return u.gvk
}

func (edit Edit) getEditor(t string, u updater) Editor {
	if t == clitypes.StackType && !edit.Raw {
		return stack.NewEditor(u)
	}

	if (t == clitypes.ServiceType || t == clitypes.ConfigType || t == clitypes.RouterType || t == clitypes.ExternalServiceType) && !edit.Raw {
		return pretty.NewEditor(u)
	}

	return raw.NewRawEditor(u)
}
