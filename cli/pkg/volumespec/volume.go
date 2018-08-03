package volumespec

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/docker/docker/api/types/mount"
	"github.com/pkg/errors"
)

// Copied from github.com/docker/cli/opts a23c5d157b5265520ae41c133e988a85ac1b4606

const endOfSpec = rune(0)

type ServiceVolumeConfig struct {
	Type        string               `yaml:",omitempty"`
	Source      string               `yaml:",omitempty"`
	Target      string               `yaml:",omitempty"`
	ReadOnly    bool                 `mapstructure:"read_only" yaml:"read_only,omitempty"`
	Consistency string               `yaml:",omitempty"`
	Bind        *ServiceVolumeBind   `yaml:",omitempty"`
	Volume      *ServiceVolumeVolume `yaml:",omitempty"`
	Tmpfs       *ServiceVolumeTmpfs  `yaml:",omitempty"`
}

type ServiceVolumeBind struct {
	Propagation string `yaml:",omitempty"`
}

// ServiceVolumeVolume are options for a service volume of type volume
type ServiceVolumeVolume struct {
	NoCopy bool `mapstructure:"nocopy" yaml:"nocopy,omitempty"`
}

// ServiceVolumeTmpfs are options for a service volume of type tmpfs
type ServiceVolumeTmpfs struct {
	Size int64 `yaml:",omitempty"`
}

// ParseVolume parses a volume spec without any knowledge of the target platform
func ParseVolume(spec string) (ServiceVolumeConfig, error) {
	volume := ServiceVolumeConfig{}

	switch len(spec) {
	case 0:
		return volume, errors.New("invalid empty volume spec")
	case 1, 2:
		volume.Target = spec
		volume.Type = string(mount.TypeVolume)
		return volume, nil
	}

	buffer := []rune{}
	for _, char := range spec + string(endOfSpec) {
		switch {
		case isWindowsDrive(buffer, char):
			buffer = append(buffer, char)
		case char == ':' || char == endOfSpec:
			if err := populateFieldFromBuffer(char, buffer, &volume); err != nil {
				populateType(&volume)
				return volume, errors.Wrapf(err, "invalid spec: %s", spec)
			}
			buffer = []rune{}
		default:
			buffer = append(buffer, char)
		}
	}

	populateType(&volume)
	return volume, nil
}

func isWindowsDrive(buffer []rune, char rune) bool {
	return char == ':' && len(buffer) == 1 && unicode.IsLetter(buffer[0])
}

func populateFieldFromBuffer(char rune, buffer []rune, volume *ServiceVolumeConfig) error {
	strBuffer := string(buffer)
	switch {
	case len(buffer) == 0:
		return errors.New("empty section between colons")
	// Anonymous volume
	case volume.Source == "" && char == endOfSpec:
		volume.Target = strBuffer
		return nil
	case volume.Source == "":
		volume.Source = strBuffer
		return nil
	case volume.Target == "":
		volume.Target = strBuffer
		return nil
	case char == ':':
		return errors.New("too many colons")
	}
	for _, option := range strings.Split(strBuffer, ",") {
		switch option {
		case "ro":
			volume.ReadOnly = true
		case "rw":
			volume.ReadOnly = false
		case "nocopy":
			volume.Volume = &ServiceVolumeVolume{NoCopy: true}
		default:
			if isBindOption(option) {
				volume.Bind = &ServiceVolumeBind{Propagation: option}
			}
			// ignore unknown options
		}
	}
	return nil
}

func isBindOption(option string) bool {
	for _, propagation := range mount.Propagations {
		if mount.Propagation(option) == propagation {
			return true
		}
	}
	return false
}

func populateType(volume *ServiceVolumeConfig) {
	switch {
	// Anonymous volume
	case volume.Source == "":
		volume.Type = string(mount.TypeVolume)
	case isFilePath(volume.Source):
		volume.Type = string(mount.TypeBind)
	default:
		volume.Type = string(mount.TypeVolume)
	}
}

func isFilePath(source string) bool {
	switch source[0] {
	case '.', '/', '~':
		return true
	}

	// windows named pipes
	if strings.HasPrefix(source, `\\`) {
		return true
	}

	first, nextIndex := utf8.DecodeRuneInString(source)
	return isWindowsDrive([]rune{first}, rune(source[nextIndex]))
}
