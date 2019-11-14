package util

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	services2 "github.com/rancher/rio/pkg/services"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func ListPods(ctx *clicontext.CLIContext, name string) ([]corev1.Pod, error) {
	obj, err := ctx.ByID(name)
	if err != nil {
		return nil, err
	}

	if v, ok := obj.Object.(*corev1.Pod); ok {
		return []corev1.Pod{
			*v,
		}, nil
	}

	podName, sel, err := ToPodNameOrSelector(obj.Object)
	if err != nil {
		return nil, err
	}

	if podName != "" {
		pod, err := ctx.Core.Pods(obj.Namespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		return []corev1.Pod{
			*pod,
		}, nil
	}

	return bySelector(ctx, sel)
}

func ToPodNameOrSelector(obj runtime.Object) (string, labels.Selector, error) {
	switch v := obj.(type) {
	case *corev1.Pod:
		return v.Name, nil, nil
	case *riov1.Service:
		app, version := services2.AppAndVersion(v)
		return "", labels.SelectorFromSet(map[string]string{
			"app":     app,
			"version": version,
		}), nil
	case *v1alpha1.TaskRun:
		return v.Status.PodName, nil, nil
	case *appv1.Deployment:
		return toSelector(v.Spec.Selector)
	case *appv1.DaemonSet:
		return toSelector(v.Spec.Selector)
	}

	return "", labels.Nothing(), nil
}

func toSelector(sel *metav1.LabelSelector) (string, labels.Selector, error) {
	l, err := metav1.LabelSelectorAsSelector(sel)
	return "", l, err
}

func bySelector(ctx *clicontext.CLIContext, selector labels.Selector) (ret []corev1.Pod, err error) {
	pds, err := ctx.Core.Pods(ctx.GetSetNamespace()).List(metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}

	for _, pod := range pds.Items {
		ret = append(ret, pod)
	}

	return ret, nil
}
