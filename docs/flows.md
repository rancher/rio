# Deployment Flows

The following are the supported ways Rio can deploy a service and a service's future versions.
There are many options beyond what is displayed here, for instance deploying from a branch, rollout duration, or using rio build.

Functional params like port are omitted for brevity.

### Manual flows

#### Image-only

- Run initial service: `rio run -n demo imageV1`

- Action to update: `rio stage --image imageV2 demo v2`

- Action to shift traffic: `rio weight demo@v2=100`

#### Manual

todo

### Automatic flows

Flows below default to update-in-place. Use the ` --template` flag to create an additional service each time from the original service template.

Traffic will shift automatically, to disable this use `--template` and `--stage-only` params. See the cli-reference docs for more options, such as rollout duration.

#### Build from source on commit - polling or webhook

- Run initial service: `rio run -n demo https://github.com/rancher/rio-demo`

- Action to update: Add a new commit to repo, for example merge a PR

#### Build from source on tag - webhook-only

- Run initial service: `rio run -n demo --build-tag --build-webhook-secret=secret https://github.com/rancher/rio-demo`

- Action to update: Create a tag

### Additional flows

#### Build-PR: spin up a new service for each PR opened on the repo.

- Run initial service: `rio run -n demo --build-pr --template --build-webhook-secret=mysecret https://github.com/rancher/rio-demo`

- Action to update: Create a new PR and a new service will be spun up. On PR merge or close, the PR service is removed.
