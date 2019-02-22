//    Copyright 2013-2017 Docker, Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       https://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package create

import (
	"strings"

	"github.com/pkg/errors"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func ParseDevices(devices []string) ([]riov1.DeviceMapping, error) {
	var result []riov1.DeviceMapping
	for _, device := range devices {
		mapping, err := parseDevice(device)
		if err != nil {
			return nil, err
		}
		result = append(result, mapping)
	}

	return result, nil
}

func parseDevice(device string) (riov1.DeviceMapping, error) {
	src := ""
	dst := ""
	permissions := "rwm"
	arr := strings.Split(device, ":")
	switch len(arr) {
	case 3:
		permissions = arr[2]
		fallthrough
	case 2:
		if validDeviceMode(arr[1]) {
			permissions = arr[1]
		} else {
			dst = arr[1]
		}
		fallthrough
	case 1:
		src = arr[0]
	default:
		return riov1.DeviceMapping{}, errors.Errorf("invalid device specification: %s", device)
	}

	if dst == "" {
		dst = src
	}

	deviceMapping := riov1.DeviceMapping{
		OnHost:      src,
		InContainer: dst,
		Permissions: permissions,
	}
	return deviceMapping, nil
}

// validDeviceMode checks if the mode for device is valid or not.
// Valid mode is a composition of r (read), w (write), and m (mknod).
func validDeviceMode(mode string) bool {
	var legalDeviceMode = map[rune]bool{
		'r': true,
		'w': true,
		'm': true,
	}
	if mode == "" {
		return false
	}
	for _, c := range mode {
		if !legalDeviceMode[c] {
			return false
		}
		legalDeviceMode[c] = false
	}
	return true
}
