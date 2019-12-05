package config

import (
	"context"
	"fmt"

	"github.com/rancher/rio/pkg/config"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/types"
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
		key:      fmt.Sprintf("%s/%s", rContext.Namespace, config.ConfigName),
		rContext: rContext,
		ctx:      ctx,
	}

	rContext.Core.Core().V1().ConfigMap().OnChange(ctx, "config", fh.onChange)
	return nil
}

type featureHandler struct {
	key          string
	ctx          context.Context
	rContext     *types.Context
	featureState sets.String
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
	return obj, f.rContext.Start(f.ctx)
}
