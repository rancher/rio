package app

import (
	"testing"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"gotest.tools/assert"
)

func TestPopulateDestinationRule(t *testing.T) {
	input := riov1.NewApp("default", "foo", riov1.App{
		Spec: riov1.AppSpec{
			Revisions: []riov1.Revision{
				{
					Version: "v0",
				},
				{
					Version: "v1",
				},
			},
		},
	})

	result := destinationRuleForService(input)

	expected := constructors.NewDestinationRule("default", "foo", v1alpha3.DestinationRule{
		Spec: v1alpha3.DestinationRuleSpec{
			Host: "foo.default.svc.cluster.local",
			Subsets: []v1alpha3.Subset{
				{
					Name: "v0",
					Labels: map[string]string{
						"version": "v0",
					},
				},
				{
					Name: "v1",
					Labels: map[string]string{
						"version": "v1",
					},
				},
			},
		},
	})

	assert.DeepEqual(t, result, expected)
}
