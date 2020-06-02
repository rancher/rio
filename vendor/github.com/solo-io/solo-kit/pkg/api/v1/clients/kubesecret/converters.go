package kubesecret

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"
	v1 "k8s.io/api/core/v1"
)

type SecretConverter interface {
	// If this method returns nil,nil the default conversion will be used.
	FromKubeSecret(ctx context.Context, rc *ResourceClient, secret *v1.Secret) (resources.Resource, error)
	// If this method returns nil,nil the default conversion will be used.
	ToKubeSecret(ctx context.Context, rc *ResourceClient, resource resources.Resource) (*v1.Secret, error)
}

type plainSecret struct{}

func (p *plainSecret) FromKubeSecret(ctx context.Context, rc *ResourceClient, secret *v1.Secret) (resources.Resource, error) {
	resource := rc.NewResource()
	// not our secret
	// should be an error on a Read, ignored on a list
	if len(secret.ObjectMeta.Annotations) == 0 || secret.ObjectMeta.Annotations[annotationKey] != rc.Kind() {
		return nil, NotOurResource
	}
	// only works for string fields
	resourceMap := make(map[string]interface{})
	for k, v := range secret.Data {
		resourceMap[k] = string(v)
	}
	if err := protoutils.UnmarshalMap(resourceMap, resource); err != nil {
		return nil, eris.Wrapf(err, "reading secret data into %v", rc.Kind())
	}
	resource.SetMetadata(kubeutils.FromKubeMeta(secret.ObjectMeta))
	return resource, nil
}

func (p *plainSecret) ToKubeSecret(ctx context.Context, rc *ResourceClient, resource resources.Resource) (*v1.Secret, error) {
	resourceMap, err := protoutils.MarshalMapEmitZeroValues(resource)
	if err != nil {
		return nil, eris.Wrapf(err, "marshalling resource as map")
	}
	kubeSecretData := make(map[string][]byte)
	for k, v := range resourceMap {
		// metadata comes from ToKubeMeta
		// status not supported
		if k == "metadata" {
			continue
		}
		switch val := v.(type) {
		case string:
			kubeSecretData[k] = []byte(val)
		default:
			// TODO: handle other field types; for now the caller
			// must know this resource client only supports map[string]string style objects
			contextutils.LoggerFrom(ctx).Warnf("invalid resource type (%v) used for plain secret. unable to "+
				"convert to kube secret. only resources with fields of type string are supported for the plain secret client.", resources.Kind(resource))
		}
	}

	meta := kubeutils.ToKubeMeta(resource.GetMetadata())
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	meta.Annotations[annotationKey] = rc.Kind()
	return &v1.Secret{
		ObjectMeta: meta,
		Data:       kubeSecretData,
	}, nil
}
