package features

import (
	"context"

	ntypes "github.com/rancher/mapper"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
)

type ControllerRegister func(ctx context.Context, rContext *types.Context) error

type FeatureSpec struct {
	Enabled     bool              `json:"enabled,omitempty"`
	Description string            `json:"description,omitempty"`
	Questions   []riov1.Question  `json:"questions,omitempty"`
	Answers     map[string]string `json:"answers,omitempty"`
	Requires    []string          `json:"features,omitempty"`
}

type FeatureController struct {
	FeatureName  string
	System       bool
	FeatureSpec  FeatureSpec
	Controllers  []ControllerRegister
	OnStop       func() error
	OnStart      func() error
	SystemStacks []*stack.SystemStack
	FixedAnswers map[string]string
	registered   bool
}

func (f *FeatureController) Register() error {
	if f.registered {
		return nil
	}

	for _, ss := range f.SystemStacks {
		qs, err := ss.Questions()
		if err != nil {
			return err
		}

		f.FeatureSpec.Questions = append(f.FeatureSpec.Questions, qs...)
	}

	Register(f)
	return nil
}

func (f *FeatureController) Name() string {
	return f.FeatureName
}

func (f *FeatureController) IsSystem() bool {
	return f.System
}

func (f *FeatureController) Spec() FeatureSpec {
	return f.FeatureSpec
}

func (f *FeatureController) Stop() error {
	if f.OnStop != nil {
		return f.OnStop()
	}

	var errs []error
	for _, ss := range f.SystemStacks {
		if err := ss.Remove(); err != nil {
			errs = append(errs, err)
		}
	}

	return ntypes.NewErrors(errs...)
}

func (f *FeatureController) Start(ctx context.Context) error {
	var errs []error
	for _, ss := range f.SystemStacks {
		ans := map[string]string{}
		for k, v := range f.FeatureSpec.Answers {
			ans[k] = v
		}
		for k, v := range f.FixedAnswers {
			ans[k] = v
		}
		if err := ss.Deploy(ans); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return ntypes.NewErrors(errs...)
	}
	rContext := types.From(ctx)
	for _, reg := range f.Controllers {
		if err := reg(ctx, rContext); err != nil {
			return err
		}
	}

	if f.OnStart != nil {
		if err := f.OnStart(); err != nil {
			return err
		}
	}

	return nil
}
