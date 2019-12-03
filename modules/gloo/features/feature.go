package features

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"github.com/rancher/rio/modules/gloo/controller/app"
	"github.com/rancher/rio/modules/gloo/controller/config"
	"github.com/rancher/rio/modules/gloo/controller/ingress"
	"github.com/rancher/rio/modules/gloo/controller/router"
	"github.com/rancher/rio/modules/gloo/controller/service"
	"github.com/rancher/rio/modules/gloo/controller/setting"
	"github.com/rancher/rio/modules/gloo/pkg/injector"
	"github.com/rancher/rio/pkg/arch"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/types"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	injector.RegisterInjector()
	crbs := rContext.RBAC.Rbac().V1().ClusterRoleBinding()
	apply := rContext.Apply.WithInjectorName("gloo-mesh").WithPatcher(rbacv1.SchemeGroupVersion.WithKind("ClusterRoleBinding"), func(namespace, name string, pt k8stypes.PatchType, data []byte) (runtime.Object, error) {
		err := crbs.Delete(name, &metav1.DeleteOptions{})
		if err == nil {
			return nil, fmt.Errorf("replace ClusterRoleBinding")
		}
		return nil, err
	})
	feature := &features.FeatureController{
		FeatureName: "gloo",
		FeatureSpec: features.FeatureSpec{
			Enabled:     arch.IsAmd64(),
			Description: "Run Gloo API gateway",
		},
		FixedAnswers: map[string]string{
			"NAMESPACE": rContext.Namespace,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Admin.Admin().V1().SystemStack(), rContext.Namespace, "gloo"),
		},
		Controllers: []features.ControllerRegister{
			app.Register,
			router.Register,
			config.Register,
			service.Register,
			ingress.Register,
			setting.Register,
		},
		OnStart: func() error {
			// handle upgrade
			return upgradeOldVirtualService(rContext.Namespace, rContext.RestConfig)
		},
	}

	return feature.Register()
}

// upgradeOldVirtualService handles compatibility from upgrading gloo from v0.20.x to v1.0.0 for virtualservices.
func upgradeOldVirtualService(systemNamespace string, config *rest.Config) error {
	dynamicClient := dynamic.NewForConfigOrDie(config)
	gvk := schema.GroupVersionResource{
		Group:    "gateway.solo.io",
		Resource: "virtualservices",
		Version:  "v1",
	}
	vss, err := dynamicClient.Resource(gvk).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, vs := range vss.Items {
		slices, ok, _ := unstructured.NestedSlice(vs.Object, "spec", "virtualHost", "routes")
		if ok {
			for _, slice := range slices {
				m, ok := slice.(map[string]interface{})
				if ok {
					if _, ok = m["matcher"]; ok {
						name, ok1, _ := unstructured.NestedString(vs.Object, "metadata", "name")
						namespace, ok2, _ := unstructured.NestedString(vs.Object, "metadata", "namespace")
						if ok1 && ok2 {
							if err := dynamicClient.Resource(gvk).Namespace(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	gvkProxy := schema.GroupVersionResource{
		Group:    "gloo.solo.io",
		Resource: "proxies",
		Version:  "v1",
	}
	proxyName := "gateway-proxy-v2"
	if _, err := dynamicClient.Resource(gvkProxy).Namespace(systemNamespace).Get(proxyName, metav1.GetOptions{}); err == nil {
		if err := dynamicClient.Resource(gvkProxy).Namespace(systemNamespace).Delete(proxyName, &metav1.DeleteOptions{}); err != nil {
			return err
		}
	}
	return err
}
