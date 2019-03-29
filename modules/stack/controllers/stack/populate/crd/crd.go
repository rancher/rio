package crd

import (
	"fmt"
	"strings"

	"github.com/rancher/mapper/convert"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/sirupsen/logrus"
	v1beta12 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Populate(stack *v1.StackFile, output *objectset.ObjectSet) error {
	if err := crdsForCRDDefs(true, stack.Kubernetes.NamespacedCustomResourceDefinitions, output); err != nil {
		return err
	}

	return crdsForCRDDefs(false, stack.Kubernetes.CustomResourceDefinitions, output)
}

func crdsForCRDDefs(namespaced bool, crdDefs []riov1.CustomResourceDefinition, output *objectset.ObjectSet) error {
	for _, crdDef := range crdDefs {
		plural := name.GuessPluralName(strings.ToLower(crdDef.Kind))
		crdName := strings.ToLower(fmt.Sprintf("%s.%s", plural, crdDef.Group))
		crd := constructors.NewCustomResourceDefinition("", crdName, v1beta12.CustomResourceDefinition{
			Spec: v1beta12.CustomResourceDefinitionSpec{
				Group: crdDef.Group,
				Names: v1beta12.CustomResourceDefinitionNames{
					Kind:     crdDef.Kind,
					ListKind: crdDef.Kind + "List",
					Plural:   plural,
				},
				Version: crdDef.Version,
			},
		})

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

		output.Add(&unstructured.Unstructured{
			Object: crdObj,
		})
	}

	return nil
}
