package integration

import (
	"github.com/rancher/rio/tests/testutil"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func domainTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService
	var domain testutil.TestDomain

	it.Before(func() {
		service.Create(t)
	})

	it.After(func() {
		service.Remove()
		domain.UnRegister()
	})

	when("a domain is created", func() {
		it("should exist with domain field correctly populated", func() {
			randomDomain := testutil.GenerateRandomDomain()
			domain.RegisterDomain(t, randomDomain, service.Name)
			assert.Equal(t, randomDomain, domain.GetDomain())
			//assert.Equal(t, randomDomain, domain.GetKubeDomain())
		})
	}, spec.Parallel())
}
