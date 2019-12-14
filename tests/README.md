# Rio testing

All commands run from root dir

- [Unit Tests](#unit-tests)
- [Integration Tests](#integration-tests)
- [Validation Tests](#validation-tests)
- [Full Suite](#full-suite)
- [Running in Docker](#running-in-docker)


### Unit Tests

```
go test -race -cover -tags=test ./...
```


### Integration Tests

Integration tests live inside the /tests/integration dir. They require a working cluster with KUBECONFIG env var set, and rio already installed.

```
go test -v -race ./tests/integration/... -integration-tests
```

Every spec should be in the `when x, it should y` format.


### Validation Tests

Validation tests live inside the /tests/validation dir. They require a working cluster with KUBECONFIG env var set, and rio already installed.
In the future, rio will be installed as part of the tests.

```
go test -v -race ./tests/validation/... -validation-tests
```

These tests should not be used to block merging PRs. They are meant to run in other CI and run less regularly. They are also the place where more obscure tests that are useful for regression can be written.

### Integration/Validation Dev Notes

* Logic lives in the testutil dir, test specs live in the integration or validation folder
* We are purposefully failing tests in the util code rather than returning errors in order to keep specs clean
    * Only fail tests in public methods
* To add a new suite, create a file and add it in the TestSuite list
* While writing use `it.Focus` or `when.Focus` to limit to your spec
* To help debug, try printing args in `RioCmd` to see what the tests are doing
* Most every `It` block spins up a new service which takes > 10 seconds. If you want to specify multiple assertions on a single service it will be much faster to use testify descriptions within a single block, example in run_test.go
* Ensure all `Remove` functions can be run on a never-instantiated object, you don't know what order will be in `it.After`.  
* Make sure you use local variables when testing in parallel [to avoid bugs](https://gist.github.com/posener/92a55c4cd441fc5e5e85f27bca008721).

Goland is useful for debugging, setup with:

* test kind: `package`
* package: `github.com/rancher/rio/tests/integration`
* environment: set your `KUBECONFIG`. `DEBUG_TEST` will print extra `rio` commands that were invoked.
* program arguments: `--integration-tests`

### Full suite

```
make ci
```

### Running in Docker
Integration and Validation tests can be run with docker:
1. Add a KUBECONFIG file to /path/to/rio/.kube/config
2. Make sure you have docker installed, then run the following in a terminal:

```bash
$ cd /path/to/rio
$ docker build -f "./tests/Dockerfile" -t rio-test .

# To run integration tests
$ docker run --rm -e test=integration rio-test:latest
# To run validation tests:
$ docker run --rm -e test=integration -e AWS_ACCESS_KEY_ID=[keyid] -e AWS_SECRET_ACCESS_KEY=[secret] -e RIO_ROUTE53_ZONEID=[zoneid] -e RIO_ROUTE53_ZONENAME=[your.zone.name] rio-test:latest

# To run with a specific version of RIO installed. Tests are always optimized for the latest version.
$ docker run --rm -e test=integration -e INSTALL_RIO_VERSION=v0.6.0-rc2 rio-test:latest
```


### Fuzzing

We fuzz inputs to the system using [go-fuzz](https://github.com/dvyukov/go-fuzz), see that for install instructions and how to deal with crashes.

```
cd tests/fuzz/parser
go-fuzz-build
go-fuzz -bin=./riofile-fuzz.zip
```
That will run forever. With a large enough corpus and a long enough run we should find any issues.
