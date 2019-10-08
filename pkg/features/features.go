package features

import (
	"context"
)

var (
	features = map[string]Feature{}
)

type Feature interface {
	Start(ctx context.Context) error
	Stop() error
	Spec() FeatureSpec
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
