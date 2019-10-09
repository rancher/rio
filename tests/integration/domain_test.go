package integration

import (
	"testing"

	"github.com/rancher/rio/tests/testutil"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
)

func domainTests(t *testing.T, when spec.G, it spec.S) {

	var service testutil.TestService
	var domain testutil.TestDomain

	it.Before(func() {
		service.Create(t, "nginx")
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
		})
	}, spec.Parallel())
}
