package systemstatus

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/rancher/rio/pkg/config"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/controllers/pkg"
	"github.com/rancher/rio/pkg/deployment"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		systemNamespace: rContext.Namespace,
		rioinfos:        rContext.Admin.Admin().V1().RioInfo(),
		configs:         rContext.Core.Core().V1().ConfigMap().Cache(),
		pods:            rContext.Core.Core().V1().Pod().Cache(),
	}

	rContext.Apps.Apps().V1().Deployment().OnChange(ctx, "systemstatus-check", h.sync)
	return nil
}

type handler struct {
	systemNamespace string
	rioinfos        adminv1controller.RioInfoController
	configs         corev1controller.ConfigMapCache
	pods            corev1controller.PodCache
}

func (h handler) sync(key string, obj *appsv1.Deployment) (*appsv1.Deployment, error) {
	if obj == nil || obj.DeletionTimestamp != nil {
		return obj, nil
	}

	if obj.Namespace != h.systemNamespace {
		return obj, nil
	}

	infoObj, err := h.rioinfos.Get("rio", metav1.GetOptions{})
	if err != nil {
		return obj, err
	}

	info := infoObj.DeepCopy()
	readyMap := info.Status.SystemComponentReadyMap
	if readyMap == nil {
		readyMap = make(map[string]string)
	}

	deploymentReady := deployment.IsReady(&obj.Status)
	if deploymentReady {
		readyMap[obj.Name] = "Ready"
	} else {
		r, err := labels.NewRequirement("app", selection.Equals, []string{obj.Name})
		if err != nil {
			return nil, err
		}
		l := labels.NewSelector().Add(*r)
		pods, err := h.pods.List(obj.Namespace, l)
		if err != nil {
			return nil, err
		}
		if len(pods) > 0 {
			readyMap[obj.Name] = pkg.PodDetail(pods[0])
		}
	}

	cm, err := h.configs.Get(h.systemNamespace, config.ConfigName)
	if err != nil {
		return nil, err
	}
	r, err := config.FromConfigMap(cm)
	if err != nil {
		return nil, err
	}

	var ready bool
	if constants.DevMode {
		ready = true
	} else {
		if _, err := http.Get(fmt.Sprintf("http://%s.%s", r.Gateway.ServiceName, r.Gateway.ServiceNamespace)); err == nil {
			ready = true
		} else {
			ready = false
		}
	}

	var update bool
	if !reflect.DeepEqual(info.Status.SystemComponentReadyMap, readyMap) {
		info.Status.SystemComponentReadyMap = readyMap
		update = true
	}

	if info.Status.Ready != ready && ready {
		if ready {
			logrus.Info("Able to access gateway service, set systemState to ready")
		}
		info.Status.Ready = ready
		update = true
	}

	if update {
		if _, err := h.rioinfos.Update(info); err != nil {
			return obj, err
		}
	}

	return obj, nil
}
