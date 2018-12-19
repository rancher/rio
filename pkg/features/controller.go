package features

import (
	"context"

	"github.com/rancher/norman/controller"
	ntypes "github.com/rancher/norman/types"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
)

type ControllerRegister func(ctx context.Context, rContext *types.Context) error

type FeatureController struct {
	FeatureName  string
	FeatureSpec  v1.FeatureSpec
	Controllers  []ControllerRegister
	OnStop       func() error
	OnChange     func(*v1.Feature) error
	OnStart      func(*v1.Feature) error
	SystemStacks []*systemstack.SystemStack
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

func (f *FeatureController) Spec() v1.FeatureSpec {
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

func (f *FeatureController) Changed(feature *v1.Feature) error {
	if f.OnChange != nil {
		if err := f.OnChange(feature); err != nil {
			return err
		}
	}

	return nil
}

func (f *FeatureController) Start(ctx context.Context, feature *v1.Feature) error {
	var errs []error
	for _, ss := range f.SystemStacks {
		ans := map[string]string{}
		for k, v := range feature.Spec.Answers {
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

	if err := controller.SyncThenStart(ctx, 5, rContext.Starters()...); err != nil {
		return err
	}

	if f.OnStart != nil {
		if err := f.OnStart(feature); err != nil {
			return err
		}
	}

	return nil
}
