// +build validation

package validation

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestSuite(t *testing.T) {
	suite := map[string]func(t *testing.T, when spec.G, it spec.S){
		"rio logs":  logTests,
		"rio run":   runTests,
		"rio scale": scaleTests,
	}
	for desc, fnc := range suite {
		spec.Run(t, desc, fnc, spec.Report(report.Log{}))
	}
}
