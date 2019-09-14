package info

import (
	"context"

	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/version"
	"github.com/rancher/rio/types"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	client := rContext.Global.Admin().V1().RioInfo()
	newInfo := adminv1.NewRioInfo("", "rio", adminv1.RioInfo{
		Status: adminv1.RioInfoStatus{
			SystemNamespace: rContext.Namespace,
			Version:         version.Version,
			GitCommit:       version.GitCommit,
		},
	})

	info, err := client.Get("rio", metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = client.Create(newInfo)
	} else if err == nil {
		info.Status.SystemNamespace = newInfo.Status.SystemNamespace
		info.Status.Version = newInfo.Status.Version
		info.Status.GitCommit = newInfo.Status.GitCommit
		_, err = client.Update(info)
	}

	return err
}
