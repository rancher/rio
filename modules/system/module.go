package system

import (
	"context"

	"github.com/rancher/rio/modules/system/features/letsencrypt"
	"github.com/rancher/rio/modules/system/features/letsencrypt/pkg/issuers"
	"github.com/rancher/rio/modules/system/features/rdns"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	if err := ensureClusterDomain(rContext.Namespace, rContext.Global.Admin().V1().ClusterDomain()); err != nil {
		return err
	}

	if err := rdns.Register(ctx, rContext); err != nil {
		return err
	}
	return letsencrypt.Register(ctx, rContext)
}

func ensureClusterDomain(ns string, clusterDomain adminv1controller.ClusterDomainClient) error {
	_, err := clusterDomain.Get(ns, constants.ClusterDomainName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := clusterDomain.Create(adminv1.NewClusterDomain(ns, constants.ClusterDomainName, adminv1.ClusterDomain{
			Spec: adminv1.ClusterDomainSpec{
				SecretRef: v1.SecretReference{
					Namespace: ns,
					Name:      issuers.RioWildcardCerts,
				},
			},
		}))
		return err
	}
	return err
}
