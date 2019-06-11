package stringers

import (
	"fmt"
	"path/filepath"
	"strings"

	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

type DataMountStringer struct {
	v1.DataMount
	defaultPrefix string
}

func (d DataMountStringer) MaybeString() interface{} {
	buf := &strings.Builder{}
	buf.WriteString(d.Name)
	if d.Key != "" {
		buf.WriteString("/")
		buf.WriteString(d.Key)
	}
	if d.Directory == d.defaultPrefix && d.Key == d.File {
		return buf.String()
	}

	buf.WriteString(":")
	buf.WriteString(d.Directory)
	if d.File != "" {
		buf.WriteString("/")
		buf.WriteString(d.File)
	}

	return buf.String()
}

func ParseDataMounts(defaultFolder string, mounts ...string) (result []v1.DataMount, err error) {
	for _, config := range mounts {
		mapping, err := parseConfig(defaultFolder, config)
		if err != nil {
			return nil, err
		}
		result = append(result, mapping)
	}

	return result, nil
}

func parseConfig(defaultFolder, mount string) (dataMount v1.DataMount, err error) {
	from, to := kv.Split(mount, ":")
	fromParts := strings.Split(from, "/")
	if len(fromParts) > 2 {
		return dataMount, fmt.Errorf("%s is invalid, src must be name or name/key", mount)
	}

	dataMount.Name = fromParts[0]
	if len(fromParts) == 2 {
		dataMount.Key = fromParts[1]
	}

	switch {
	case to == "":
		dataMount.Directory = defaultFolder
		dataMount.File = dataMount.Key
	case dataMount.Key == "":
		dataMount.Directory = to
	default:
		dataMount.Directory = filepath.Dir(to)
		dataMount.File = filepath.Base(to)
	}

	return
}
