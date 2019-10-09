# Rio testing

All commands run from root dir

- [Unit tests](#unit-tests)
- [Integration tests](#integration-tests)
- [Full Suite](#full-suite)
- [Notes](#notes)


### Unit tests

```
go test -race -cover -tags=test ./...
```


### Integration tests

Integration tests live inside the /tests/integration dir. They require a working cluster with KUBECONFIG env var set, and rio already installed.

```
go test -v -race ./tests/integration/... -integration-tests
```

Every spec should be in the `when x, it should y` format.

Dev Notes

* Logic lives in the testutil dir, test specs live in the integration folder
* We are purposefully failing tests in the util code rather than returning errors in order to keep specs clean
    * Only fail tests in public methods. Public methods should not call each other.
* To add a new suite, create a file and add it in the TestSuite list
* While writing use `it.Focus` or `when.Focus` to limit to your spec
* Use parallel tests as much as possible
* Most every `It` block spins up a new service which takes > 10 seconds. If you want to specify multiple assertions on a single service it will be much faster to use testify descriptions within a single block, example in run_test.go
* Ensure all `Remove` functions can be run on a never-instantiated object, you don't know what order will be in `it.After`.  
* Make sure you use local variables when testing in parallel [to avoid bugs](https://gist.github.com/posener/92a55c4cd441fc5e5e85f27bca008721).

Goland is useful for debugging, setup with:

* test kind: `package`
* package: `github.com/rancher/rio/tests/integration`
* environment: set your `KUBECONFIG`
* program arguments: `--integration-tests`


### Full suite

```
make ci
```


