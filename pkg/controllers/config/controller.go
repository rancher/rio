package config

import (
	"context"
	"fmt"

	"github.com/rancher/rio/pkg/config"
	"github.com/rancher/rio/pkg/features"
	adminv1 "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/slice"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func Register(ctx context.Context, rContext *types.Context) error {
	_, err := rContext.Core.Core().V1().ConfigMap().Create(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.ConfigName,
			Namespace: rContext.Namespace,
		},
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	fh := &featureHandler{
		key:        fmt.Sprintf("%s/%s", rContext.Namespace, config.ConfigName),
		infos:      rContext.Admin.Admin().V1().RioInfo(),
		rContext:   rContext,
		ctx:        ctx,
		namespace:  rContext.Namespace,
		configmaps: rContext.Core.Core().V1().ConfigMap(),
	}

	rContext.Core.Core().V1().ConfigMap().OnChange(ctx, "config", fh.onChange)
	return nil
}

type featureHandler struct {
	key          string
	ctx          context.Context
	rContext     *types.Context
	featureState sets.String
	infos        adminv1.RioInfoClient
	configmaps   corev1controller.ConfigMapClient
	namespace    string
}

func (f *featureHandler) getFeatureConfig(obj *corev1.ConfigMap) (enabled sets.String, err error) {
	all := sets.String{}
	enabled = sets.String{}
	for _, f := range features.GetFeatures() {
		all.Insert(f.Name())
		if f.Spec().Enabled {
			enabled.Insert(f.Name())
		}
	}

	if obj == nil {
		return
	}

	configStr := obj.Data["config"]
	if configStr == "" {
		return
	}

	config, err := config.FromConfigMap(obj)
	if err != nil {
		return enabled, err
	}

	wildcard, ok := config.Features["*"]
	if ok && wildcard.Enabled != nil {
		if *wildcard.Enabled {
			enabled.Insert(all.List()...)
		} else {
			enabled = sets.String{}
		}
	}

	for name, f := range config.Features {
		if f.Enabled == nil {
			continue
		}
		if !*f.Enabled {
			enabled.Delete(name)
		} else {
			enabled.Insert(name)
		}
	}

	return
}

func (f *featureHandler) start(started, enable sets.String, name string) (string, bool, error) {
	if started.Has(name) {
		return "", true, nil
	}

	feature := features.GetFeature(name)
	if feature == nil {
		return "", true, fmt.Errorf("failed to find feature %s", name)
	}

	for _, dep := range feature.Spec().Requires {
		missing, ok, err := f.start(started, enable, dep)
		if err != nil || !ok {
			return missing, ok, err
		}
	}

	if enable.Has(name) {
		started.Insert(name)
		return "", true, feature.Start(f.ctx)
	}

	return name, false, nil
}

func (f *featureHandler) onChange(key string, obj *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if key != f.key {
		return obj, nil
	}

	newState, err := f.getFeatureConfig(obj)
	if err != nil {
		return obj, err
	}

	if f.featureState == nil {
	} else if f.featureState.Equal(newState) {
		return obj, nil
	} else {
		logrus.Fatal("Feature configuration has changed, quiting")
	}

	started := sets.String{}
	for _, feature := range features.GetFeatures() {
		name := feature.Name()
		if newState.Has(name) {
			if missing, ok, err := f.start(started, newState, name); err != nil {
				return obj, err
			} else if !ok {
				if name != missing {
					logrus.Infof("Not starting %s because feature %s is not enabled", name, missing)
				}
			}
		} else {
			err := feature.Stop()
			if err != nil {
				logrus.Errorf("failed to stop feature %s: %v", name, err)
			}
		}
	}

	f.featureState = newState
	info, err := f.infos.Get("rio", metav1.GetOptions{})
	if err != nil {
		return obj, err
	}
	if !slice.StringsEqual(info.Status.EnabledFeatures, newState.List()) {
		info.Status.EnabledFeatures = newState.List()
		if _, err := f.infos.Update(info); err != nil {
			return obj, err
		}
	}

	if err := f.updateConfigMap(obj); err != nil {
		return obj, err
	}
	return obj, f.rContext.Start(f.ctx)
}

func (f *featureHandler) updateConfigMap(obj *corev1.ConfigMap) error {
	conf, err := config.FromConfigMap(obj)
	if err != nil {
		return err
	}
	if conf.Features == nil {
		conf.Features = map[string]config.FeatureConfig{}
	}

	state := f.featureState

	t := true
	for _, feature := range features.GetFeatures() {
		f := conf.Features[feature.Name()]
		if state.Has(feature.Name()) {
			f.Enabled = &t
		} else {
			f.Enabled = new(bool)
		}
		f.Description = feature.Spec().Description
		conf.Features[feature.Name()] = f
	}
	cm, err := config.SetConfig(obj, conf)
	if err != nil {
		return err
	}

	if _, err := f.configmaps.Update(cm); err != nil {
		return err
	}
	return nil
}
