package rollout

import (
	"testing"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func TestWeightAdjustOne(t *testing.T) {
	versions := []string{"v0", "v1"}
	weights := []int{0, 0}
	result := map[string]riov1.ServiceObservedWeight{
		"v0": {
			Weight: 50,
		},
		"v1": {
			Weight: 50,
		},
	}
	for i := 0; i < 20; i++ {
		magicSteal(versions, weights, result, -5)
	}

	if result["v0"].Weight != 0 {
		t.Fatalf("v0 weight should be 0 after adjusting, result: %+v", result)
	}
	if result["v1"].Weight != 0 {
		t.Fatalf("v0 weight should be 0 after adjusting, result: %+v", result)
	}
}

func TestWeightAdjustTwo(t *testing.T) {
	versions := []string{"v0", "v1", "v2"}
	weights := []int{50, 30, 20}
	result := map[string]riov1.ServiceObservedWeight{
		"v0": {
			Weight: 40,
		},
		"v1": {
			Weight: 20,
		},
		"v2": {
			Weight: 10,
		},
	}
	for i := 0; i < 3; i++ {
		magicSteal(versions, weights, result, 10)
	}

	if result["v0"].Weight != 50 {
		t.Fatalf("v0 weight should be 50 after adjusting, result: %+v", result)
	}
	if result["v1"].Weight != 30 {
		t.Fatalf("v1 weight should be 30 after adjusting, result: %+v", result)
	}
	if result["v2"].Weight != 20 {
		t.Fatalf("v2 weight should be 20 after adjusting, result: %+v", result)
	}
}

func TestWeightAdjustThree(t *testing.T) {
	versions := []string{"v0", "v1"}
	weights := []int{100, 0}
	result := map[string]riov1.ServiceObservedWeight{
		"v0": {
			Weight: 0,
		},
		"v1": {
			Weight: 0,
		},
	}
	magicSteal(versions, weights, result, -100)
	if result["v0"].Weight != 0 {
		t.Fatalf("v0 weight should be 0 after adjusting, result: %v", result["v0"].Weight)
	}
	if result["v1"].Weight != 0 {
		t.Fatalf("v1 weight should be 0 after adjusting, result: %v", result["v1"].Weight)
	}
}
