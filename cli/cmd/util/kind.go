package util

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/pkg/types"
	gvk2 "github.com/rancher/wrangler/pkg/gvk"
	"github.com/rancher/wrangler/pkg/kv"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

func GetID(object runtime.Object, allNamespace bool) (string, error) {
	gvk, err := gvk2.Get(object)
	if err != nil {
		return "", err
	}
	metaObj, err := meta.Accessor(object)
	if err != nil {
		return "", err
	}
	kind := strings.ToLower(gvk.Kind)

	id := ""
	if allNamespace {
		id = fmt.Sprintf("%s/%s/%s", strings.ToLower(gvk.Kind), metaObj.GetNamespace(), metaObj.GetName())
	} else {
		id = fmt.Sprintf("%s/%s", strings.ToLower(gvk.Kind), metaObj.GetName())
	}

	if kind == types.ServiceType {
		_, id = kv.Split(id, "/")
	}
	return id, nil
}
