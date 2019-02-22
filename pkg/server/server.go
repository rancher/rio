package server

import (
	"context"

	"github.com/rancher/norman"
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/api/setup"
	"github.com/rancher/rio/controllers"
	rTypes "github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/apiextensions.k8s.io/v1beta1"
	buildv1alpha1 "github.com/rancher/rio/types/apis/build.knative.dev/v1alpha1"
	cmv1alpha1 "github.com/rancher/rio/types/apis/certmanager.k8s.io/v1alpha1"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	policyv1beta1 "github.com/rancher/rio/types/apis/policy/v1beta1"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	projectschema "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1/schema"
	autoscalev1 "github.com/rancher/rio/types/apis/rio-autoscale.cattle.io/v1"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	rioschema "github.com/rancher/rio/types/apis/rio.cattle.io/v1/schema"
	storagev1 "github.com/rancher/rio/types/apis/storage.k8s.io/v1"
	webhookv1 "github.com/rancher/rio/types/apis/webhookinator.rio.cattle.io/v1"
	"github.com/rancher/types/apis/apps/v1beta2"
	v1 "github.com/rancher/types/apis/core/v1"
	v3 "github.com/rancher/types/apis/management.cattle.io/v3"
	rbacv1 "github.com/rancher/types/apis/rbac.authorization.k8s.io/v1"
)

func NewConfig(runDns bool) *norman.Config {
	return &norman.Config{
		Name: "rio",
		Schemas: []*types.Schemas{
			rioschema.Schemas,
			projectschema.Schemas,
		},

		CRDs: map[*types.APIVersion][]string{
			&rioschema.Version: {
				"service",
				"stack",
				"externalService",
				"routeSet",
				"volume",
				"config",
			},
			&projectschema.Version: {
				"setting",
				"publicDomain",
				"feature",
				"listenConfig",
			},
		},

		Clients: []norman.ClientFactory{
			autoscalev1.Factory,
			buildv1alpha1.Factory,
			cmv1alpha1.Factory,
			policyv1beta1.Factory,
			projectv1.Factory,
			rbacv1.Factory,
			riov1.Factory,
			storagev1.Factory,
			v1alpha3.Factory,
			v1beta1.Factory,
			v1beta2.Factory,
			v1.Factory,
			v3.Factory,
			webhookv1.Factory,
		},

		CustomizeSchemas: setup.Types,

		GlobalSetup: rTypes.BuildContext,

		MasterSetup: func(ctx context.Context) (context.Context, error) {
			rTypes.From(ctx).InCluster = runDns
			return ctx, nil
		},

		MasterControllers: []norman.ControllerRegister{
			rTypes.Register(controllers.Register),
		},
	}
}
