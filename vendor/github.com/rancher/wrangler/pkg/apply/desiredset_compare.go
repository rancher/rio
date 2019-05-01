package apply

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	LabelApplied = "objectset.rio.cattle.io/applied"
)

var (
	patchCache     = map[schema.GroupVersionKind]patchCacheEntry{}
	patchCacheLock = sync.Mutex{}
)

type patchCacheEntry struct {
	patchType types.PatchType
	lookup    strategicpatch.LookupPatchMeta
}

func prepareObjectForCreate(obj runtime.Object) (runtime.Object, error) {
	serialized, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	obj = obj.DeepCopyObject()
	meta, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}
	annotations := meta.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations[LabelApplied] = appliedToAnnotation(serialized)
	meta.SetAnnotations(annotations)

	return obj, nil
}

func originalAndModified(oldMetadata v1.Object, newObject runtime.Object) ([]byte, []byte, error) {
	original, err := getOriginal(oldMetadata)
	if err != nil {
		return nil, nil, err
	}

	newObject, err = prepareObjectForCreate(newObject)
	if err != nil {
		return nil, nil, err
	}

	modified, err := json.Marshal(newObject)

	return original, modified, err
}

func emptyMaps(data map[string]interface{}, keys ...string) bool {
	for _, key := range append(keys, "__invalid_key__") {
		if len(data) == 0 {
			// map is empty so all children are empty too
			return true
		} else if len(data) > 1 {
			// map has more than one key so not empty
			return false
		}

		value, ok := data[key]
		if !ok {
			// map has one key but not what we are expecting so not considered empty
			return false
		}

		data = toMapInterface(value)
	}

	return true
}

func sanitizePatch(patch []byte) ([]byte, error) {
	mod := false
	data := map[string]interface{}{}
	err := json.Unmarshal(patch, &data)
	if err != nil {
		return nil, err
	}

	if _, ok := data["kind"]; ok {
		mod = true
		delete(data, "kind")
	}

	if _, ok := data["apiVersion"]; ok {
		mod = true
		delete(data, "apiVersion")
	}

	if deleted := removeCreationTimestamp(data); deleted {
		mod = true
	}

	if emptyMaps(data, "metadata", "annotations") {
		return []byte("{}"), nil
	}

	if !mod {
		return patch, nil
	}

	return json.Marshal(data)
}

func applyPatch(gvk schema.GroupVersionKind, patcher Patcher, debugID string, oldObject, newObject runtime.Object) (bool, error) {
	oldMetadata, err := meta.Accessor(oldObject)
	if err != nil {
		return false, err
	}

	original, modified, err := originalAndModified(oldMetadata, newObject)
	if err != nil {
		return false, err
	}

	current, err := json.Marshal(oldObject)
	if err != nil {
		return false, err
	}

	patchType, patch, err := doPatch(gvk, original, modified, current)
	if err != nil {
		return false, errors.Wrap(err, "patch generation")
	}

	if string(patch) == "{}" {
		return false, nil
	}

	patch, err = sanitizePatch(patch)
	if err != nil {
		return false, err
	}

	if string(patch) == "{}" {
		return false, nil
	}

	logrus.Debugf("DesiredSet - Patch %s %s/%s for %s -- [%s, %s, %s, %s]", gvk, oldMetadata.GetNamespace(), oldMetadata.GetName(), debugID,
		patch, original, modified, current)

	logrus.Debugf("DesiredSet - Updated %s %s/%s for %s -- %s %s", gvk, oldMetadata.GetNamespace(), oldMetadata.GetName(), debugID, patchType, patch)
	_, err = patcher(oldMetadata.GetNamespace(), oldMetadata.GetName(), patchType, patch)

	return true, err
}

func (o *desiredSet) compareObjects(gvk schema.GroupVersionKind, patcher Patcher, client dynamic.NamespaceableResourceInterface, debugID string, oldObject, newObject runtime.Object, force bool) error {
	oldMetadata, err := meta.Accessor(oldObject)
	if err != nil {
		return err
	}

	if ran, err := applyPatch(gvk, patcher, debugID, oldObject, newObject); err != nil {
		return err
	} else if !ran {
		logrus.Debugf("DesiredSet - No change(2) %s %s/%s for %s", gvk, oldMetadata.GetNamespace(), oldMetadata.GetName(), debugID)
	}

	return nil
}

