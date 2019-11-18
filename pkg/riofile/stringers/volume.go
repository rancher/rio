package stringers

import (
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

var (
	hostPathTypes = map[string]string{
		"directoryorcreate": "DirectoryOrCreate",
		"directory":         "Directory",
		"fileorcreate":      "FileOrCreate",
		"file":              "File",
		"socket":            "Socket",
		"chardevice":        "CharDevice",
		"blockdevice":       "BlockDevice",
	}
)

func ParseVolumes(vols ...string) (result []v1.Volume, err error) {
	for _, vol := range vols {
		v, err := ParseVolume(vol)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return
}

func ParseVolume(v string) (volume v1.Volume, err error) {
	parts := strings.Split(v, ",")
	name, path := kv.Split(parts[0], ":")
	if path != "" {
		volume.Path = path
		if strings.HasPrefix(name, "/") {
			volume.HostPath = name
		} else {
			volume.Name = name
		}
	} else {
		volume.Path = name
	}

	for k, v := range kv.SplitMapFromSlice(parts[1:]) {
		k = strings.ToLower(k)
		switch k {
		case "persistent":
			value, _ := strconv.ParseBool(v)
			volume.Persistent = value
		case "hosttype":
			if volume.HostPath == "" && volume.Name != "" {
				volume.HostPath = volume.Name
				volume.Name = ""
			}

			hostPathType := hostPathTypes[strings.ToLower(v)]
			if hostPathType == "" {
				return volume, fmt.Errorf("invalid HostPathType %s", v)
			}
			hpt := corev1.HostPathType(hostPathType)
			volume.HostPathType = &hpt
		}
	}

	return volume, nil
}

type VolumeStringer struct {
	v1.Volume
}

func (v VolumeStringer) MaybeString() interface{} {
	buf := &strings.Builder{}
	if v.Name != "" {
		buf.WriteString(v.Name)
		buf.WriteString(":")
	} else if v.HostPath != "" {
		buf.WriteString(v.HostPath)
		buf.WriteString(":")
	}

	buf.WriteString(v.Path)

	if v.HostPathType != nil {
		buf.WriteString(fmt.Sprintf(",hostPathType=%s", *v.HostPathType))
	}

	if v.HostPathType == nil && v.HostPath != "" && !strings.HasPrefix(v.HostPath, "/") {
		buf.WriteString(",hostPathType=DirectoryOrCreate")
	}

	if v.Persistent {
		buf.WriteString(",persistent")
	}

	return buf.String()
}
