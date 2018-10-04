package input

import (
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	"k8s.io/api/core/v1"
)

type IstioDeployment struct {
	LBNamespace     *v1.Namespace
	LBService       *v1.Service
	VirtualServices []*v1alpha3.VirtualService
}
