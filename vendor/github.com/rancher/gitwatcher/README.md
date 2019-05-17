gitwatcher
========

## Sample

Note: scm configuration & git repository secrets management are beyond gitwatcher's scope.

Assume that githubconfig is set in proj-abc namespace and a git credential is store in user-def namespace(by configuring github in pipeline project settings)

Assume that rancher server will have a hook endpoint that proxy to gitwatcher(for testing change [here](https://github.com/gitlawr/gitwatcher/blob/prototype/pkg/utils/utils.go#L47))

1. Create a gitWebHookReceiver:
```
- apiVersion: gitwatcher.cattle.io/v1
  kind: GitWatcher
  metadata:
    name: test
    namespace: proj-abc
  spec:
    enabled: true
    pr: true
    provider: github
    repositoryCredentialSecretName: user-def:proj-abc-github-gitlawr
    repositoryUrl: https://github.com/gitlawr/test.git
    executionLabels:
      provider: github
```
gitWebHookReceiver controller will register a webhook in the repo.

2. Create a PR, an execution will be created by gitwatcher
```
- apiVersion: gitwatcher.cattle.io/v1
  kind: GitCommit
  metadata:
    generateName: test-
    name: test-n4fhv
    namespace: proj-abc
    labels:
      provider: github
  spec:
    author: gitlawr
    branch: master
    commit: 6d3f4956cff0
    pr: "1"
    gitWebHookReceiverName: proj-abc:test
    message: prmessage
    repositoryUrl: https://github.com/gitlawr/test.git
    sourceLink: https://github.com/gitlawr/test/pull/1
    title: prtitle
```

3. An external execution controller handle this execution, setting `Handled` condition & `statusUrl`
```
apiVersion: gitwatcher.cattle.io/v1
kind: GitCommit
metadata:
  generateName: test-
  name: test-n4fhv
  namespace: proj-abc
  labels:
    provider: github
spec:
  ...
status:
  conditions:
  - lastUpdateTime: 2018-12-06T14:15:43+08:00
    status: "Unknown"
    type: Handled
  statusUrl: http://myexample.com
```

4. gitwatcher will update build status on github according to `Handled` condition
```
apiVersion: gitwatcher.cattle.io/v1
kind: GitCommit
metadata:
  generateName: test-
  name: test-n4fhv
  namespace: proj-abc
  labels:
    provider: github
spec:
  ...
status:
  appliedStatus: "Unknown"
  conditions:
  - lastUpdateTime: 2018-12-06T14:15:43+08:00
    status: "Unknown"
    type: Handled
  statusUrl: http://myexample.com
```

## Building

`make`


## Running

`./bin/gitwatcher`

## License
Copyright (c) 2018 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
