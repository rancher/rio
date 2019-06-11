package clicontext

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/types"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	projectv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/rancher/wrangler/pkg/merr"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
)

func init() {
	lookup.RegisterType(types.ConfigType, lookup.NamespaceScopedNameType, lookup.SingleNameNameType)
	lookup.RegisterType(types.RouterType, lookup.NamespaceScopedNameType, lookup.SingleNameNameType)
	lookup.RegisterType(types.ExternalServiceType, lookup.NamespaceScopedNameType, lookup.SingleNameNameType)
	lookup.RegisterType(types.AppType, lookup.NamespaceScopedNameType, lookup.SingleNameNameType)
	lookup.RegisterType(types.ServiceType, lookup.NamespaceScopedNameType, lookup.SingleNameNameType, lookup.VersionedSingleNameNameType, lookup.VersionedStackScopedNameType)
	lookup.RegisterType(types.PodType, lookup.NamespaceScopedNameType, lookup.ThreePartsNameType)
	lookup.RegisterType(types.NamespaceType, lookup.SingleNameNameType)
	lookup.RegisterType(types.FeatureType, lookup.SingleNameNameType)
	lookup.RegisterType(types.PublicDomainType, lookup.NamespaceScopedNameType, lookup.FullDomainNameTypeNameType)
	lookup.RegisterType(types.BuildType, lookup.NamespaceScopedNameType)
	lookup.RegisterType(types.SecretType, lookup.NamespacedSecretNameType)
}

