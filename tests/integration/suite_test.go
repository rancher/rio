package integration

import (
	"os"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"github.com/rancher/rio/tests/testutil"
)

func TestMain(m *testing.M) {
	testutil.PreCheck()
	os.Exit(m.Run())
}

func TestSuite(t *testing.T) {
	suite := spec.New("my suite", spec.Report(report.Terminal{}))
	specs := map[string]func(t *testing.T, when spec.G, it spec.S){
		"run":             runTests,
		"scale":           scaleTests,
		"weight":          weightTests,
		"endpoint":        endpointTests,
		"domain":          domainTests,
		"route":           routeTests,
		"config":          configTests,
		"export":          exportTests,
		"externalService": externalServiceTests,
	}
	for desc, fnc := range specs {
		suite(desc, fnc)
	}
	suite.Run(t)
}
