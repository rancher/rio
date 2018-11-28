package service

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/rancher/rio/api/service"
	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/pkg/deploy/stack/populate/k8sservice"
	"github.com/rancher/rio/pkg/deploy/stack/populate/podcontrollers"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	appsv1 "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Populate(stack *input.Stack, output *output.Deployment) error {
	var err error

	serviceSet, err := CollectionServices(stack.Services)
	if err != nil {
		return err
	}

	for _, s := range serviceSet.List() {
		k8sservice.Populate(stack, s, output)
		if err := podcontrollers.Populate(stack, s, output); err != nil {
			return err
		}
	}

	for _, externalService := range stack.ExternalServices {
		populateServiceFromExternal(externalService, output)
	}

	for _, route := range stack.RouteSet {
		populateDeploymentFromRoute(stack.Space, route, output)
		populateServiceFromRoute(route, output)
	}

	return nil
}

func populateServiceFromExternal(e *v1beta1.ExternalService, output *output.Deployment) error {
	target := e.Spec.Target
	if !strings.HasPrefix(target, "https://") && !strings.HasPrefix(target, "http://") {
		target = "http://" + target
	}
	u, err := url.Parse(target)
	if err != nil {
		return err
	}
	if ip := net.ParseIP(u.Host); ip == nil {
		// set service to external name
		service := &v1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      e.Name,
				Namespace: e.Namespace,
				Labels:    e.Labels,
			},
			Spec: v1.ServiceSpec{
				Type:         v1.ServiceTypeExternalName,
				ExternalName: u.Host,
			},
		}
		output.Services[service.Name] = service
	} else {
		targetPort, _ := strconv.ParseInt(u.Port(), 10, 64)
		if targetPort == 0 {
			if u.Scheme == "http" {
				targetPort = 80
			} else if u.Scheme == "https" {
				targetPort = 443
			}
		}
		service := &v1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      e.Name,
				Namespace: e.Namespace,
				Labels:    e.Labels,
			},
			Spec: v1.ServiceSpec{
				Type:         v1.ServiceTypeExternalName,
				ExternalName: replaceWithXIP(u.Host),
			},
		}
		output.Services[service.Name] = service
	}
	return nil
}

func replaceWithXIP(ip string) string {
	return fmt.Sprintf("%s.xip.io", ip)
}

func populateServiceFromRoute(r *v1beta1.RouteSet, output *output.Deployment) {
	service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.Namespace,
			Labels: map[string]string{
				"app":                   r.Name,
				"rio.cattle.io/version": "v0",
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app":                   r.Name,
				"rio.cattle.io/version": "v0",
			},
			Ports: []v1.ServicePort{
				{
					Name:       "http-80-80",
					Protocol:   v1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
		},
	}
	output.Services[service.Name] = service
}

func getExternalDomain(name, namespace, space string) string {
	return fmt.Sprintf("%s.%s", service.HashIfNeed(name, strings.SplitN(namespace, "-", 2)[0], space), settings.ClusterDomain.Get())
}

func populateDeploymentFromRoute(space string, r *v1beta1.RouteSet, output *output.Deployment) {
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        r.Name,
			Namespace:   r.Namespace,
			Labels:      r.Labels,
			Annotations: map[string]string{},
		},
		Spec: appsv1.DeploymentSpec{
			Paused:   false,
			Replicas: &[]int32{1}[0],
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":                   r.Name,
					"rio.cattle.io/version": "v0",
				},
			},
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "route-redirect",
							Image:           settings.RouteStubtImage.Get(),
							ImagePullPolicy: v1.PullAlways,
							Stdin:           true,
							TTY:             true,
							Env: []v1.EnvVar{
								{
									Name:  "SERVER_REDIRECT",
									Value: getExternalDomain(r.Name, r.Namespace, space),
								},
							},
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                   r.Name,
						"rio.cattle.io/version": "v0",
					},
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
		},
	}
	output.Deployments[dep.Name] = dep
}
