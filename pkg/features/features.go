package features

import (
	"context"

	"github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
)

var (
	features = map[string]Feature{}
)

type Feature interface {
	Start(ctx context.Context, feature *v1.Feature) error
	Changed(feature *v1.Feature) error
	Stop() error
	Spec() v1.FeatureSpec
	Name() string
}

func Register(feature Feature) {
	features[feature.Name()] = feature
}

func GetFeature(name string) Feature {
	return features[name]
}

func GetFeatures() []Feature {
	var result []Feature
	for _, f := range features {
		result = append(result, f)
	}

	return result
}
