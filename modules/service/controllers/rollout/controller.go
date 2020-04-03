package rollout

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/indexes"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/generic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const PromoteWeight = 10000

type rolloutHandler struct {
	services      riov1controller.ServiceController
	serviceCache  riov1controller.ServiceCache
	client        riov1controller.ServiceClient
	lastWrite     map[string]metav1.Time
	lastWriteLock sync.RWMutex
}

func Register(ctx context.Context, rContext *types.Context) error {
	rh := &rolloutHandler{
		services:      rContext.Rio.Rio().V1().Service(),
		serviceCache:  rContext.Rio.Rio().V1().Service().Cache(),
		client:        rContext.Rio.Rio().V1().Service(),
		lastWrite:     make(map[string]metav1.Time),
		lastWriteLock: sync.RWMutex{},
	}
	rContext.Rio.Rio().V1().Service().OnChange(ctx, "rollout", rh.rollout)
	return nil
}

func (rh *rolloutHandler) rollout(key string, svc *riov1.Service) (*riov1.Service, error) {
	if svc == nil || svc.DeletionTimestamp != nil {
		return nil, nil
	}
	appName, _ := services.AppAndVersion(svc)
	if appName == "" {
		return nil, generic.ErrSkip
	}
	allSvcs, err := rh.serviceCache.GetByIndex(indexes.ServiceByApp, fmt.Sprintf("%s/%s", svc.Namespace, appName))
	if err != nil {
		return svc, err
	}
	var svcs []*riov1.Service
	for _, s := range allSvcs {
		// We don't ever want to assign weight to a template
		if !s.Spec.Template {
			svcs = append(svcs, s)
		}
	}
	if len(svcs) == 0 {
		return svc, nil
	}

	// When services are initiated with no weight or computedWeight, set initial ComputedWeights balanced evenly
	var updatedNeeded []string
	if !computedWeightsExist(svcs) {
		var added int
		add := int(float64(PromoteWeight) / float64(len(svcs)))
		for i, s := range svcs {
			s.Status.ComputedWeight = new(int)
			if i != len(svcs)-1 {
				*s.Status.ComputedWeight = add
				added += add
			} else {
				*s.Status.ComputedWeight = PromoteWeight - added
			}
			updatedNeeded = append(updatedNeeded, serviceKey(s))
		}
	}

	for _, s := range svcs {
		// If pause is on, or if any revision is not ready but has weight allocated, return
		if blocksRollout(s.Spec.RolloutConfig) || (!s.Status.DeploymentReady && s.Spec.Weight != nil && *s.Spec.Weight > 0) {
			err = rh.updateServices(svcs, updatedNeeded)
			if err != nil {
				return svc, err
			}
			return svc, nil
		}
	}
	for _, s := range svcs {
		if s.Spec.Weight == nil || (s.Status.ComputedWeight == nil && *s.Spec.Weight == 0) || (s.Status.ComputedWeight != nil && (*s.Spec.Weight == *s.Status.ComputedWeight)) {
			continue // this rev is already at desired weight, nothing to do
		}
		if s.Status.ComputedWeight == nil {
			s.Status.ComputedWeight = new(int)
		}
		computedWeight := *s.Status.ComputedWeight
		weightToAdjust := *s.Spec.Weight - computedWeight

		if incrementalRollout(s.Spec.RolloutConfig) {
			rh.lastWriteLock.Lock() // Don't allow anyone else to read while we might write, avoids competing writes. todo: Would be nice to convert this to key based locking.
			lastSvcWrite := rh.lastWrite[serviceKey(s)]
			if time.Now().Before(lastSvcWrite.Add(time.Duration(s.Spec.RolloutConfig.IntervalSeconds) * time.Second)) {
				rh.lastWriteLock.Unlock()
				rh.enqueueService(s)
				continue // this protects the service from scaling early, can't trust that next enqueue is from here
			}
			if abs(weightToAdjust) < s.Spec.RolloutConfig.Increment || (weightToAdjust > 0 && allOtherServicesOff(s, svcs)) { // adjust entire amount
				computedWeight += weightToAdjust
			} else { // only adjust one increment
				oneIncrement := incrementFlux(s.Spec.RolloutConfig.Increment, *s.Spec.Weight, computedWeight)
				if weightToAdjust < 0 {
					oneIncrement = -oneIncrement
				}
				computedWeight += oneIncrement
			}
			*s.Status.ComputedWeight = computedWeight
			rh.lastWrite[serviceKey(s)] = metav1.NewTime(time.Now())
			rh.lastWriteLock.Unlock()
			rh.enqueueService(s)
		} else {
			// immediate rollout
			*s.Status.ComputedWeight += weightToAdjust
		}
		updatedNeeded = append(updatedNeeded, serviceKey(s))
	}
	err = rh.updateServices(svcs, updatedNeeded)
	if err != nil {
		return svc, err
	}
	return svc, nil
}

func (rh *rolloutHandler) updateServices(svcs []*riov1.Service, updateNeeded []string) error {
	for _, s := range svcs {
		if contains(updateNeeded, serviceKey(s)) {
			_, err := rh.client.UpdateStatus(s)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// sleep in background and run again after interval period
func (rh *rolloutHandler) enqueueService(s *riov1.Service) {
	rh.services.EnqueueAfter(s.Namespace, s.Name, time.Duration(s.Spec.RolloutConfig.IntervalSeconds)*time.Second)
}

func serviceKey(s *riov1.Service) string {
	return fmt.Sprintf("%s-%s-%s", s.Namespace, s.Spec.App, s.Name)
}

// Is pause true
func blocksRollout(rc *riov1.RolloutConfig) bool {
	return rc != nil && rc.Pause
}

// incrementalRollout returns whether we want to perform intervaled or immediate rollout
func incrementalRollout(rc *riov1.RolloutConfig) bool {
	return rc != nil && rc.Increment != 0 && rc.IntervalSeconds != 0
}

// Do any services have a ComutedWeight set ?
func computedWeightsExist(svcs []*riov1.Service) bool {
	for _, s := range svcs {
		if s.Status.ComputedWeight != nil {
			return true
		}
	}
	return false
}

// are all other services besides this service at zero or nil computedWeight ?
// Purpose: if we have svc-a at 60 and svc-b at 0, then just bump svc-a direct to 100%
func allOtherServicesOff(curr *riov1.Service, svcs []*riov1.Service) bool {
	for _, z := range svcs {
		if z.Name != curr.Name {
			if z.Status.ComputedWeight != nil && *z.Status.ComputedWeight > 0 {
				return false
			}
		}
	}
	return true
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Scale increment dynamically based on currWeight's percentage of goalWeight.
// Less than 50% currWeight from goalWeight is slower, greater is faster. And inverse when going down.
func incrementFlux(increment, goalWeight, currWeight int) int {
	if goalWeight < 100 {
		return increment // don't attempt to downscale promote's (goal is zero) or small changes
	}
	if currWeight < 1 {
		currWeight = 1
	}
	diff := float64(currWeight) / float64(goalWeight)
	var rate float64
	if currWeight > goalWeight {
		rate = diff - 0.5
	} else {
		rate = diff + 0.5
	}
	if rate > 1.25 {
		rate = 1.25 // max high rate is low to avoid small changes going too quickly
	}
	if rate < 0.5 {
		rate = 0.5 // slow early changes are good
	}
	return int(math.Floor(float64(increment) * rate))
}
