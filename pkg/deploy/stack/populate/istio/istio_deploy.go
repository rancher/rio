package istio

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/pkg/settings"
	"istio.io/istio/pilot/pkg/model"
)

func Populate(stack *input.Stack, output *output.Deployment) error {
	if stack.Stack != nil && stack.Stack.Spec.DisableMesh {
		return nil
	}

	if settings.ClusterDomain.Get() == "" {
		return fmt.Errorf("waiting for cluster domain")
	}

	vs, err := virtualservices(stack)
	if err != nil || len(vs) == 0 {
		return err
	}

	ds, err := destinationRules(stack)
	if err != nil {
		return err
	}

	ds, err = convertIstioObjects(ds)
	if err != nil {
		return err
	}

	for _, d := range ds {
		output.DestinationRules[d.Name] = d
	}

	vs, err = convertIstioObjects(vs)
	if err != nil {
		return err
	}

	for _, v := range vs {
		output.VirtualServices[v.Name] = v
	}

	return nil
}

func convertIstioObjects(objs []*output.IstioObject) ([]*output.IstioObject, error) {
	for _, obj := range objs {
		if pb, ok := obj.Spec.(proto.Message); ok {
			m, err := model.ToJSONMap(pb)
			if err != nil {
				return nil, err
			}
			obj.Spec = m
		}
	}

	return objs, nil
}
