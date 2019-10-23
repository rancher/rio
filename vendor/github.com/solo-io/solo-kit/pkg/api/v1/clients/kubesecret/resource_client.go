package kubesecret

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/common"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type ResourceClient struct {
	cache cache.KubeCoreCache
	Kube  kubernetes.Interface
	common.KubeCoreResourceClient
	ownerLabel   string
	resourceName string
	// should we marshal/unmarshal these secrets assuming their structure is map[string]string ?
	// custom logic to convert the secret to a resource
	secretConverter SecretConverter
}

const annotationKey = "resource_kind"

var (
	NotOurResource = fmt.Errorf("not kube secret resource")
)

func (rc *ResourceClient) FromKubeSecret(secret *v1.Secret) (resources.Resource, error) {
	resource := rc.NewResource()
	// not our secret
	// should be an error on a Read, ignored on a list
	if len(secret.ObjectMeta.Annotations) == 0 || secret.ObjectMeta.Annotations[annotationKey] != rc.Kind() {
		return nil, NotOurResource
	}
	// convert mapstruct to our object
	resourceMap, err := protoutils.MapStringStringToMapStringInterface(toStringStringMap(secret.Data))
	if err != nil {
		return nil, errors.Wrapf(err, "parsing secret data as map[string]interface{}")
	}

	if err := protoutils.UnmarshalMap(resourceMap, resource); err != nil {
		return nil, errors.Wrapf(err, "reading secret data into %v", rc.Kind())
	}
	resource.SetMetadata(kubeutils.FromKubeMeta(secret.ObjectMeta))
	return resource, nil
}

func toStringStringMap(input map[string][]byte) map[string]string {
	output := make(map[string]string)
	for k, v := range input {
		output[k] = string(v)
	}
	return output
}

func fromStringStringMap(input map[string]string) map[string][]byte {
	output := make(map[string][]byte)
	for k, v := range input {
		output[k] = []byte(v)
	}
	return output
}

func (rc *ResourceClient) ToKubeSecret(ctx context.Context, resource resources.Resource) (*v1.Secret, error) {
	resourceMap, err := protoutils.MarshalMap(resource)
	if err != nil {
		return nil, errors.Wrapf(err, "marshalling resource as map")
	}
	resourceData, err := protoutils.MapStringInterfaceToMapStringString(resourceMap)
	if err != nil {
		return nil, errors.Wrapf(err, "internal err: converting resource map to map[string]string")
	}
	// metadata moves over to kube style
	delete(resourceData, "metadata")

	meta := kubeutils.ToKubeMeta(resource.GetMetadata())
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	meta.Annotations[annotationKey] = rc.Kind()
	return &v1.Secret{
		ObjectMeta: meta,
		Data:       fromStringStringMap(resourceData),
	}, nil
}

func NewResourceClient(kube kubernetes.Interface, resourceType resources.Resource, plainSecrets bool, kubeCache cache.KubeCoreCache) (*ResourceClient, error) {
	var sc SecretConverter
	if plainSecrets {
		sc = new(plainSecret)
	}
	return NewResourceClientWithSecretConverter(kube, resourceType, kubeCache, sc)
}

func NewResourceClientWithSecretConverter(kube kubernetes.Interface, resourceType resources.Resource, kubeCache cache.KubeCoreCache, sc SecretConverter) (*ResourceClient, error) {
	return &ResourceClient{
		cache: kubeCache,
		Kube:  kube,
		KubeCoreResourceClient: common.KubeCoreResourceClient{
			ResourceType: resourceType,
		},
		resourceName:    reflect.TypeOf(resourceType).String(),
		secretConverter: sc,
	}, nil
}

var _ clients.ResourceClient = &ResourceClient{}

func (rc *ResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()

	// TODO(yuval-k): do NOT use the lister on read. As Read is mainly called after write, to
	// refresh the object. In theory we could use the object returned from the write call to kubernetes
	// but that requres further investigation.
	secret, err := rc.Kube.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.NewNotExistErr(namespace, name, err)
		}
		return nil, errors.Wrapf(err, "reading secret from kubernetes")
	}

	resource, err := rc.fromKubeResource(opts.Ctx, secret)
	if err == NotOurResource {
		return nil, errors.Errorf("secret %v is not kind %v", name, rc.Kind())
	}
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (rc *ResourceClient) resourceToKubeSecret(ctx context.Context, resource resources.Resource) (*v1.Secret, error) {

	if rc.secretConverter != nil {
		secret, err := rc.secretConverter.ToKubeSecret(ctx, rc, resource)
		if err != nil || secret != nil {
			return secret, err
		}
	}

	return rc.ToKubeSecret(ctx, resource)
}

func (rc *ResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	secret, err := rc.resourceToKubeSecret(opts.Ctx, resource)
	if err != nil {
		return nil, err
	}

	original, err := rc.Read(meta.Namespace, meta.Name, clients.ReadOpts{
		Ctx: opts.Ctx,
	})
	if original != nil && err == nil {
		if !opts.OverwriteExisting {
			return nil, errors.NewExistErr(meta)
		}
		if meta.ResourceVersion != original.GetMetadata().ResourceVersion {
			return nil, errors.NewResourceVersionErr(meta.Namespace, meta.Name, meta.ResourceVersion, original.GetMetadata().ResourceVersion)
		}
		if _, err := rc.Kube.CoreV1().Secrets(secret.Namespace).Update(secret); err != nil {
			return nil, errors.Wrapf(err, "updating kube secret %v", secret.Name)
		}
	} else {
		if _, err := rc.Kube.CoreV1().Secrets(secret.Namespace).Create(secret); err != nil {
			return nil, errors.Wrapf(err, "creating kube secret %v", secret.Name)
		}
	}

	// return a read object to update the resource version
	return rc.Read(secret.Namespace, secret.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *ResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	if !rc.exist(namespace, name) {
		if !opts.IgnoreNotExist {
			return errors.NewNotExistErr(namespace, name)
		}
		return nil
	}

	if err := rc.Kube.CoreV1().Secrets(namespace).Delete(name, nil); err != nil {
		return errors.Wrapf(err, "deleting secret %v", name)
	}
	return nil
}

func (rc *ResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()

	if rc.cache.NamespacedSecretLister(namespace) == nil {
		return nil, errors.Errorf("namespaces is not watched")
	}
	secretList, err := rc.cache.NamespacedSecretLister(namespace).List(labels.SelectorFromSet(opts.Selector))
	if err != nil {
		return nil, errors.Wrapf(err, "listing secrets in %v", namespace)
	}
	var resourceList resources.ResourceList
	for _, secret := range secretList {
		resource, err := rc.fromKubeResource(opts.Ctx, secret)
		// not our resource, ignore it
		if err == NotOurResource {
			continue
		}
		if err != nil {
			return nil, err
		}
		resourceList = append(resourceList, resource)
	}

	sort.SliceStable(resourceList, func(i, j int) bool {
		return resourceList[i].GetMetadata().Name < resourceList[j].GetMetadata().Name
	})

	return resourceList, nil
}

func (rc *ResourceClient) fromKubeResource(ctx context.Context, secret *v1.Secret) (resources.Resource, error) {

	if rc.secretConverter != nil {
		r, err := rc.secretConverter.FromKubeSecret(ctx, rc, secret)
		if r != nil || err != nil {
			return r, err
		}
	}

	return rc.FromKubeSecret(secret)
}

func (rc *ResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	return common.KubeResourceWatch(rc.cache, rc.List, namespace, opts)
}

func (rc *ResourceClient) exist(namespace, name string) bool {
	_, err := rc.Kube.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	return err == nil
}
