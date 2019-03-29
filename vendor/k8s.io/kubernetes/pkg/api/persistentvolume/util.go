/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package persistentvolume

import (
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/features"
)

// DropDisabledFields removes disabled fields from the pv spec.
// This should be called from PrepareForCreate/PrepareForUpdate for all resources containing a pv spec.
func DropDisabledFields(pvSpec *api.PersistentVolumeSpec, oldPVSpec *api.PersistentVolumeSpec) {
	if !utilfeature.DefaultFeatureGate.Enabled(features.BlockVolume) {
		// TODO(liggitt): change this to only drop pvSpec.VolumeMode if (oldPVSpec == nil || oldPVSpec.VolumeMode == nil)
		// Requires more coordinated changes to validation
		pvSpec.VolumeMode = nil
		if oldPVSpec != nil {
			oldPVSpec.VolumeMode = nil
		}
	}

	if !utilfeature.DefaultFeatureGate.Enabled(features.CSIPersistentVolume) {
		// if this is a new PV, or the old PV didn't already have the CSI field, clear it
		if oldPVSpec == nil || oldPVSpec.PersistentVolumeSource.CSI == nil {
			pvSpec.PersistentVolumeSource.CSI = nil
		}
	}
}
