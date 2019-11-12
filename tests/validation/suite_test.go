package validation

import (
	"os"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"github.com/rancher/rio/tests/testutil"
)

func TestMain(m *testing.M) {
	testutil.ValidationPreCheck()
	_, _ = testutil.KubectlCmd([]string{"create", "namespace", testutil.TestingNamespace})
	os.Exit(m.Run())
}

func TestSuite(t *testing.T) {
	suite := spec.New("validation suite", spec.Report(report.Terminal{}))
	specs := map[string]func(t *testing.T, when spec.G, it spec.S){
		"autoscale":   autoscaleTests,
		"domainTests": domainTests,
		"scale":       scaleTests,
		"weight":      weightTests,
	}
	for desc, fnc := range specs {
		suite(desc, fnc)
	}
	suite.Run(t)
}
