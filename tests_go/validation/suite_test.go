package validation

import (
	"os"
	"testing"

	"github.com/rancher/rio/tests_go/testutil"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestMain(m *testing.M) {
	testutil.PreCheck()
	os.Exit(m.Run())
}

func TestSuite(t *testing.T) {
	suite := map[string]func(t *testing.T, when spec.G, it spec.S){
		"rio logs":  logTests,
		"rio run":   runTests,
		"rio scale": scaleTests,
	}
	for desc, fnc := range suite {
		spec.Run(t, desc, fnc, spec.Report(report.Terminal{}))
	}
}
