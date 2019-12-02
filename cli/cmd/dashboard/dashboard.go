package dashboard

import (
	"net/url"
	"strings"

	"github.com/pkg/browser"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/progress"
	v3 "github.com/rancher/rio/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/rio/pkg/config"
	"github.com/rancher/rio/pkg/randomtoken"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Dashboard struct {
	ResetAdmin bool `desc:"Reset admin password"`
}

func (d *Dashboard) Run(ctx *clicontext.CLIContext) error {
	if err := enableDashboard(ctx); err != nil {
		return err
	}

	dashboardURL, err := waitDashboard(ctx)
	if err != nil {
		return err
	}

	token, err := setup(ctx, d.ResetAdmin)
	if err != nil {
		return err
	}

	u, err := url.Parse(dashboardURL)
	if err != nil {
		return err
	}

	if token != "" {
		q := u.Query()
		q.Set("setup", token)
		u.RawQuery = q.Encode()
	}
	u.Path = "/dashboard/"

	q := u.String()
	logrus.Infof("Opening browser to %s", q)
	return browser.OpenURL(u.String())
}

func reset(ctx *clicontext.CLIContext) error {
	err := ctx.Mgmt.Users().Delete("admin", nil)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	err = ctx.Mgmt.Settings().Delete("first-login", nil)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	p := progress.NewWriter()
	for {
		_ = ctx.Mgmt.Users().Delete("admin", nil)
		_, err := ctx.Mgmt.Users().Get("admin", metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return nil
		}
		p.Display("Resetting admin", 1)
	}
}

func setup(ctx *clicontext.CLIContext, resetAdmin bool) (string, error) {
	if resetAdmin {
		if err := reset(ctx); err != nil {
			return "", err
		}
	}

	adm, err := ctx.Mgmt.Users().Get("admin", metav1.GetOptions{})
	if err == nil {
		if strings.HasPrefix(adm.Password, "$") {
			return "", nil
		}
		return adm.Password, nil
	}

	err = ctx.Mgmt.Settings().Delete("first-login", nil)
	if err != nil && !errors.IsNotFound(err) {
		return "", err
	}

	ctx.Mgmt.Settings().Create(&v3.Setting{
		ObjectMeta: metav1.ObjectMeta{
			Name: "first-login",
		},
		Default:    "true",
		Value:      "true",
		Customized: true,
	})

	token, err := randomtoken.Generate()
	if err != nil {
		return "", err
	}

	_, err = ctx.Mgmt.Users().Create(&v3.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "admin",
		},
		Username:           "admin",
		Password:           token,
		MustChangePassword: true,
	})

	return token, err
}

func waitDashboard(ctx *clicontext.CLIContext) (string, error) {
	w := progress.NewWriter()
	first := true
	for {
		if !first {
			w.Display("Waiting for dashboard service to be ready", 1)
		}
		first = false

		svc, err := ctx.Rio.Services(ctx.SystemNamespace).Get("dashboard", metav1.GetOptions{})
		if errors.IsNotFound(err) {
			continue
		} else if err != nil {
			return "", err
		}

		if !svc.Status.DeploymentReady ||
			len(svc.Status.AppEndpoints) == 0 ||
			!strings.HasPrefix(svc.Status.AppEndpoints[0], "https://") {
			continue
		}

		return svc.Status.AppEndpoints[0], nil
	}
}

func enableDashboard(ctx *clicontext.CLIContext) error {
	cm, err := ctx.Core.ConfigMaps(ctx.SystemNamespace).Get(config.ConfigName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	cfg, err := config.FromConfigMap(cm)
	if err != nil {
		return err
	}

	f := cfg.Features["dashboard"]
	if f.Enabled != nil && *f.Enabled {
		return nil
	}

	f.Enabled = &[]bool{true}[0]
	if cfg.Features == nil {
		cfg.Features = map[string]config.FeatureConfig{}
	}

	cfg.Features["dashboard"] = f

	cm, err = config.SetConfig(cm, cfg)
	if err != nil {
		return err
	}

	_, err = ctx.Core.ConfigMaps(ctx.SystemNamespace).Update(cm)
	return err
}
