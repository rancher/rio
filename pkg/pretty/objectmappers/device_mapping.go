package objectmappers

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/mapper/mappers"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func NewDeviceMapping(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &DeviceMappingStringer{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			objs, err := ParseDevices(str)
			if err != nil {
				return nil, err
			}
			return objs[0], nil
		},
	}
}

type DeviceMappingStringer struct {
	v1.DeviceMapping
}

func (d DeviceMappingStringer) MaybeString() interface{} {
	result := d.OnHost
	if len(d.InContainer) > 0 {
		if len(result) > 0 {
			result += ":"
		}
		result += d.InContainer
	}
	if len(d.Permissions) > 0 {
		if len(result) > 0 {
			result += ":"
		}
		result += d.Permissions
	}

	return result
}

func ParseDevices(devices ...string) ([]riov1.DeviceMapping, error) {
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
