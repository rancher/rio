package validation

import (
	"fmt"
	"os"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"github.com/rancher/rio/tests/testutil"
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	err := testutil.ValidationPreCheck()
	if err != nil {
		fmt.Println(err)
		return 0
	}
	testutil.CreateNS()
	return m.Run()
}

func TestSuite(t *testing.T) {
	suite := spec.New("validation suite", spec.Report(report.Terminal{}))
	specs := map[string]func(t *testing.T, when spec.G, it spec.S){
		"autoscale": autoscaleTests,
		"domain":    domainTests,
		"exec":      execTests,
		"scale":     scaleTests,
		"weight":    weightTests,
	}
	for desc, fnc := range specs {
		suite(desc, fnc)
	}
	suite.Run(t)
}
