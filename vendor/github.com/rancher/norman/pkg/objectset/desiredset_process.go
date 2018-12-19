package objectset

import (
	"fmt"
	"sort"

	errors2 "k8s.io/apimachinery/pkg/api/errors"

	"github.com/pkg/errors"
	"github.com/rancher/norman/types"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

var (
	deletePolicy = v1.DeletePropagationBackground
)

func (o *DesiredSet) process(inputID, debugID string, set labels.Selector, gvk schema.GroupVersionKind, objs map[objectKey]runtime.Object) {
	client, ok := o.clients[gvk]
	if !ok {
		o.err(fmt.Errorf("failed to find client for %s for %s", gvk, debugID))
		return
	}

	indexer := client.Generic().Informer().GetIndexer()

	existing, err := list(indexer, set)
	if err != nil {
		o.err(fmt.Errorf("failed to list %s for %s", gvk, debugID))
		return
	}

	toCreate, toDelete, toUpdate := compareSets(existing, objs)
	for _, k := range toCreate {
		obj := objs[k]
		obj, err := prepareObjectForCreate(inputID, obj)
		if err != nil {
			o.err(errors.Wrapf(err, "failed to prepare create %s %s for %s", k, gvk, debugID))
			continue
		}

		_, err = client.ObjectClient().Create(obj)
		if errors2.IsAlreadyExists(err) {
			// Taking over an object that wasn't previously managed by us
			existingObj, err := client.ObjectClient().GetNamespaced(k.namespace, k.name, v1.GetOptions{})
			if err == nil {
				toUpdate = append(toUpdate, k)
				existing[k] = existingObj
				continue
			}
		}
		if err != nil {
			o.err(errors.Wrapf(err, "failed to create %s %s for %s", k, gvk, debugID))
			continue
		}
		logrus.Infof("DesiredSet - Created %s %s for %s", gvk, k, debugID)
	}

	for _, k := range toUpdate {
		err := o.compareObjects(client.ObjectClient(), debugID, inputID, existing[k], objs[k], len(toCreate) > 0 || len(toDelete) > 0)
		if err != nil {
			o.err(errors.Wrapf(err, "failed to update %s %s for %s", k, gvk, debugID))
			continue
		}
	}

	for _, k := range toDelete {
		err := client.ObjectClient().DeleteNamespaced(k.namespace, k.name, &v1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		})
		if err != nil {
			o.err(errors.Wrapf(err, "failed to delete %s %s for %s", k, gvk, debugID))
			continue
		}
		logrus.Infof("DesiredSet - Delete %s %s for %s", gvk, k, debugID)
	}
}

func compareSets(existingSet, newSet map[objectKey]runtime.Object) (toCreate, toDelete, toUpdate []objectKey) {
	for k := range newSet {
		if _, ok := existingSet[k]; ok {
			toUpdate = append(toUpdate, k)
		} else {
			toCreate = append(toCreate, k)
		}
	}

	for k := range existingSet {
		if _, ok := newSet[k]; !ok {
			toDelete = append(toDelete, k)
		}
	}

	sortObjectKeys(toCreate)
	sortObjectKeys(toDelete)
	sortObjectKeys(toUpdate)

	return
}

func sortObjectKeys(keys []objectKey) {
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})
}

func list(indexer cache.Indexer, selector labels.Selector) (map[objectKey]runtime.Object, error) {
	var (
		errs []error
		objs = map[objectKey]runtime.Object{}
	)

	err := cache.ListAllByNamespace(indexer, "", selector, func(obj interface{}) {
		metadata, err := meta.Accessor(obj)
		if err != nil {
			errs = append(errs, err)
			return
		}

		objs[objectKey{
			namespace: metadata.GetNamespace(),
			name:      metadata.GetName(),
		}] = obj.(runtime.Object)
	})
	if err != nil {
		errs = append(errs, err)
	}

	return objs, types.NewErrors(errs...)
}
