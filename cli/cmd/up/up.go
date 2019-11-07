package up

import (
	"fmt"

	"github.com/rancher/rio/cli/cmd/up/pkg"
	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/riofile/stringers"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/wrangler/pkg/gvk"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Up struct {
	Name               string   `desc:"Set stack name, defaults to current directory name"`
	Answers            string   `desc:"Set answer file"`
	F_File             string   `desc:"Set rio file"`
	P_Parallel         bool     `desc:"Run builds in parallel"`
	Branch             string   `desc:"Set branch when pointing stack to git repo" default:"master"`
	Revision           string   `desc:"Set revision"`
	BuildWebhookSecret string   `desc:"Set GitHub webhook secret name"`
	BuildCloneSecret   string   `desc:"Set git clone secret name"`
	Permission         []string `desc:"Permissions to grant to container's service account in current namespace"`
}

const (
	stackLabel = "rio.cattle.io/stack"
)

func (u *Up) Run(c *clicontext.CLIContext) error {
	if u.Name == "" {
		u.Name = pkg.GetCurrentDir()
	}

	stack, err := u.ensureStack(c)
	if err != nil {
		return err
	}

	// if format is `rio up https://repo`, set build parameters
	if len(c.CLI.Args()) > 0 {
		if err := u.setStack(c, stack); err != nil {
			return err
		}
		return c.UpdateObject(stack)
	}

	content, answer, err := u.loadFileAndAnswer(c)
	if err != nil {
		return err
	}

	return u.up(content, answer, stack, c)
}

func (u *Up) setStack(c *clicontext.CLIContext, existing *riov1.Stack) error {
	if len(c.CLI.Args()) == 1 {
		var err error
		if existing.Spec.Build == nil {
			existing.Spec.Build = &riov1.StackBuild{}
		}
		existing.Spec.Build.Repo = c.CLI.Args()[0]
		existing.Spec.Build.Branch = u.Branch
		existing.Spec.Build.Revision = u.Revision
		existing.Spec.Build.WebhookSecretName = u.BuildWebhookSecret
		existing.Spec.Build.CloneSecretName = u.BuildCloneSecret
		existing.Spec.Permissions, err = stringers.ParsePermissions(u.Permission...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *Up) setBuild(c *clicontext.CLIContext) error {
	s := riov1.NewStack(c.GetSetNamespace(), u.Name, riov1.Stack{
		Spec: riov1.StackSpec{
			Build: &riov1.StackBuild{
				Repo:     c.CLI.Args()[0],
				Branch:   u.Branch,
				Revision: u.Revision,
			},
		},
	})

	existing, err := c.Rio.Stacks(c.GetSetNamespace()).Get(u.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return c.Create(s)
		}
	}
	if existing.Spec.Build == nil {
		existing.Spec.Build = &riov1.StackBuild{}
	}
	existing.Spec.Build.Repo = c.CLI.Args()[0]
	existing.Spec.Build.Branch = u.Branch
	existing.Spec.Build.Revision = u.Revision
	return c.UpdateObject(existing)
}

func (u *Up) loadFileAndAnswer(c *clicontext.CLIContext) (string, map[string]string, error) {
	answers, err := pkg.LoadAnswer(u.Answers)
	if err != nil {
		return "", nil, err
	}

	content, err := pkg.LoadRiofile(u.F_File)
	if err != nil {
		return "", nil, err
	}
	return string(content), answers, nil
}

func (u *Up) up(content string, answers map[string]string, s *riov1.Stack, c *clicontext.CLIContext) error {
	deployStack := stack.NewStack([]byte(content), answers)
	imageBuilds, err := deployStack.GetImageBuilds()
	if err != nil {
		return err
	}

	images, err := pkg.Build(imageBuilds, c, u.P_Parallel)
	if err != nil {
		return err
	}

	if err := deployStack.SetServiceImages(images); err != nil {
		return err
	}

	objs, err := deployStack.GetObjects()
	if err != nil {
		return err
	}
	objs, err = setObjLabels(objs, s)
	if err != nil {
		return err
	}
	gvks, err := convertObjs(objs)
	if err != nil {
		return fmt.Errorf("error converting objs to gvks, stack may be out of date: %w", err)
	}

	var knowngvks []schema.GroupVersionKind
	if len(s.Spec.AdditionalGroupVersionKinds) == 0 {
		knowngvks = gvks
	} else {
		knowngvks = s.Spec.AdditionalGroupVersionKinds
	}

	err = c.Apply.WithListerNamespace(c.GetSetNamespace()).WithDefaultNamespace(c.GetSetNamespace()).WithOwner(s).WithSetOwnerReference(true, true).WithGVK(knowngvks...).WithDynamicLookup().ApplyObjects(objs...)
	if err != nil {
		return err
	}

	stackToUpdate := s.DeepCopy()
	stackToUpdate.Spec.AdditionalGroupVersionKinds = gvks
	return c.UpdateObject(stackToUpdate)
}

// ensureStack creates one if one does not exist
func (u *Up) ensureStack(c *clicontext.CLIContext) (*riov1.Stack, error) {
	s := riov1.NewStack(c.GetSetNamespace(), u.Name, riov1.Stack{})

	existing, err := c.Rio.Stacks(c.GetSetNamespace()).Get(u.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return c.Rio.Stacks(c.GetSetNamespace()).Create(s)
		}
		return nil, err
	}

	return existing, err
}

func setObjLabels(objs []runtime.Object, s *riov1.Stack) ([]runtime.Object, error) {
	for _, obj := range objs {

		obj, ok := obj.(metav1.Object)
		if !ok {
			return nil, fmt.Errorf("runtime.Object failed type assertion with metav1.Object")
		}
		labels := obj.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		}
		// match debugID of stack
		labels[stackLabel] = s.Name
		obj.SetLabels(labels)
	}

	return objs, nil
}

func convertObjs(objs []runtime.Object) ([]schema.GroupVersionKind, error) {
	gvks := make([]schema.GroupVersionKind, 0, len(objs))
	for _, obj := range objs {
		groupVersionKind, err := gvk.Get(obj)
		if err != nil {
			return nil, err
		}
		if groupVersionKind.Empty() {
			return nil, fmt.Errorf("groupVersionKind shouldn't be empty")
		}
		gvks = append(gvks, groupVersionKind)
	}

	seen := map[string]bool{}
	var r []schema.GroupVersionKind
	for _, gvk := range gvks {
		if seen[gvk.String()] {
			continue
		}
		r = append(r, gvk)
		seen[gvk.String()] = true
	}
	return r, nil
}
