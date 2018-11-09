package crd

import (
	"fmt"
	"strings"

	"github.com/rancher/norman/name"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rio/pkg/deploy/stackdef/output"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"github.com/sirupsen/logrus"
	v1beta12 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Populate(stack *v1beta1.InternalStack, output *output.Deployment) error {
	if err := crdsForCRDDefs(true, stack.Kubernetes.NamespacedCustomResourceDefinitions, output); err != nil {
		return err
	}

	return crdsForCRDDefs(false, stack.Kubernetes.CustomResourceDefinitions, output)
}

func crdsForCRDDefs(namespaced bool, crdDefs []v1beta1.CustomResourceDefinition, output *output.Deployment) error {
	for _, crdDef := range crdDefs {
		plural := name.GuessPluralName(strings.ToLower(crdDef.Kind))
		crdName := strings.ToLower(fmt.Sprintf("%s.%s", plural, crdDef.Group))
		crd := &v1beta12.CustomResourceDefinition{
			TypeMeta: v1.TypeMeta{
				Kind:       "CustomResourceDefinition",
				APIVersion: "apiextensions.k8s.io/v1beta1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: crdName,
			},
			Spec: v1beta12.CustomResourceDefinitionSpec{
				Group: crdDef.Group,
				Names: v1beta12.CustomResourceDefinitionNames{
					Kind:     crdDef.Kind,
					ListKind: crdDef.Kind + "List",
					Plural:   plural,
				},
				Version: crdDef.Version,
			},
		}

		if namespaced {
			crd.Spec.Scope = v1beta12.NamespaceScoped
		} else {
			crd.Spec.Scope = v1beta12.ClusterScoped
		}

		// k8s 1.11 will not accept CRD with status field and marshalling CRD will always put a status field
		// so workaround by converting to map
		crdObj, err := convert.EncodeToMap(crd)
		if err != nil {
			logrus.Errorf("failed to marshal CRD %v: %v", crd, err)
			return err
		}
		delete(crdObj, "status")

		output.CRDs[crd.Name] = &unstructured.Unstructured{
			Object: crdObj,
		}
	}

	return nil
}
