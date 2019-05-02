package appweight

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	sh := appWeightHandler{
		services: rContext.Rio.Rio().V1().Service(),
		apps:     rContext.Rio.Rio().V1().App(),
	}

	updator := riov1controller.UpdateAppOnChange(rContext.Rio.Rio().V1().App().Updater(), sh.sync)
	rContext.Rio.Rio().V1().App().OnChange(ctx, "app-weight", updator)
	return nil
}

type appWeightHandler struct {
	services riov1controller.ServiceController
	apps     riov1controller.AppController
}

func (s appWeightHandler) sync(key string, obj *riov1.App) (*riov1.App, error) {
	if obj == nil {
		return nil, nil
	}

	app := obj.DeepCopy()

	if app.Status.RevisionWeight == nil {
		app.Status.RevisionWeight = make(map[string]riov1.ServiceObservedWeight, 0)
	}

	// set initial weight status
	if len(app.Status.RevisionWeight) == 0 {
		var added int
		for i, rev := range app.Spec.Revisions {
			if i != len(app.Spec.Revisions)-1 {
				add := int(100.0 / float64(len(app.Spec.Revisions)))
				app.Status.RevisionWeight[rev.Version] = riov1.ServiceObservedWeight{
					Weight:      add,
					LastWrite:   metav1.NewTime(time.Now()),
					ServiceName: rev.ServiceName,
				}
				added += add
			} else {
				app.Status.RevisionWeight[rev.Version] = riov1.ServiceObservedWeight{
					Weight:      100 - added,
					LastWrite:   metav1.NewTime(time.Now()),
					ServiceName: rev.ServiceName,
				}
			}
		}
	}

	versMap := map[string]struct{}{}
	var toDeletes []string
	for _, rev := range app.Spec.Revisions {
		versMap[rev.Version] = struct{}{}
	}
	for ver := range app.Status.RevisionWeight {
		if _, ok := versMap[ver]; !ok {
			toDeletes = append(toDeletes, ver)
		}
	}
	for _, toDelete := range toDeletes {
		if app.Status.RevisionWeight[toDelete].Weight == 0 {
			logrus.Infof("cleaning up non-existing revision %s", toDelete)
			delete(app.Status.RevisionWeight, toDelete)
		}
	}

	ready := true
	for _, rev := range app.Spec.Revisions {
		if !rev.DeploymentReady && rev.AdjustedWeight != 0 {
			ready = false
			break
		}
	}
	if !ready {
		return app, nil
	}

	for index, rev := range app.Spec.Revisions {
		observed := app.Status.RevisionWeight[rev.Version]
		observed.ServiceName = rev.ServiceName
		if rev.AdjustedWeight == observed.Weight {
			continue
		}

		weightToAdjust := rev.AdjustedWeight - observed.Weight
		revision := app.DeepCopy().Spec.Revisions
		versions, weights := versionAndSpecs(append(revision[0:index], revision[index+1:]...))
		for _, toDelete := range toDeletes {
			if app.Status.RevisionWeight[toDelete].Weight > 0 {
				versions = append(versions, toDelete)
				weights = append(weights, 0)
			}
		}

		if isRolloutSet(rev) {
			if time.Now().Before(observed.LastWrite.Add(time.Second * time.Duration(rev.RolloutInterval))) {
				go func() {
					time.Sleep(time.Second * time.Duration(rev.RolloutInterval))
					s.apps.Enqueue(app.Namespace, app.Name)
				}()
				break
			}

			if abs(weightToAdjust) < rev.RolloutIncrement {
				observed.Weight += weightToAdjust
				magicSteal(versions, weights, app.Status.RevisionWeight, -weightToAdjust)
			} else {
				rolloutamount := rev.RolloutIncrement
				if weightToAdjust < 0 {
					rolloutamount = -rolloutamount
				}
				observed.Weight += rolloutamount
				magicSteal(versions, weights, app.Status.RevisionWeight, -rolloutamount)
			}
			time.Sleep(time.Second * time.Duration(rev.RolloutInterval))
			observed.LastWrite = metav1.NewTime(time.Now())
			app.Status.RevisionWeight[rev.Version] = observed
		} else {
			weightRev := app.Status.RevisionWeight[rev.Version]
			weightRev.Weight += weightToAdjust
			app.Status.RevisionWeight[rev.Version] = weightRev
			magicSteal(versions, weights, app.Status.RevisionWeight, -weightToAdjust)
		}
		// only execute one revision at one sync call
		break
	}
	return app, nil
}

func isRolloutSet(rev riov1.Revision) bool {
	return rev.Rollout && rev.RolloutIncrement != 0 && rev.RolloutInterval > 0
}

func versionAndSpecs(specs []riov1.Revision) ([]string, []int) {
	var versions []string
	var weights []int
	for _, spec := range specs {
		versions = append(versions, spec.Version)
		weights = append(weights, spec.AdjustedWeight)
	}
	return versions, weights
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

/*
	Steal weight from other service. Don't try to read it. :)
*/
func magicSteal(versions []string, weightSpecs []int, result map[string]riov1.ServiceObservedWeight, weightToAdjust int) {
	if len(versions) == 0 {
		return
	}

	for i, ver := range versions {
		rev := result[ver]
		toAdjust := rev.Weight - weightSpecs[i]
		if negative(toAdjust, weightToAdjust) {
			if toAdjust+weightToAdjust >= 0 {
				if abs(toAdjust) > abs(weightToAdjust) {
					rev.Weight = rev.Weight + weightToAdjust
					weightToAdjust = 0
					result[ver] = rev
					break
				} else {
					rev.Weight = rev.Weight - toAdjust
					weightToAdjust += toAdjust
					result[ver] = rev
				}
			}
		}
	}

	if weightToAdjust == 0 {
		return
	}

	for _, ver := range versions {
		rev := result[ver]
		if weightToAdjust < 0 && rev.Weight < weightToAdjust {
			weightToAdjust += rev.Weight
			rev.Weight = 0
			result[versions[0]] = rev
			continue
		} else {
			rev.Weight += weightToAdjust
			weightToAdjust = 0
			result[versions[0]] = rev
			break
		}
	}

	return
}

func negative(a, b int) bool {
	return a*b < 0
}
