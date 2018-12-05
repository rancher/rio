package data

import (
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/space.cattle.io/v1beta1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DefaultFeatureList = []*v1beta1.Feature{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nfs",
		},
		Spec: v1beta1.FeatureSpec{
			Description: "Enable nfs volume feature",
			Enable:      false,
			Questions: []v3.Question{
				{
					Variable:    "NFS_SERVER_HOSTNAME",
					Description: "Hostname of NFS server",
				},
				{
					Variable:    "NFS_SERVER_EXPORT_PATH",
					Description: "Export path of NFS server",
				},
			},
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name: "monitoring",
		},
		Spec: v1beta1.FeatureSpec{
			Description: "Enable monitoring feature",
			Enable:      false,
		},
	},
}

func addFeatures(rContext *types.Context) error {
	for _, feature := range DefaultFeatureList {
		if _, err := rContext.Global.Feature.Create(feature); err != nil && !errors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}
