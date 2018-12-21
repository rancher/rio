package feature

import (
	"context"
	"sync"
	"time"

	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	all = "_all_"
)

func Register(ctx context.Context, rContext *types.Context) error {
	rContext.Global.Feature.Generic().SetThreadinessOverride(1)

	f := &featureHandler{
		ctx:            ctx,
		featuresClient: rContext.Global.Feature,
		featuresCache:  rContext.Global.Feature.Cache(),
		featureState:   map[string]func(){},
		processor: objectset.NewProcessor("features").
			Client(rContext.Global.Feature),
	}

	f.featuresClient.Interface().Controller().AddHandler(ctx, "features-controller", f.sync)
	f.featuresClient.OnChange(ctx, "features-controller", f.onChange)
	f.featuresClient.OnRemove(ctx, "features-controller", f.onRemove)
	f.featuresClient.Enqueue("", all)

	return nil
}

type featureHandler struct {
	sync.Mutex

	ctx            context.Context
	processor      objectset.Processor
	featuresClient v1.FeatureClient
	featuresCache  v1.FeatureClientCache
	featureState   map[string]func()
}

func (f *featureHandler) sync(key string, obj *v1.Feature) (runtime.Object, error) {
	if key != all {
		return obj, nil
	}

	os := objectset.NewObjectSet()
	for _, feature := range features.GetFeatures() {
		featureObj := v1.NewFeature(settings.RioSystemNamespace, feature.Name(), v1.Feature{
			Spec: feature.Spec(),
		})
		os.Add(featureObj)
	}

	return obj, f.processor.NewDesiredSet(nil, os).Apply()
}

func (f *featureHandler) onRemove(obj *v1.Feature) (runtime.Object, error) {
	if obj.Namespace != settings.RioSystemNamespace {
		return obj, nil
	}

	feature := features.GetFeature(obj.Name)
	if feature == nil {
		return obj, nil
	}

	return f.stop(obj, feature)
}

func (f *featureHandler) onChange(obj *v1.Feature) (runtime.Object, error) {
	if obj.Namespace != settings.RioSystemNamespace {
		return obj, nil
	}

	if obj.Spec.Enabled {
		if err := f.checkDeps(obj); err != nil {
			return obj, err
		}
	}

	feature := features.GetFeature(obj.Name)
	if feature == nil {
		return obj, nil
	}

	if !obj.Spec.Enabled {
		return f.stop(obj, feature)
	}

	if err := f.start(obj, feature); err != nil {
		return obj, err
	}

	return obj, feature.Changed(obj)
}

func (f *featureHandler) checkDeps(obj *v1.Feature) error {
	for _, depName := range obj.Spec.Requires {
		dep, err := f.getDepFeature(depName, obj.Namespace)
		if !dep.Spec.Enabled {
			dep = dep.DeepCopy()
			dep.Spec.Enabled = true
			dep, err = f.featuresClient.Update(dep)
			if err != nil {
				return err
			}
		}

		if _, err := f.onChange(dep); err != nil {
			return err
		}
	}

	return nil
}

func (f *featureHandler) start(obj *v1.Feature, feature features.Feature) error {
	if f.isEnabled(obj.Name) {
		v1.FeatureConditionEnabled.True(obj)
		return nil
	}

	_, err := v1.FeatureConditionEnabled.Do(obj, func() (runtime.Object, error) {
		subCtx, cancel := context.WithCancel(f.ctx)
		logrus.Infof("Starting feature %s", feature.Name())
		if err := feature.Start(subCtx, obj); err != nil {
			cancel()
			return obj, err
		}

		go func() {
			<-f.ctx.Done()
			cancel()
		}()

		f.featureState[feature.Name()] = cancel
		return obj, nil
	})

	return err
}

func (f *featureHandler) stop(obj *v1.Feature, feature features.Feature) (runtime.Object, error) {
	return v1.FeatureConditionEnabled.Do(obj, func() (runtime.Object, error) {
		err := feature.Stop()
		if err != nil {
			return obj, err
		}

		c, ok := f.featureState[feature.Name()]
		if ok {
			c()
		}

		return obj, nil
	})
}

func (f *featureHandler) isEnabled(name string) bool {
	f.Lock()
	defer f.Unlock()
	_, ok := f.featureState[name]
	return ok
}

func (f *featureHandler) getDepFeature(depName, namespace string) (*v1.Feature, error) {
	start := time.Millisecond * 250
	var feature *v1.Feature
	var err error
	for i := 0; i < 5; i++ {
		feature, err = f.featuresCache.Get(namespace, depName)
		if err == nil {
			break
		}
		time.Sleep(start)
		start *= 2
	}
	return feature, err
}