func (c *CLIContext) getResource(r types.Resource) (ret types.Resource, err error) {
	switch r.Type {
	case clitypes.ServiceType:
		if strings.Contains(r.Name, ":") {
			appName, version := kv.Split(r.Name, ":")
			app, err := c.Rio.Apps(r.Namespace).Get(appName, metav1.GetOptions{})
			if err != nil {
				return ret, err
			}
			for _, rev := range app.Spec.Revisions {
				if rev.Version == version {
					var svc *riov1.Service
					svc, err = c.Rio.Services(r.Namespace).Get(rev.ServiceName, metav1.GetOptions{})
					if err != nil {
						return ret, err
					}
					svc.APIVersion = riov1.SchemeGroupVersion.String()
					svc.Kind = "Service"
					r.Object = svc
					r.Name = rev.ServiceName
					r.FullType = clitypes.ServiceTypeFull
				}
			}
		} else {
			var svc *riov1.Service
			svc, err = c.Rio.Services(r.Namespace).Get(r.Name, metav1.GetOptions{})
			svc.APIVersion = riov1.SchemeGroupVersion.String()
			svc.Kind = "Service"
			r.Object = svc
			r.FullType = clitypes.ServiceTypeFull
		}
	case clitypes.AppType:
		r.Object, err = c.Rio.Apps(r.Namespace).Get(r.Name, metav1.GetOptions{})
		r.FullType = clitypes.AppTypeFull
	case clitypes.PodType:
		podName, containerName := kv.Split(r.Name, "/")
		pod, err := c.Core.Pods(r.Namespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return r, err
		}
		if containerName != "" {
			for _, container := range pod.Spec.Containers {
				if container.Name == containerName {
					pod.Spec.Containers = []corev1.Container{
						container,
					}
					pod.Spec.InitContainers = nil
					break
				}
			}
			for _, container := range pod.Spec.InitContainers {
				if container.Name == containerName {
					pod.Spec.InitContainers = []corev1.Container{
						container,
					}
					pod.Spec.Containers = nil
					break
				}
			}
		}
		r.Object = pod
	case clitypes.ConfigType:
		r.Object, err = c.Core.ConfigMaps(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.RouterType:
		r.Object, err = c.Rio.Routers(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.ExternalServiceType:
		r.Object, err = c.Rio.ExternalServices(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.PublicDomainType:
		r.Object, err = c.Project.PublicDomains(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.NamespaceType:
		r.Object, err = c.Core.Namespaces().Get(r.Name, metav1.GetOptions{})
	case clitypes.FeatureType:
		r.Object, err = c.Project.Features(c.SystemNamespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.BuildType:
		r.Object, err = c.Build.Builds(r.Namespace).Get(r.Name, metav1.GetOptions{})
	case clitypes.SecretType:
		r.Object, err = c.Core.Secrets(r.Namespace).Get(r.Name, metav1.GetOptions{})
	default:
		return r, fmt.Errorf("unknown by id type %s", r.Type)
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
		err = c.Project.PublicDomains(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
	case clitypes.BuildType:
		err = c.Build.Builds(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
	case clitypes.AppType:
		app := r.Object.(*riov1.App)
		var errs []error
		for _, rev := range app.Spec.Revisions {
			newerr := c.Rio.Services(r.Namespace).Delete(rev.ServiceName, &metav1.DeleteOptions{})
			if newerr != nil && !errors.IsNotFound(err) {
				errs = append(errs, newerr)
			}
		}
		newerr := c.Rio.Apps(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
		if newerr != nil {
			errs = append(errs, newerr)
		}
		err = merr.NewErrors(errs...)
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
		if err != nil {
			return err
		}
		app, version := services.AppAndVersion(o)
		fmt.Printf("%s/%s:%s\n", o.Namespace, app, version)
		return nil
	case *corev1.Pod:
		_, err = c.Core.Pods(o.Namespace).Create(o)
	case *corev1.ConfigMap:
		_, err = c.Core.ConfigMaps(o.Namespace).Create(o)
	case *riov1.Router:
		_, err = c.Rio.Routers(o.Namespace).Create(o)
	case *riov1.ExternalService:
		_, err = c.Rio.ExternalServices(o.Namespace).Create(o)
	case *projectv1.PublicDomain:
		_, err = c.Project.PublicDomains(o.Namespace).Create(o)
	default:
		return fmt.Errorf("unknown delete type %v", reflect.TypeOf(obj))
	}
	if err != nil {
		return err
	}
	fmt.Printf("%s/%s\n", metadata.GetNamespace(), metadata.GetName())
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
	case *projectv1.Feature:
		_, err = c.Project.Features(o.Namespace).Update(o)
	case *projectv1.PublicDomain:
		_, err = c.Project.PublicDomains(o.Namespace).Update(o)
	default:
		return fmt.Errorf("unknown delete type %v", reflect.TypeOf(obj))
	}
	return
}

func (c *CLIContext) List(typeName string) (ret []runtime.Object, err error) {
	switch typeName {
	case clitypes.NamespaceType:
		return c.listNamespace(c.SystemNamespace, typeName)
	case clitypes.FeatureType:
		return c.listFeatures()
	default:
		namespaces, err := c.List(types.NamespaceType)
		if err != nil {
			return nil, err
		}

		lock := sync.Mutex{}
		eg := errgroup.Group{}
		for _, ns := range namespaces {
			meta, err := meta.Accessor(ns)
			if err != nil {
				return nil, err
			}
			if c.CLI.GlobalString("namespace") != "" && meta.GetName() != c.CLI.GlobalString("namespace") {
				continue
			}
			if !c.ShowSystem && meta.GetName() == c.SystemNamespace {
				continue
			}
			if c.ShowSystem && meta.GetName() != c.SystemNamespace {
				continue
			}
			eg.Go(func() error {
				obj, err := c.listNamespace(meta.GetName(), typeName)
				if err != nil {
					return err
				}
				lock.Lock()
				ret = append(ret, obj...)
				lock.Unlock()
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			return ret, err
		}
		sort.Slice(ret, func(i, j int) bool {
			meta1, _ := meta.Accessor(ret[i])
			meta2, _ := meta.Accessor(ret[j])
			return meta1.GetNamespace()+"/"+meta1.GetName() < meta2.GetNamespace()+"/"+meta2.GetName()
		})

		return ret, nil
	}
}

func (c *CLIContext) listFeatures() (ret []runtime.Object, err error) {
	system, err := c.Project.Features(c.SystemNamespace).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"rio.cattle.io/system": "true",
		}).String(),
	})
	if err != nil {
		return nil, err
	}

	req, _ := labels.NewRequirement("rio.cattle.io/system", selection.NotIn, []string{"true"})
	nonSystem, err := c.Project.Features(c.SystemNamespace).List(metav1.ListOptions{
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
	case clitypes.NamespaceType:
		objs, err := c.Core.Namespaces().List(opts)
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.AppType:
		objs, err := c.Rio.Apps(namespace).List(opts)
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.ServiceType:
		objs, err := c.Rio.Services(namespace).List(opts)
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
		objs, err := c.Core.ConfigMaps(namespace).List(opts)
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
	case clitypes.ExternalServiceType:
		objs, err := c.Rio.ExternalServices(namespace).List(opts)
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.FeatureType:
		objs, err := c.Project.Features(c.SystemNamespace).List(projectOpts)
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.PublicDomainType:
		objs, err := c.Project.PublicDomains(namespace).List(opts)
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.SecretType:
		objs, err := c.Core.Secrets(namespace).List(opts)
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	case clitypes.BuildType:
		objs, err := c.Build.Builds(namespace).List(opts)
		for i := range objs.Items {
			ret = append(ret, &objs.Items[i])
		}
		return ret, err
	default:
		return nil, fmt.Errorf("unknown list type %s", typeName)
	}
}
