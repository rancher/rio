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

// Resources is a versioned group of resources.
type Resources struct {
	// Version information.
	Version string

	// Items in the group.
	Items map[string]Resource
}

// IndexResourcesByName creates a map from the resource name to the resource.
func IndexResourcesByName(items []Resource) map[string]Resource {
	indexed := make(map[string]Resource, len(items))
	for _, item := range items {
		indexed[item.Self().Name] = item
	}
	return indexed
}

// NewResources creates a new resource group.
func NewResources(version string, items []Resource) Resources {
	return Resources{
		Version: version,
		Items:   IndexResourcesByName(items),
	}
}

type Snapshot interface {
	Consistent() error
	GetResources(typ string) Resources
	Clone() Snapshot
}

type NilSnapshot struct{}

func (NilSnapshot) Consistent() error                 { return nil }
func (NilSnapshot) GetResources(typ string) Resources { return Resources{} }
func (NilSnapshot) Clone() Snapshot                   { return NilSnapshot{} }

var _ Snapshot = NilSnapshot{}
