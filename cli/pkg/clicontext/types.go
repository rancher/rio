package clicontext

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"k8s.io/apimachinery/pkg/selection"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/types"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func init() {
	lookup.RegisterType(types.ConfigType, lookup.StackScopedNameType, lookup.SingleNameNameType)
	lookup.RegisterType(types.VolumeType, lookup.StackScopedNameType, lookup.SingleNameNameType)
	lookup.RegisterType(types.RouterType, lookup.StackScopedNameType, lookup.SingleNameNameType)
	lookup.RegisterType(types.ExternalServiceType, lookup.StackScopedNameType, lookup.SingleNameNameType)
	lookup.RegisterType(types.ServiceType, lookup.StackScopedNameType, lookup.SingleNameNameType, lookup.VersionedSingleNameNameType, lookup.VersionedStackScopedNameType)
	lookup.RegisterType(types.PodType, lookup.FourPartsNameType, lookup.ThreePartsNameType)
	lookup.RegisterType(types.StackType, lookup.SingleNameNameType)
	lookup.RegisterType(types.FeatureType, lookup.SingleNameNameType)
	lookup.RegisterType(types.PublicDomainType, lookup.FullDomainNameTypeNameType)

}

func (c *CLIContext) getResource(r types.Resource) (ret types.Resource, err error) {
	switch r.Type {
	case clitypes.ServiceType:
		r.Object, err = c.Rio.Services(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.StackType:
		r.Object, err = c.Rio.Stacks(c.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.PodType:
		r.Object, err = c.Core.Pods(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.ConfigType:
		r.Object, err = c.Rio.Configs(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.RouterType:
		r.Object, err = c.Rio.Services(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.VolumeType:
		r.Object, err = c.Rio.Volumes(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.ExternalServiceType:
		r.Object, err = c.Rio.ExternalServices(r.Namespace).Get(r.Name, metav1.GetOptions{})
	default:
		return r, fmt.Errorf("unknown by id type %s", r.Type)
	}

	return r, err
}

func (c *CLIContext) DeleteResource(r types.Resource) (err error) {
	switch r.Type {
	case clitypes.ServiceType:
		err = c.Rio.Services(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
	case clitypes.StackType:
		err = c.Rio.Stacks(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
	case clitypes.PodType:
		err = c.Core.Pods(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
	case clitypes.ConfigType:
		err = c.Rio.Configs(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
	case clitypes.RouterType:
		err = c.Rio.Services(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
	case clitypes.VolumeType:
		err = c.Rio.Volumes(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
	case clitypes.ExternalServiceType:
		err = c.Rio.ExternalServices(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
	default:
		return fmt.Errorf("unknown delete type %s", r.Type)
	}
	return
}

func (c *CLIContext) Create(obj runtime.Object) (err error) {
	metadata, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	if metadata.GetName() == "" {
		metadata.SetName(strings.Replace(namesgenerator.GetRandomName(0), "_", "-", -1))
	}

	switch o := obj.(type) {
	case *v1.Service:
		_, err = c.Rio.Services(o.Namespace).Create(o)
	case *v1.Stack:
		_, err = c.Rio.Stacks(o.Namespace).Create(o)
	case *corev1.Pod:
		_, err = c.Core.Pods(o.Namespace).Create(o)
	case *v1.Config:
		_, err = c.Rio.Configs(o.Namespace).Create(o)
	case *v1.Router:
		_, err = c.Rio.Routers(o.Namespace).Create(o)
	case *v1.Volume:
		_, err = c.Rio.Volumes(o.Namespace).Create(o)
	case *v1.ExternalService:
		_, err = c.Rio.ExternalServices(o.Namespace).Create(o)
	default:
		return fmt.Errorf("unknown delete type %v", reflect.TypeOf(obj))
	}
	return
}

func (c *CLIContext) UpdateObject(obj runtime.Object) (err error) {
	switch o := obj.(type) {
	case *v1.Service:
		_, err = c.Rio.Services(o.Namespace).Update(o)
	case *v1.Stack:
		_, err = c.Rio.Stacks(o.Namespace).Update(o)
	case *corev1.Pod:
		_, err = c.Core.Pods(o.Namespace).Update(o)
	case *v1.Config:
		_, err = c.Rio.Configs(o.Namespace).Update(o)
	case *v1.Router:
		_, err = c.Rio.Routers(o.Namespace).Update(o)
	case *v1.Volume:
		_, err = c.Rio.Volumes(o.Namespace).Update(o)
	case *v1.ExternalService:
		_, err = c.Rio.ExternalServices(o.Namespace).Update(o)
	case *projectv1.Feature:
		_, err = c.Project.Features(o.Namespace).Update(o)
	default:
		return fmt.Errorf("unknown delete type %v", reflect.TypeOf(obj))
	}
	return
}

func (c *CLIContext) List(typeName string) (ret []runtime.Object, err error) {
	switch typeName {
	case clitypes.StackType:
		return c.listNamespace(c.Namespace, typeName)
	case clitypes.FeatureType:
		return c.listFeatures()
	default:
		stacks, err := c.List(types.StackType)
		if err != nil {
			return nil, err
		}

		lock := sync.Mutex{}
		eg := errgroup.Group{}
		for _, stack := range stacks {
			meta, err := meta.Accessor(stack)
			if err != nil {
				return nil, err
			}
			eg.Go(func() error {
				obj, err := c.listNamespace(meta.GetNamespace(), typeName)
				if err != nil {
					return err
				}
				lock.Lock()
				ret = append(ret, obj...)
				lock.Unlock()
				return nil
			})
		}

		return ret, eg.Wait()
	}
}

func (c *CLIContext) listFeatures() (ret []runtime.Object, err error) {
	system, err := c.Project.Features("").List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"rio.cattle.io/system": "true",
		}).String(),
	})
	if err != nil {
		return nil, err
	}

	req, _ := labels.NewRequirement("rio.cattle.io/system", selection.NotIn, []string{"true"})
	nonSystem, err := c.Project.Features(c.Namespace).List(metav1.ListOptions{
		LabelSelector: labels.NewSelector().Add(*req).String(),
	})
	if err != nil {
		return nil, err
	}

	for _, obj := range append(nonSystem.Items, system.Items...) {
		copy := obj
		ret = append(ret, &copy)
	}

	return
}

func (c *CLIContext) listNamespace(namespace, typeName string) (ret []runtime.Object, err error) {
	projectOpts := metav1.ListOptions{}
	opts := metav1.ListOptions{}

	switch typeName {
	case clitypes.ServiceType:
		objs, err := c.Rio.Services(namespace).List(opts)
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.StackType:
		objs, err := c.Rio.Stacks(namespace).List(opts)
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.PodType:
		objs, err := c.Core.Pods(namespace).List(opts)
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.ConfigType:
		objs, err := c.Rio.Configs(namespace).List(opts)
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.RouterType:
		objs, err := c.Rio.Routers(namespace).List(opts)
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.VolumeType:
		objs, err := c.Rio.Volumes(namespace).List(opts)
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.ExternalServiceType:
		objs, err := c.Rio.ExternalServices(namespace).List(opts)
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.FeatureType:
		objs, err := c.Project.Features(c.Namespace).List(projectOpts)
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	default:
		return nil, fmt.Errorf("unknown list type %s", typeName)
	}
}
