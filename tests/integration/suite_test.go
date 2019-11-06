package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/rancher/rio/tests/testutil"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var context *testutil.TestContext

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	err := testutil.IntegrationPreCheck()
	if err != nil {
		fmt.Println(err)
		return 0
	}
	context, err = testutil.NewTestContext()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	defer context.Cleanup()
	return m.Run()
}

func TestSuite(t *testing.T) {
	suite := spec.New("integration suite", spec.Report(report.Terminal{}), spec.Parallel())
	specs := map[string]func(t *testing.T, when spec.G, it spec.S){
		"attach":          attachTests,
		"autoscale":       autoscaleTests,
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
