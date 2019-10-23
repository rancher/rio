// Copyright 2018 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package cache

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
)

type TypedResources map[string]Resources

type GenericSnapshot struct {
	typedResources TypedResources
}

// Combine snapshots with distinct types to one.
func (s *GenericSnapshot) Combine(a *GenericSnapshot) (*GenericSnapshot, error) {
	if s.typedResources == nil {
		return a, nil
	} else if a.typedResources == nil {
		return s, nil
	}
	combined := TypedResources{}
	for k, v := range s.typedResources {
		combined[k] = v
	}
	for k, v := range a.typedResources {
		if _, ok := combined[k]; ok {
			return nil, errors.New("overlapping types found")
		}
		combined[k] = v
	}
	return NewGenericSnapshot(combined), nil
}

// Combine snapshots with distinct types to one.
func (s *GenericSnapshot) Merge(newSnap *GenericSnapshot) (*GenericSnapshot, error) {
	if s.typedResources == nil {
		return newSnap, nil
	}
	combined := TypedResources{}
	for k, v := range s.typedResources {
		combined[k] = v
	}
	for k, v := range newSnap.typedResources {
		combined[k] = v
	}
	return NewGenericSnapshot(combined), nil
}

// NewSnapshot creates a snapshot from response types and a version.
func NewGenericSnapshot(resources TypedResources) *GenericSnapshot {
	return &GenericSnapshot{
		typedResources: resources,
	}
}
func NewEasyGenericSnapshot(version string, resources ...[]Resource) *GenericSnapshot {
	t := TypedResources{}

	for _, resources := range resources {
		for _, resource := range resources {
			r := t[resource.Self().Type]
			if r.Items == nil {
				r.Items = make(map[string]Resource)
				r.Version = version
			}
			r.Items[resource.Self().Name] = resource
			t[resource.Self().Type] = r
		}
	}

	return &GenericSnapshot{
		typedResources: t,
	}
}

func (s *GenericSnapshot) Consistent() error {
	if s == nil {
		return nil
	}

	var required []XdsResourceReference

	for _, resources := range s.typedResources {
		for _, resource := range resources.Items {
			required = append(required, resource.References()...)
		}
	}

	for _, ref := range required {
		if resources, ok := s.typedResources[ref.Type]; ok {
			if _, ok := resources.Items[ref.Name]; ok {
				return fmt.Errorf("required resource not in snapshot: %s %s", ref.Type, ref.Name)
			}
		} else {
			return fmt.Errorf("required resource not in snapshot: %s %s", ref.Type, ref.Name)
		}
	}

	return nil
}

// GetResources selects snapshot resources by type.
func (s *GenericSnapshot) GetResources(typ string) Resources {
	if s == nil {
		return Resources{}
	}

	return s.typedResources[typ]
}

func (s *GenericSnapshot) Clone() Snapshot {
	typedResourcesCopy := make(TypedResources)
	for typeName, resources := range s.typedResources {
		resourcesCopy := Resources{
			Version: resources.Version,
			Items:   make(map[string]Resource, len(resources.Items)),
		}
		for k, v := range resources.Items {
			resourcesCopy.Items[k] = proto.Clone(v.ResourceProto()).(Resource)
		}
		typedResourcesCopy[typeName] = resourcesCopy
	}
	return &GenericSnapshot{typedResources: typedResourcesCopy}
}
