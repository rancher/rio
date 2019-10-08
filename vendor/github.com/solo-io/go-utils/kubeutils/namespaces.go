package kubeutils

import (
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNamespacesInParallel(kube kubernetes.Interface, namespaces ...string) error {
	eg := errgroup.Group{}
	for _, namespace := range namespaces {
		namespace := namespace
		eg.Go(func() error {
			_, err := kube.CoreV1().Namespaces().Create(&v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			})
			return err
		})
	}
	return eg.Wait()
}

func DeleteNamespacesInParallelBlocking(kube kubernetes.Interface, namespaces ...string) error {
	eg := errgroup.Group{}
	for _, namespace := range namespaces {
		namespace := namespace
		eg.Go(func() error {
			return kube.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
		})
	}
	return eg.Wait()
}

func DeleteNamespacesInParallel(kube kubernetes.Interface, namespaces ...string) {
	for _, namespace := range namespaces {
		namespace := namespace
		go func() {
			kube.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
		}()
	}
}
