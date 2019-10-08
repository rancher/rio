package externalservice

import (
	"context"

	"github.com/rancher/rio/modules/service/controllers/externalservice/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	v1.RegisterExternalServiceGeneratingHandler(ctx,
		rContext.Rio.Rio().V1().ExternalService(),
		rContext.Apply.WithCacheTypes(rContext.Core.Core().V1().Service(),
			rContext.Core.Core().V1().Endpoints()),
		"ExternalServiceDeployed",
		"externalservice",
		generate,
		nil)

	return nil
}

func generate(obj *riov1.ExternalService, status riov1.ExternalServiceStatus) ([]runtime.Object, riov1.ExternalServiceStatus, error) {
	os := objectset.NewObjectSet()
	err := populate.ServiceForExternalService(obj, os)
	return os.All(), status, err
}
