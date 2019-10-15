package apply

import (
	"encoding/json"
	"fmt"
	"reflect"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	defaultReconcilers = map[schema.GroupVersionKind]Reconciler{
		v1.SchemeGroupVersion.WithKind("Service"):  reconcileService,
		batchv1.SchemeGroupVersion.WithKind("Job"): reconcileJob,
	}
)

func reconcileService(oldObj, newObj runtime.Object) (bool, error) {
	oldSvc, ok := oldObj.(*v1.Service)
	if !ok {
		oldSvc = &v1.Service{}
		if err := convertObj(oldObj, oldSvc); err != nil {
			return false, err
		}
	}
	newSvc, ok := newObj.(*v1.Service)
	if !ok {
		newSvc = &v1.Service{}
		if err := convertObj(newObj, newSvc); err != nil {
			return false, err
		}
	}

	if newSvc.Spec.Type != "" && newSvc.Spec.Type != newSvc.Spec.Type {
		return false, ErrReplace
	}

	return false, nil
}

func reconcileJob(oldObj, newObj runtime.Object) (bool, error) {
	oldSvc, ok := oldObj.(*batchv1.Job)
	if !ok {
		oldSvc = &batchv1.Job{}
		if err := convertObj(oldObj, oldSvc); err != nil {
			return false, err
		}
	}

	newSvc, ok := newObj.(*batchv1.Job)
	if !ok {
		newSvc = &batchv1.Job{}
		if err := convertObj(newObj, newSvc); err != nil {
			return false, err
		}
	}

	if !equality.Semantic.DeepEqual(oldSvc.Spec.Template, newSvc.Spec.Template) {
		return false, ErrReplace
	}

	return false, nil
}

func convertObj(src interface{}, obj interface{}) error {
	uObj, ok := src.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected unstructured but got %v", reflect.TypeOf(src))
	}

	bytes, err := uObj.MarshalJSON()
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, obj)
}
