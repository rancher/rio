package services

import (
	"errors"
	"math"
	"time"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

const (
	DefaultInterval        = 4
	PromoteWeight          = 10000
	DefaultRolloutDuration = 60 * time.Second
)

func GenerateWeightAndRolloutConfig(svc *riov1.Service, svcs []*riov1.Service, targetPercentage int, duration time.Duration, pause bool) (int, *riov1.RolloutConfig, error) {
	if duration.Hours() > 10 {
		return 0, nil, errors.New("cannot perform rollout longer than 10 hours") // over 10 hours we go under increment of 1/10k, given 2 second. Also see safety valve below in increment.
	}
	if len(svcs) == 0 {
		return targetPercentage * 100, &riov1.RolloutConfig{}, nil
	}

	currComputedWeight := 0
	if svc.Status.ComputedWeight != nil && *svc.Status.ComputedWeight > 0 {
		currComputedWeight = *svc.Status.ComputedWeight
	}

	totalCurrWeight := 0
	for _, s := range svcs {
		if s.Status.ComputedWeight != nil && *s.Status.ComputedWeight > 0 {
			totalCurrWeight += *s.Status.ComputedWeight
		}
	}
	if targetPercentage == CalcWeightPercentage(currComputedWeight, totalCurrWeight) {
		return 0, nil, errors.New("cannot rollout, already at target percentage")
	}
	totalCurrWeightOtherSvcs := totalCurrWeight - currComputedWeight
	newComputedWeight := calcComputedWeight(targetPercentage, totalCurrWeightOtherSvcs)

	// if not immediate rollout figure out increment
	increment := 0
	if duration.Seconds() >= 2.0 {
		var err error
		increment, err = calcIncrement(duration, targetPercentage, totalCurrWeight, totalCurrWeightOtherSvcs)
		if err != nil {
			return 0, nil, err
		}
	}
	rolloutConfig := &riov1.RolloutConfig{
		Pause:           pause,
		Increment:       increment,
		IntervalSeconds: DefaultInterval,
	}
	return newComputedWeight, rolloutConfig, nil
}

// Get curr weight as percentage, rounded to nearest percent
func CalcWeightPercentage(weight, totalWeight int) int {
	if totalWeight == 0 || weight == 0 {
		return 0
	}
	return int(math.Round(float64(weight) / float64(totalWeight) / 0.01))
}

// Find the weight that would hit our target percentage without touching other service weights
// ie: if 2 svcs at 50/50 and you want one at 75%, newComputedWeight would be 150
func calcComputedWeight(targetPercentage int, totalCurrWeightOtherSvcs int) int {
	if targetPercentage == 100 {
		return PromoteWeight
	} else if totalCurrWeightOtherSvcs > 0 {
		return int(float64(totalCurrWeightOtherSvcs)/(1-(float64(targetPercentage)/100))) - totalCurrWeightOtherSvcs
	}
	return targetPercentage
}

// Determine increment we should step by given duration
// Note that we don't care (because blind to direction of scaling) if increment is larger than newComputedWeight, rollout controller will handle overflow case
func calcIncrement(duration time.Duration, targetPercentage, totalCurrWeight, totalCurrWeightOtherSvcs int) (int, error) {
	steps := duration.Seconds() / float64(DefaultInterval) // First get rough amount of steps we want to take
	if steps < 1.0 {
		steps = 1.0
	}
	newComputedWeight := calcComputedWeight(targetPercentage, totalCurrWeightOtherSvcs)
	totalNewWeight := totalCurrWeightOtherSvcs + newComputedWeight // Given the future total weight which includes our newWeight...
	difference := totalNewWeight - totalCurrWeight                 // Find the difference between future total weight and current total weight
	if targetPercentage == 100 {
		difference = PromoteWeight // In this case the future total weight is now only our newComputedWeight and difference is always 1k
	}
	if difference == 0 {
		return 0, nil // if there is no difference return now so we don't error out below
	}
	increment := int(math.Abs(math.Round(float64(difference) / steps))) // Divide by steps to get rough increment
	if increment == 0 {                                                 // Error out if increment was below 1, and thus rounded to 0
		return 0, errors.New("unable to perform rollout, given duration too long for current weight")
	}
	return increment, nil

}
