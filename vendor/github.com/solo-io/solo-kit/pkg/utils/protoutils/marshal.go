// ilackarms: This file contains more than just proto-utils at this point. Should be split, or
// moved to a general serialization util package

package protoutils

import (
	"bytes"
	"encoding/json"

	v1 "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"

	"sigs.k8s.io/yaml"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

var jsonpbMarshaler = &jsonpb.Marshaler{OrigName: false}
var jsonpbMarshalerEmitZeroValues = &jsonpb.Marshaler{OrigName: false, EmitDefaults: true}

func UnmarshalBytes(data []byte, into resources.Resource) error {
	if protoInto, ok := into.(proto.Message); ok {
		return jsonpb.Unmarshal(bytes.NewBuffer(data), protoInto)
	}
	return json.Unmarshal(data, into)
}

func MarshalBytes(res resources.Resource) ([]byte, error) {
	if pb, ok := res.(proto.Message); ok {
		buf := &bytes.Buffer{}
		err := jsonpbMarshaler.Marshal(buf, pb)
		return buf.Bytes(), err
	}
	return json.Marshal(res)
}

func UnmarshalYAML(data []byte, into resources.Resource) error {
	jsn, err := yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}

	if protoInto, ok := into.(proto.Message); ok {
		return jsonpb.Unmarshal(bytes.NewBuffer(jsn), protoInto)
	}
	return json.Unmarshal(data, into)
}

func MarshalYAML(res resources.Resource) ([]byte, error) {
	var jsn []byte
	if pb, ok := res.(proto.Message); ok {
		buf := &bytes.Buffer{}
		if err := jsonpbMarshaler.Marshal(buf, pb); err != nil {
			return nil, err
		}
		jsn = buf.Bytes()
	} else {
		var err error
		jsn, err = json.Marshal(res)
		if err != nil {
			return nil, err
		}
	}
	return yaml.JSONToYAML(jsn)
}

func MarshalBytesEmitZeroValues(res resources.Resource) ([]byte, error) {
	if pb, ok := res.(proto.Message); ok {
		buf := &bytes.Buffer{}
		err := jsonpbMarshalerEmitZeroValues.Marshal(buf, pb)
		return buf.Bytes(), err
	}
	return json.Marshal(res)
}

func MarshalMap(from resources.Resource) (map[string]interface{}, error) {
	data, err := MarshalBytes(from)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	return m, err
}

func MarshalMapEmitZeroValues(from resources.Resource) (map[string]interface{}, error) {
	data, err := MarshalBytesEmitZeroValues(from)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	return m, err
}

func UnmarshalMap(m map[string]interface{}, into resources.Resource) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return UnmarshalBytes(data, into)
}

// ilackarms: help come up with a better name for this please
// values in stringMap are yaml encoded or error
// used by configmap resource client
func MapStringStringToMapStringInterface(stringMap map[string]string) (map[string]interface{}, error) {
	interfaceMap := make(map[string]interface{})
	for k, strVal := range stringMap {
		var interfaceVal interface{}
		if err := yaml.Unmarshal([]byte(strVal), &interfaceVal); err != nil {
			return nil, errors.Errorf("%v cannot be parsed as yaml", strVal)
		} else {
			interfaceMap[k] = interfaceVal
		}
	}
	return interfaceMap, nil
}

// reverse of previous
func MapStringInterfaceToMapStringString(interfaceMap map[string]interface{}) (map[string]string, error) {
	stringMap := make(map[string]string)
	for k, interfaceVal := range interfaceMap {
		yml, err := yaml.Marshal(interfaceVal)
		if err != nil {
			return nil, errors.Wrapf(err, "map values must be serializable to json")
		}
		stringMap[k] = string(yml)
	}
	return stringMap, nil
}

// convert raw Kube JSON to a Solo-Kit resource
func UnmarshalResource(kubeJson []byte, resource resources.Resource) error {
	var resourceCrd v1.Resource
	if err := json.Unmarshal(kubeJson, &resourceCrd); err != nil {
		return errors.Wrapf(err, "unmarshalling from raw json")
	}
	resource.SetMetadata(kubeutils.FromKubeMeta(resourceCrd.ObjectMeta))
	if withStatus, ok := resource.(resources.InputResource); ok {
		resources.UpdateStatus(withStatus, func(status *core.Status) {
			*status = resourceCrd.Status
		})
	}

	if resourceCrd.Spec != nil {
		if err := UnmarshalMap(*resourceCrd.Spec, resource); err != nil {
			return errors.Wrapf(err, "parsing resource from crd spec %v in namespace %v into %T", resourceCrd.Name, resourceCrd.Namespace, resource)
		}
	}

	return nil
}
