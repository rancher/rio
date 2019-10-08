package stringers

import (
	"fmt"
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
	if d.Target == d.defaultPrefix || d.Target == "" {
		return buf.String()
	}

	buf.WriteString(":")
	buf.WriteString(d.Target)

	return buf.String()
}

func ParseDataMount(mount string) (v1.DataMount, error) {
	return parseConfig(mount)
}

func parseConfig(mount string) (dataMount v1.DataMount, err error) {
	from, to := kv.Split(mount, ":")
	fromParts := strings.Split(from, "/")
	if len(fromParts) > 2 {
		return dataMount, fmt.Errorf("%s is invalid, src must be name or name/key", mount)
	}

	dataMount.Name = fromParts[0]
	if len(fromParts) == 2 {
		dataMount.Key = fromParts[1]
	}

	dataMount.Target = to
	return
}
