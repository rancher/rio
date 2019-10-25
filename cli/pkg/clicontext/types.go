package clicontext

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/rancher/rio/cli/pkg/types"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	projectv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/gvk"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (c *CLIContext) getResource(r types.Resource) (ret types.Resource, err error) {
	switch r.Type {
	case clitypes.ServiceType:
		r.Object, err = c.Rio.Services(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.ConfigType:
		r.Object, err = c.Core.ConfigMaps(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.RouterType:
		r.Object, err = c.Rio.Routers(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.ExternalServiceType:
		r.Object, err = c.Rio.ExternalServices(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.PublicDomainType:
		r.Object, err = c.Project.PublicDomains().Get(r.Name, metav1.GetOptions{})
	case clitypes.NamespaceType:
		r.Object, err = c.Core.Namespaces().Get(r.Name, metav1.GetOptions{})
	case clitypes.BuildType:
		r.Object, err = c.Build.TaskRuns(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.SecretType:
		r.Object, err = c.Core.Secrets(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.StackType:
		r.Object, err = c.Rio.Stacks(r.Namespace).Get(r.Name, metav1.GetOptions{})
	default:
		return r, fmt.Errorf("unknown by id type %s", r.Type)
	}
	if err == nil && r.Object != nil {
		return r, gvk.Set(r.Object)
	}

	return r, err
}

func (c *CLIContext) DeleteResource(r types.Resource) (err error) {
	switch r.Type {
	case clitypes.ServiceType:
		err = c.Rio.Services(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
	case clitypes.PodType:
		err = c.Core.Pods(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
	case clitypes.ConfigType:
		err = c.Core.ConfigMaps(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
	case clitypes.RouterType:
		err = c.Rio.Routers(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
	case clitypes.ExternalServiceType:
		err = c.Rio.ExternalServices(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
	case clitypes.PublicDomainType:
		err = c.Project.PublicDomains().Delete(r.Name, &metav1.DeleteOptions{})
	case clitypes.BuildType:
		err = c.Build.TaskRuns(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
	case clitypes.StackType:
		err = c.Rio.Stacks(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
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

	rand.Seed(time.Now().UnixNano())
	if metadata.GetName() == "" {
		metadata.SetName(strings.Replace(namesgenerator.GetRandomName(2), "_", "-", -1))
	}

	_, err = c.Core.Namespaces().Get(metadata.GetNamespace(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			if _, err := c.Core.Namespaces().Create(constructors.NewNamespace(metadata.GetNamespace(), corev1.Namespace{})); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	switch o := obj.(type) {
	case *riov1.Service:
		_, err = c.Rio.Services(o.Namespace).Create(o)
	case *corev1.Pod:
		_, err = c.Core.Pods(o.Namespace).Create(o)
	case *corev1.ConfigMap:
		_, err = c.Core.ConfigMaps(o.Namespace).Create(o)
	case *riov1.Router:
		_, err = c.Rio.Routers(o.Namespace).Create(o)
	case *riov1.ExternalService:
		_, err = c.Rio.ExternalServices(o.Namespace).Create(o)
	case *projectv1.PublicDomain:
		_, err = c.Project.PublicDomains().Create(o)
	case *riov1.Stack:
		_, err = c.Rio.Stacks(o.Namespace).Create(o)
	default:
		return fmt.Errorf("unknown type %v", reflect.TypeOf(obj))
	}
	if err != nil {
		return err
	}
	g, err := gvk.Get(obj)
	if err != nil {
		return err
	}
	fmt.Printf("%s/%s/%s\n", strings.ToLower(g.Kind), metadata.GetNamespace(), metadata.GetName())
	return nil
}

func (c *CLIContext) UpdateObject(obj runtime.Object) (err error) {
	switch o := obj.(type) {
	case *riov1.Service:
		_, err = c.Rio.Services(o.Namespace).Update(o)
	case *corev1.Pod:
		_, err = c.Core.Pods(o.Namespace).Update(o)
	case *corev1.ConfigMap:
		_, err = c.Core.ConfigMaps(o.Namespace).Update(o)
	case *riov1.Router:
		_, err = c.Rio.Routers(o.Namespace).Update(o)
	case *riov1.ExternalService:
		_, err = c.Rio.ExternalServices(o.Namespace).Update(o)
	case *projectv1.PublicDomain:
		_, err = c.Project.PublicDomains().Update(o)
	case *riov1.Stack:
		_, err = c.Rio.Stacks(o.Namespace).Update(o)
	default:
		return fmt.Errorf("unknown type %v", reflect.TypeOf(obj))
	}
	return
}

func (c *CLIContext) List(typeName string) (ret []runtime.Object, err error) {
	switch typeName {
	case clitypes.NamespaceType:
		return c.listNamespace(c.SystemNamespace, typeName)
	default:
		obj, err := c.listNamespace(c.GetSetNamespace(), typeName)
		if err != nil {
			return ret, err
		}
		for _, o := range obj {
			ret = append(ret, o)
		}

		return ret, nil
	}
}

func (c *CLIContext) listNamespace(namespace, typeName string) (ret []runtime.Object, err error) {
	opts := metav1.ListOptions{}

	switch typeName {
	case clitypes.NamespaceType:
		objs, err := c.Core.Namespaces().List(opts)
		if err != nil {
			return ret, err
		}
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.ServiceType:
		objs, err := c.Rio.Services(namespace).List(opts)
		if err != nil {
			return ret, err
		}
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.PodType:
		objs, err := c.Core.Pods(namespace).List(opts)
		if err != nil {
			return ret, err
		}
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.ConfigType:
		objs, err := c.Core.ConfigMaps(namespace).List(opts)
		if err != nil {
			return ret, err
		}
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.RouterType:
		objs, err := c.Rio.Routers(namespace).List(opts)
		if err != nil {
			return ret, err
		}
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.ExternalServiceType:
		objs, err := c.Rio.ExternalServices(namespace).List(opts)
		if err != nil {
			return ret, err
		}
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.PublicDomainType:
		objs, err := c.Project.PublicDomains().List(opts)
		if err != nil {
			return ret, err
		}
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.SecretType:
		opts = metav1.ListOptions{
			FieldSelector: "type!=kubernetes.io/service-account-token,type!=istio.io/key-and-cert",
		}
		objs, err := c.Core.Secrets(namespace).List(opts)
		if err != nil {
			return ret, err
		}
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.BuildType:
		objs, err := c.Build.TaskRuns(namespace).List(opts)
		if err != nil {
			return ret, err
		}
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.StackType:
		objs, err := c.Rio.Stacks(namespace).List(opts)
		if err != nil {
			return ret, err
		}
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	default:
		return nil, fmt.Errorf("unknown list type %s", typeName)
	}
}
