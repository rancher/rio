package test

import (
	"fmt"
	"testing"

	gvk2 "github.com/rancher/wrangler/pkg/gvk"
	"github.com/rancher/wrangler/pkg/objectset"
	"gotest.tools/assert"
	meta2 "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

func AssertObjects(t *testing.T, expected runtime.Object, os *objectset.ObjectSet) {
	o := GetObject(t, expected, os)
	assert.DeepEqual(t, expected, o)
}

func GetObject(t *testing.T, expected runtime.Object, os *objectset.ObjectSet) runtime.Object {
	gvk, err := gvk2.Get(expected)
	if err != nil {
		t.Fatal(err)
	}

	meta, err := meta2.Accessor(expected)
	if err != nil {
		t.Fatal(err)
	}

	objects := os.ObjectsByGVK()

	objectByName, ok := objects[gvk]
	assert.Assert(t, ok, fmt.Sprintf("gvk %s should exist in objectset", gvk.String()))
	objKey := objectset.ObjectKey{
		Name:      meta.GetName(),
		Namespace: meta.GetNamespace(),
	}
	o, ok := objectByName[objKey]
	assert.Assert(t, ok, fmt.Sprintf("object key %s should exist in objectset", objKey.String()))
	return o
}
