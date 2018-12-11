package data

import (
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DefaultFeatureList = []*projectv1.Feature{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nfs",
		},
		Spec: projectv1.FeatureSpec{
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
		Spec: projectv1.FeatureSpec{
			Description: "Enable monitoring feature",
			Enable:      false,
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name: "letsencrypt",
		},
		Spec: projectv1.FeatureSpec{
			Description: "Enable let's encrypt feature",
			Enable:      true,
			Questions: []v3.Question{
				{
					Variable:    settings.RioWildcardType,
					Description: "Type of certificates for rio wildcards domain",
					Default:     settings.StagingType,
					Options:     []string{settings.StagingType, settings.ProductionType, settings.SelfSignedType},
				},
				{
					Variable:    settings.PublicDomainType,
					Description: "Type of certificates for rio public domain",
					Default:     settings.ProductionType,
					Options:     []string{settings.StagingType, settings.ProductionType, settings.SelfSignedType},
				},
				{
					Variable:    settings.CertManagerImageType,
					Description: "Choose which cert-manager image to use",
					Default:     settings.CertManagerImage.Get(),
				},
			},
			Answers: map[string]string{
				settings.RioWildcardType:      settings.StagingType,
				settings.PublicDomainType:     settings.ProductionType,
				settings.CertManagerImageType: settings.CertManagerImage.Get(),
			},
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
