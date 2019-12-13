package integration

import (
	"os"
	"testing"

	"github.com/rancher/rio/tests/testutil"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestMain(m *testing.M) {
	testutil.IntegrationPreCheck()
	_, _ = testutil.KubectlCmd([]string{"create", "namespace", testutil.TestingNamespace})
	os.Exit(m.Run())
}

func TestSuite(t *testing.T) {
	suite := spec.New("integration suite", spec.Report(report.Terminal{}), spec.Parallel())
	specs := map[string]func(t *testing.T, when spec.G, it spec.S){
		"attach":          attachTests,
		"build":           buildTests,
		"config":          configTests,
		"domain":          domainTests,
		"export":          exportTests,
		"externalService": externalServiceTests,
		"log":             logTests,
		"rbac":            rbacTests,
		"riofile":         riofileTests,
		"route":           routeTests,
		"run":             runTests,
		"scale":           scaleTests,
		"stage":           stageTests,
		"weight":          weightTests,
	}
	for desc, fnc := range specs {
		suite(desc, fnc)
	}
	suite.Run(t)
}
