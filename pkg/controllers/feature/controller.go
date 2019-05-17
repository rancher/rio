package feature

import (
	"context"
	"sync"
	"time"

	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
	trigger2 "github.com/rancher/wrangler/pkg/trigger"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	rContext.Global.SetThreadiness(v1.SchemeGroupVersion.WithKind("Feature"), 1)

	f := &featureHandler{
		ctx:            ctx,
		namespace:      rContext.Namespace,
		featuresClient: rContext.Global.Admin().V1().Feature(),
		featuresCache:  rContext.Global.Admin().V1().Feature().Cache(),
		featureState:   map[string]func(){},
		apply: rContext.Apply.
			WithSetID("features").
			WithStrictCaching().
			WithCacheTypes(rContext.Global.Admin().V1().Feature()),
	}

	trigger := trigger2.New(f.featuresClient)
	trigger.OnTrigger(ctx, "features-controller", f.syncAll)

	f.featuresClient.OnChange(ctx, "features-controller", projectv1controller.UpdateFeatureOnChange(f.featuresClient.Updater(), f.onChange))
	f.featuresClient.OnRemove(ctx, "features-controller", projectv1controller.UpdateFeatureOnChange(f.featuresClient.Updater(), f.onRemove))

	return nil
}

type featureHandler struct {
	sync.Mutex

	namespace      string
	ctx            context.Context
	apply          apply.Apply
	featuresClient projectv1controller.FeatureController
	featuresCache  projectv1controller.FeatureCache
	featureState   map[string]func()
}

func (f *featureHandler) syncAll() error {
	os := objectset.NewObjectSet()
	for _, feature := range features.GetFeatures() {
		featureObj := v1.NewFeature(f.namespace, feature.Name(), v1.Feature{
			Spec: feature.Spec(),
		})
		if feature.IsSystem() {
			featureObj.Labels = map[string]string{
				"rio.cattle.io/system": "true",
			}
		}
		os.Add(featureObj)
	}

	return f.apply.Apply(os)
}

func (f *featureHandler) onRemove(key string, obj *v1.Feature) (*v1.Feature, error) {
	if obj == nil {
		return nil, nil
	}

	if obj.Namespace != f.namespace {
		return obj, nil
	}

	feature := features.GetFeature(obj.Name)
	if feature == nil {
		return obj, nil
	}

	return f.stop(obj, feature)
}

func (f *featureHandler) onChange(key string, obj *v1.Feature) (*v1.Feature, error) {
	if obj == nil {
		return nil, nil
	}

	if obj.Namespace != f.namespace {
		return obj, nil
	}

	if isEnabled(obj) {
		if err := f.checkDeps(obj); err != nil {
			return obj, err
		}
	}

	feature := features.GetFeature(obj.Name)
	if feature == nil {
		return obj, nil
	}

	if !isEnabled(obj) {
		return f.stop(obj, feature)
	}

	if err := f.start(obj, feature); err != nil {
		return obj, err
	}

	return obj, feature.Changed(obj)
}

func isEnabled(obj *v1.Feature) bool {
	if obj.Status.EnableOverride != nil {
		return *obj.Status.EnableOverride
	}
	return obj.Spec.Enabled
}

func (f *featureHandler) checkDeps(obj *v1.Feature) error {
	t := true
	for _, depName := range obj.Spec.Requires {
		dep, err := f.getDepFeature(depName, obj.Namespace)
		if err != nil {
			return err
		}
		if !isEnabled(dep) {
			dep = dep.DeepCopy()
			dep.Status.EnableOverride = &t
			dep, err = f.featuresClient.Update(dep)
			if err != nil {
				return err
			}
		}

		if _, err := f.onChange("", dep); err != nil {
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

	return v1.FeatureConditionEnabled.Do(func() (runtime.Object, error) {
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
}

func (f *featureHandler) stop(obj *v1.Feature, feature features.Feature) (*v1.Feature, error) {
	return obj, v1.FeatureConditionEnabled.Do(func() (runtime.Object, error) {
		err := feature.Stop()
		if err != nil {
			return obj, err
		}

		c, ok := f.featureState[feature.Name()]
		if ok {
			c()
		}
		delete(f.featureState, feature.Name())
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