func removeCreationTimestamp(data map[string]interface{}) bool {
	metadata, ok := data["metadata"]
	if !ok {
		return false
	}

	data = toMapInterface(metadata)
	if _, ok := data["creationTimestamp"]; ok {
		delete(data, "creationTimestamp")
		return true
	}

	return false
}

func getOriginal(obj v1.Object) ([]byte, error) {
	original := appliedFromAnnotation(obj.GetAnnotations()[LabelApplied])
	if len(original) == 0 {
		return []byte("{}"), nil
	}

	mapObj := &unstructured.Unstructured{}
	err := json.Unmarshal(original, mapObj)
	if err != nil {
		return nil, err
	}

	removeCreationTimestamp(mapObj.Object)

	objCopy, err := prepareObjectForCreate(mapObj)
	if err != nil {
		return nil, err
	}

	return json.Marshal(objCopy)
}

func appliedFromAnnotation(str string) []byte {
	if len(str) == 0 || str[0] == '{' {
		return []byte(str)
	}

	b, err := base64.RawStdEncoding.DecodeString(str)
	if err != nil {
		return nil
	}

	r, err := gzip.NewReader(bytes.NewBuffer(b))
	if err != nil {
		return nil
	}

	b, err = ioutil.ReadAll(r)
	if err != nil {
		return nil
	}

	return b
}

func appliedToAnnotation(b []byte) string {
	if len(b) < 1024 {
		return string(b)
	}
	buf := &bytes.Buffer{}
	w := gzip.NewWriter(buf)
	if _, err := w.Write(b); err != nil {
		return string(b)
	}
	if err := w.Close(); err != nil {
		return string(b)
	}
	return base64.RawStdEncoding.EncodeToString(buf.Bytes())
}

// doPatch is adapted from "kubectl apply"
func doPatch(gvk schema.GroupVersionKind, original, modified, current []byte) (types.PatchType, []byte, error) {
	var patchType types.PatchType
	var patch []byte
	var lookupPatchMeta strategicpatch.LookupPatchMeta

	patchType, lookupPatchMeta, err := getPatchStyle(gvk)
	if err != nil {
		return patchType, nil, err
	}

	if patchType == types.StrategicMergePatchType {
		patch, err = strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, true)
	} else {
		patch, err = jsonmergepatch.CreateThreeWayJSONMergePatch(original, modified, current)
	}

	if err != nil {
		logrus.Errorf("Failed to calcuated patch: %v", err)
	}

	return patchType, patch, err
}

func getPatchStyle(gvk schema.GroupVersionKind) (types.PatchType, strategicpatch.LookupPatchMeta, error) {
	var (
		patchType       types.PatchType
		lookupPatchMeta strategicpatch.LookupPatchMeta
	)

	patchCacheLock.Lock()
	entry, ok := patchCache[gvk]
	patchCacheLock.Unlock()

	if ok {
		return entry.patchType, entry.lookup, nil
	}

	versionedObject, err := scheme.Scheme.New(gvk)

	if runtime.IsNotRegisteredError(err) {
		patchType = types.MergePatchType
	} else if err != nil {
		return patchType, nil, err
	} else {
		patchType = types.StrategicMergePatchType
		lookupPatchMeta, err = strategicpatch.NewPatchMetaFromStruct(versionedObject)
		if err != nil {
			return patchType, nil, err
		}
	}

	patchCacheLock.Lock()
	patchCache[gvk] = patchCacheEntry{
		patchType: patchType,
		lookup:    lookupPatchMeta,
	}
	patchCacheLock.Unlock()

	return patchType, lookupPatchMeta, nil
}

func toMapInterface(obj interface{}) map[string]interface{} {
	v, _ := obj.(map[string]interface{})
	return v
}
