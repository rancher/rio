Rio
===

## Rio is currently being massively overhauled. Expect to see something around May/June.

1. Simple, fun, end-to-end container experience
1. Cloud Native Container Distribution

Rio is a user oriented end-to-end container solution with a focus on keeping containers simple and
combating the current trend of complexity. It's kept fun and simple through it's familiar and
opinionated user experience.  Additionally, Rio is a "Cloud Native Container Distribution"
meaning it includes builtin Cloud Native technologies such as Kubernetes, Istio, Containerd, etc.
so that the user need not be an expert in installing, using, and maintaining these systems.

[![Rio Demo](https://img.youtube.com/vi/8YkIycwad2w/0.jpg)](https://www.youtube.com/watch?v=8YkIycwad2w)


## Current Status: Early Preview

This is an early preview, features may be broken, not work as described, and has been known to be irresistibly drawn
to large cities, where it will back up sewers, reverse street signs, and steal everyone's left shoe.
Please try it out and file bugs.

### Goals

1. Fun. Containers should be fun.
1. Simple. Simple can only be achieved by applying some opinion and as such Rio is an opinionated tool.
1. Portable. Each Rio cluster should have the same functionality available to it.  The differences between clusters
are only speed, reliability, and permissions.  Running in production should be just as simple as your laptop (and vice versa)
1. Secure. Rio will by default use the best security settings and encryption and such will be enabled by default.
1. Product Grade. Running Rio should give you a production worthy system and not need you to bolt on infinite more ops tools.
1. Cloud Native Distribution.  Rio will include all the the key cloud native technologies by default such that each user does not
need to be an expert in the details.

## Quick Start

1. Have a Kubernetes cluster running. Rio can be running in any kubernetes cluster that supports ingress. (https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/)
   For example, if you want to use [k3s](https://k3s.io/), run `curl -sfL https://get.k3s.io | sh -`.
2. Download the latest Rio CLI release from the [release page](https://github.com/rancher/rio/releases), Linux, Mac and Windows binaries are provided.
3. Set up KUBECONFIG variable. `export KUBECONFIG=/path/to/kubeconfig`.
4. Run `rio install`. 

You should be able to see logs like this
```
Deploying Rio control plane....
Rio control plane is deployed. Run `kubectl -n rio-system describe deploy rio-controller` to get more detail.
Welcome to Rio!
```
Waiting for all the management pods to be up and running. `kubectl get po -n rio-system`

Done! Now try [an example](./README.md#rio-stage-options-service_id_name).

## Installation

Download: [Linux, Mac, Windows](https://github.com/rancher/rio/releases)

Run `rio install`.

Follow the onscreen prompts and Rio will try to install itself into the current `kubectl` cluster.  Please note `cluster-admin`
privileges are required for Rio.  This will probably changes, but for now we need the world.

All the prerequisite that Rio needs is just a KUBECONFIG file. Rio itself Contains the CLI and management plane controller 
which install CRD and start controller in the current cluster. By default if you install rio in your current cluster you will
automatically get a DNS record registered for your ingress IPs. If you run `rio info` you should be able to see the domain.

```
» rio info                                                                                   
Rio Version: dev
Cluster Domain: xxxxx.on-rio.io
System Namespace: rio-system
```

Using Rio
=========

## Concepts

### Service

The main unit that is being dealt with in Rio are services.  Services are just a collection of containers that provide a
similar function.  When you run containers in Rio you are really creating a Scalable Service.  `rio run` and `rio create` will
create a service.  You can later scale that service with `rio scale`.  Services are assigned a DNS name so that group
of containers can be accessed from other services.

### Apps

App contains multiple service revisions. Each service is rio can uniquely be identified as a revision, group by app and version.
App aggregates all the revisions by app name, provides an entry to access all the revisions. What percentage of traffic goes to which 
revision depends on how much weight is set on each revision.

### Service Mesh

Rio has a built in service mesh, powered by Istio and Envoy.  The service mesh provides all of the core communication
abilities for services to talk to each other, inbound traffic and outbound traffic.  All traffic can be encrypted,
validated, and routed dynamically according to the configuration.  Rio specifically does not require the user to
understand much about the underlying service mesh.  Just know that all communication is going through the service mesh.

### Router

Router is an abstract layer sitting on top of services, it provides a configuration to different services. It can define various 
routing rules to match different backends. Router also provides an entry to access and a Dns name in the cluster.

### ExternalService

ExternalService provides a way to create dns record for services that are outside of mesh. It can be IP addresses and FQDN.

## Basics

For each of these command you can run `rio --help` to get all the available options.

### rio run [OPTIONS] IMAGE [COMMAND] [ARG...]

Run a scalable service with given options.  There's a lot of options, run `rio run --help`

### rio create [OPTIONS] IMAGE [COMMAND] [ARG...]

The same as run but create a service with scale=0.  To start the service afterwards
run `rio scale`

### rio ps [OPTIONS] [STACK...]

List the running services or containers.  By default `rio ps` will show services.  To
view the individual containers backing the service run `rio ps -c` or you can run
`rio ps myservice` to list the containers backing a specific service.

### rio scale [SERVICE=NUMBER...]

Scale a service up or down. You can pass as many services as you wish for example

    rio scale myservice=3 otherstack/myservice2=1

### rio rm [ID_OR_NAME...]

Delete a resource.  `rio rm` will delete most any resource by ID or name excepts nodes.
If the name matches multiple resources the CLI will ask you which specific one to delete.
You can use IDs and the `--type` option to narrow down to delete specific things and not use
fuzzy matching.

### rio inspect [ID_OR_NAME...]

Return the raw json API response of the object.  You can use `--format` to change
to yaml or format the output using go formatting.

## Rio Files

Services in Rio can be imported, exported and dynamically edited. The syntax of the rio files
is an extension of the docker-compose format.  We wish to be backwards compatible with
docker-compose where feasible.  This means Rio should be able to run a docker-compose file, but
a Rio stack file will not run in docker-compose as we are only backwards compatible.  

Rio files has three parts: Services, Config and raw Kubernetes manifest. Below is an example of more complex stack 
file that is used to deploy build compenents.

```yaml
configs:
  logging:
    content: |-
      loglevel.controller: info
        loglevel.creds-init: info
        loglevel.git-init: info
        loglevel.webhook: info
        zap-logger-config: |
          {
            "level": "info",
            "development": false,
            "sampling": {
              "initial": 100,
              "thereafter": 100
            },
            "outputPaths": ["stdout"],
            "errorOutputPaths": ["stderr"],
            "encoding": "json",
            "encoderConfig": {
              "timeKey": "",
              "levelKey": "level",
              "nameKey": "logger",
              "callerKey": "caller",
              "messageKey": "msg",
              "stacktraceKey": "stacktrace",
              "lineEnding": "",
              "levelEncoder": "",
              "timeEncoder": "",
              "durationEncoder": "",
              "callerEncoder": ""
            }
          }

services:
  buildkit:
    maxScale: 10
    minScale: 1
    concurrency: 5
    ports:
    - 9001/tcp,buildkit,internal=true
    systemSpec:
      podSpec:
        containers:
        - image: moby/buildkit:v0.3.3
          args:
          - --addr
          - tcp://0.0.0.0:9001
          name: buildkitd
          ports:
          - containerPort: 9001
          securityContext:
            privileged: true
  registry:
    image: registry:2
    labels:
      request-subdomain: 'true'
    secrets:
    - rio-wildcard:/etc/registry
    env:
    - REGISTRY_HTTP_ADDR=0.0.0.0:443
    - REGISTRY_HTTP_TLS_CERTIFICATE=/etc/registry/tls.crt
    - REGISTRY_HTTP_TLS_KEY=/etc/registry/tls.key
    ports:
    - 443:443/tcp,registry
    volumes:
    - storage-registry:/var/lib/registry
  webhook:
    global_permissions:
    - "* webhookinator.rio.cattle.io/gitwebhookreceivers"
    - "* webhookinator.rio.cattle.io/gitwebhookexecutions"
    - '* configmaps'
    - '* events'
    - secrets
    image: daishan1992/webhookinator:dev
    args:
    - webhookinator
    - --listen-address
    - :8090
    imagePullPolicy: always
    ports:
    - 8090/tcp,http-webhookinator
  build-controller:
    global_permissions:
    - '* pods'
    - '* namespaces'
    - '* secrets'
    - '* events'
    - '* serviceaccounts'
    - '* configmaps'
    - '* extentions/deployments'
    - 'create,get,list,watch,patch,update,delete build.knative.dev/builds'
    - 'create,get,list,watch,patch,update,delete build.knative.dev/builds/status'
    - 'create,get,list,watch,patch,update,delete build.knative.dev/buildtemplates'
    - 'create,get,list,watch,patch,update,delete build.knative.dev/clusterbuildtemplates'
    - '* caching.internal.knative.dev/images'
    - '* apiextensions.k8s.io/customresourcedefinitions'
    image: gcr.io/knative-releases/github.com/knative/build/cmd/controller@sha256:77b883fec7820bd3219c011796f552f15572a037895fbe7a7c78c7328fd96187
    configs:
    - logging/content:/etc/config-logging
    env:
    - SYSTEM_NAMESPACE=${NAMESPACE}
    args:
    - -logtostderr
    - -stderrthreshold
    - INFO
    - -creds-image
    - gcr.io/knative-releases/github.com/knative/build/cmd/creds-init@sha256:ebf58f848c65c50a7158a155db7e0384c3430221564c4bbaf83e8fbde8f756fe
    - -git-image
    - gcr.io/knative-releases/github.com/knative/build/cmd/git-init@sha256:09f22919256ba4f7451e4e595227fb852b0a55e5e1e4694cb7df5ba0ad742b23
    - -nop-image
    - gcr.io/knative-releases/github.com/knative/build/cmd/nop@sha256:a318ee728d516ff732e2861c02ddf86197e52c6288049695781acb7710c841d4


kubernetes:
  manifest: |-
    apiVersion: apiextensions.k8s.io/v1beta1
    kind: CustomResourceDefinition
    metadata:
      labels:
        knative.dev/crd-install: "true"
      name: builds.build.knative.dev
    spec:
      additionalPrinterColumns:
      - JSONPath: .status.conditions[?(@.type=="Succeeded")].status
        name: Succeeded
        type: string
      - JSONPath: .status.conditions[?(@.type=="Succeeded")].reason
        name: Reason
        type: string
      - JSONPath: .status.startTime
        name: StartTime
        type: date
      - JSONPath: .status.completionTime
        name: CompletionTime
        type: date
      group: build.knative.dev
      names:
        categories:
        - all
        - knative
        kind: Build
        plural: builds
      scope: Namespaced
      subresources:
        status: {}
      version: v1alpha1
    ---
    apiVersion: caching.internal.knative.dev/v1alpha1
    kind: Image
    metadata:
      name: creds-init
      namespace: ${NAMESPACE}
    spec:
      image: gcr.io/knative-releases/github.com/knative/build/cmd/creds-init@sha256:ebf58f848c65c50a7158a155db7e0384c3430221564c4bbaf83e8fbde8f756fe
    ---
    apiVersion: caching.internal.knative.dev/v1alpha1
    kind: Image
    metadata:
      name: git-init
      namespace: ${NAMESPACE}
    spec:
      image: gcr.io/knative-releases/github.com/knative/build/cmd/git-init@sha256:09f22919256ba4f7451e4e595227fb852b0a55e5e1e4694cb7df5ba0ad742b23
    ---
    apiVersion: caching.internal.knative.dev/v1alpha1
    kind: Image
    metadata:
      name: gcs-fetcher
      namespace: ${NAMESPACE}
    spec:
      image: gcr.io/cloud-builders/gcs-fetcher
    ---
    apiVersion: caching.internal.knative.dev/v1alpha1
    kind: Image
    metadata:
      name: nop
      namespace: ${NAMESPACE}
    spec:
      image: gcr.io/knative-releases/github.com/knative/build/cmd/nop@sha256:a318ee728d516ff732e2861c02ddf86197e52c6288049695781acb7710c841d4
    ---
    apiVersion: build.knative.dev/v1alpha1
    kind: ClusterBuildTemplate
    metadata:
      name: buildkit
    spec:
      parameters:
      - name: IMAGE
        description: Where to publish the resulting image
      - name: DOCKERFILE
        description: The name of the Dockerfile
        default: "Dockerfile"
      - name: PUSH
        description: Whether push or not
        default: "true"
      - name: DIRECTORY
        description: The directory containing the app
        default: "/workspace"
      - name: BUILDKIT_CLIENT_IMAGE
        description: The name of the BuildKit client (buildctl) image
        default: "moby/buildkit:v0.3.1-rootless@sha256:2407cc7f24e154a7b699979c7ced886805cac67920169dcebcca9166493ee2b6"
      - name: BUILDKIT_DAEMON_ADDRESS
        description: The address of the BuildKit daemon (buildkitd) service
        default: "tcp://buildkitd:1234"
      steps:
      - name: build-and-push
        image: $${BUILDKIT_CLIENT_IMAGE}
        workingDir: $${DIRECTORY}
        command: ["buildctl", "--addr=$${BUILDKIT_DAEMON_ADDRESS}", "build",
                  "--progress=plain",
                  "--frontend=dockerfile.v0",
                  "--frontend-opt", "filename=$${DOCKERFILE}",
                  "--local", "context=.", "--local", "dockerfile=.",
                  "--exporter=image", "--exporter-opt", "name=$${IMAGE}", "--exporter-opt", "push=$${PUSH}"]

```

### rio export SERVICE_NAME

Export a specific service.  This will print the service to standard out.  You can pipe the out
of the export command to a file using the shell, for example `rio export myservice > rio-stack.yml`

### rio apply -f FILE|-

Import a stack from file or standard in.

1. Create a rio file with the following content.

```yaml
services:
  demo:
    image: ibuildthecloud/demo:v1
    ports:
    - 80/http
```

2. Run the following scripts 
```bash
# Create stack foo from standard input
cat service.yml | rio apply -f -

# see if the service 
rio ps
```

Apply a remote url
```bash
$ rio apply -f https://raw.githubusercontent.com/StrongMonkey/demo/master/demo-stack.yaml
INFO[0000] Deploying rio-file to namespace [default] from https://raw.githubusercontent.com/StrongMonkey/demo/master/demo-stack.yaml
$ rio revision 
NAME              IMAGE                    CREATED              STATE     SCALE     ENDPOINT                                   WEIGHT                               DETAIL
default/demo:v0   ibuildthecloud/demo:v1   About a minute ago   active    1         https://demo-v0-default.8axlxl.on-rio.io   =============================> 100   
```

### rio edit $NAME

Edit a specific service and run `rio up` with the new contents. Only works for services right now

### Questions

When running `up` stack files can prompt the user for questions.  To define a question add questions
to your stack file as follows

```yaml
services:
  foo:
    environment:
      VAR: ${BLAH}
    image: nginx
    
template:
  questions:
  - variable: BLAH
    description: "You should answer something good"
```

The values of the questions can be references anywhere in the stack file using
${..} syntax.  The following bash style variables are supported (using [github.com/drone/envsubst](http://github.com/drone/envsubst))

```
${var^}
${var^^}
${var,}
${var,,}
${var:position}
${var:position:length}
${var#substring}
${var##substring}
${var%substring}
${var%%substring}
${var/substring/replacement}
${var//substring/replacement}
${var/#substring/replacement}
${var/%substring/replacement}
${#var}
${var=default}
${var:=default}
${var:-default}
```

Questions have the following fields

| Name                | Description |
| ------------------- | ----------- |
| variable            | The variable name to reference using ${...} syntax |
| label               | A friend name for the question |
| description         | A longer description of the question |
| type                | The field type: string, int, bool, enum.  default is string |
| required            | The answer can not be blank |
| default             | Default value of the answer if not specified by the user |
| group               | Group the question with questions in the same group (Most used by UI) |
| min_length          | Minimum length of the answer |
| max_length          | Maximum length of the answer |
| min                 | Minimum value of an int answer |
| max                 | Maximum value of an int answer |
| options             | An array of valid answers for type enum questions |
| valid_chars         | Answer must be composed of only these characters |
| invalid_chars       | Answer must not have any of these characters |
| subquestions        | A list of questions that are considered child questions |
| show_if             | Ask question only if this evaluates to true, more info on syntax below |
| show_subquestion_if | Ask subquestions if this evaluates to true |

For `showIf` and `showSubquestionsIf` the syntax is `VARIABLE=VALUE [&& VARIABLE=VALUE]`.  For example

```
questions:
- variable: STORAGE
  description: Do you want to use persistent storage
  type: bool
- variable: STORAGE_TYPE
  type: enum
  options:
  - aws
  - local
- variable: SIZE
  description: Size of the volume to create
  show_if: STORAGE
- variable: AWS_KEY
  description: Enter AWS API key
  show_if: STORAGE && STORAGE_TYPE=aws
```

### Templating

All stack files are go templates so any go templating can be used.  Please remember that
heavy use of templating makes the stack files hard to read so use conservatively.  Also,
all stack files must render with all empty variable.

`Values` is put into the template context as a map of all variable values, for example `{{ if eq .Values.STORAGE "true" }}`.
Regardless of the type of the question, the `Values` map will always contain strings.

## Interacting with Services

### rio exec [OPTIONS] CONTAINER COMMAND [ARG...]

Launch a new command in a service or container

### rio attach [OPTIONS] CONTAINER

Attach to an existing running process in a service or container

### rio logs [OPTIONS] [CONTAINER_OR_SERVICE...]

Get or tail logs of a service or container

## Service Config Files

In a stack you can define the contents of files that can then be injected into containers. For example

```
configs:
  index:
    content: "<h1>hi</h1>"
    
services:
  nginx:
    image: nginx
    ports:
    - 80/http
    configs:
    - index:/usr/share/nginx/html/index.html

```

Configs can be defined using a string or base64 format, using
either the contents or encoded keys for string or base64
respectively.

```
configs:
  index:
    # String format
    content: "<h1>hi</h1>"
  index2:
    # base64, for binary data
    encoded: PGgxPmhpPC9oMT4K
```

### rio cat [NAME...]

Echo to standard out the contents of the reference config

### rio config ls

List all configs

### rio config create NAME FILE|-

Create a config of the given NAME from FILE or from standard input if `-` is passed.

### rio config update NAME FILE|-

Update a config of the given NAME from FILE or from standard input if `-` is passed.

### rio edit

The standard `rio edit` command can edit configs also

## Cluster Domain and TLS

By default Rio will create a DNS record pointing to your ingress IPs. Rio also uses Letsencrypt to create
a wildcard certificate for the cluster domain so that all the traffic to access your application can be encrypted.
For example, When you deploy your workload, you can access your workload in HTTPS. The domain always follow the format
of ${app}-${namespace}.${cluster-domain}. You can see your cluster domain by running `rio info`.


```bash
# See cluster info
$ rio info
Rio Version: dev
Cluster Domain: 8axlxl.on-rio.io
System Namespace: rio-system

# Run your workload
$ rio run -p 80/http --name svc --scale=3 ibuildthecloud/demo:v1

# See the endpoint of your workload 
$ rio ps
NAME          ENDPOINT                               SCALE     WEIGHT
default/svc   https://svc-default.8axlxl.on-rio.io   v0/3      v0/100%

### Access your workload
$ curl https://svc-default.8axlxl.on-rio.io
Hello World
```

## Service Mesh

### Staging Versions

Service mesh will route traffic to given services.  Services can have multiple versions of
the service deployed at once and then the you can control how much traffic, or which traffic
is routed to each version.

### rio stage [OPTIONS] SERVICE_ID_NAME

The `rio stage` command takes all the same options as `rio create` but instead of updating
the existing service, it will stage a new version of the service.  For example, below is
scenario to do a canary deployment.

```bash

# Create a new service
$ rio run -p 80/http --name svc --scale=3 ibuildthecloud/demo:v1

# Ensure service is running and determine public URL
$ rio revision default/svc
AME              IMAGE                    CREATED          STATE     SCALE     ENDPOINT                                     WEIGHT                               DETAIL
default/svc:v0   ibuildthecloud/demo:v1   16 minutes ago   active    3         https://svc-v0-default.8axlxl.on-rio.io      =============================> 100   

# Stage new version, updating just the docker image and assigning it to "v3" version.
$ rio stage --image=ibuildthecloud/demo:v3 default/svc:v3 

# To change the spec of new service
$ rio stage --edit test/svc:v3

# Notice a new URL was created for your staged service
$ rio revision default/svc
NAME             IMAGE                    CREATED          STATE     SCALE     ENDPOINT                                     WEIGHT                               DETAIL
default/svc:v0   ibuildthecloud/demo:v1   16 minutes ago   active    3         https://svc-v0-default.8axlxl.on-rio.io      =============================> 100   
default/svc:v3   ibuildthecloud/demo:v3   24 seconds ago   active    3         https://svc-v3-default.8axlxl.on-rio.io 

# Access current revision
$ curl -s https://svc-v0-default.8axlxl.on-rio.io
Hello World

# Access staged service under new URL
$ curl -s https://svc-v3-default.8axlxl.on-rio.io
Hello World v3

# Show access url for all the revision
$ rio ps
NAME          ENDPOINT                               SCALE        WEIGHT
default/svc   https://svc-default.8axlxl.on-rio.io   v0/3; v3/3   v0/100%; v3/0%

# Access the app(stands for all the revision). Note that right now there is no traffic to v3.
$ curl -s https://svc-default.8axlxl.on-rio.io
Hello World

# Promote v3 service. The traffic will be shifted to v3 gradually. By default we apply 5% shift every 5 seconds, but it can be configred
# using flags `--rollout-increment` and `--rollout-interval`. To turn off rollout(traffic percentage will be changed to
# the desired value immediately), run `--no-rollout`.
$ rio promote default/svc:v3
NAME             IMAGE                    CREATED          STATE     SCALE     ENDPOINT                                  WEIGHT                         DETAIL
default/svc:v0   ibuildthecloud/demo:v1   37 minutes ago   active    3         https://svc-v0-default.8axlxl.on-rio.io   ========================> 85   
default/svc:v3   ibuildthecloud/demo:v3   21 minutes ago   active    3         https://svc-v3-default.8axlxl.on-rio.io   ===> 15

# Access the app. You should be able to see traffic routing to the new revision
$ curl https://svc-default.8axlxl.on-rio.io
Hello World
$ curl https://svc-default.8axlxl.on-rio.io
Hello World v3

# Wait for v3 to be 100% weight. Access the app, all traffic should be routed to new revision right now.
$ rio revision default/svc
NAME             IMAGE                    CREATED          STATE     SCALE     ENDPOINT                                  WEIGHT                               DETAIL
default/svc:v0   ibuildthecloud/demo:v1   42 minutes ago   active    3         https://svc-v0-default.8axlxl.on-rio.io                                        
default/svc:v3   ibuildthecloud/demo:v3   26 minutes ago   active    3         https://svc-v3-default.8axlxl.on-rio.io   =============================> 100  
$ curl https://svc-default.8axlxl.on-rio.io
Hello World v3

# Adjust weight
$ rio weight default/svc:v0=5% default/svc:v3=95%
NAME             IMAGE                    CREATED          STATE     SCALE     ENDPOINT                                  WEIGHT                            DETAIL
default/svc:v0   ibuildthecloud/demo:v1   44 minutes ago   active    3         https://svc-v0-default.8axlxl.on-rio.io   > 5                               
default/svc:v3   ibuildthecloud/demo:v3   27 minutes ago   active    3         https://svc-v3-default.8axlxl.on-rio.io   ===========================> 95   

```

### rio stage [OPTIONS] SERVICE_ID_NAME

```bash
# Export to see v0 service
$ rio export default/svc:v0
kubernetes:
  type: kubernetes
services:
  svc:
    cpus: "0"
    image: ibuildthecloud/demo:v1
    imagePullPolicy: IfNotPresent
    ports:
    - "80"
    rollout: true
    rolloutIncrement: 5
    rolloutInterval: 5
    scale: 3
    type: service
    weight: 95
type: riofile
```

## Autoscaling

By default rio will enable autoscaling for workloads. Depends on Qps and Current active requests on your workload,
Rio will scale the workload to the proper scale.

```bash
# Run a workload, set minimal scale and maximum scale
$ rio run -p 8080/http --name autoscale --scale=1-20 strongmonkey1992/autoscale:v0 
default/autoscale

# Put some load to the workload. We use tool [hey](https://github.com/rakyll/hey) to put loads.
$ hey -z 600s -c 60 https://autoscale-default.8axlxl.on-rio.io

# Noted that service has been scaled to 6
$ rio revision default/autoscale
NAME                   IMAGE                           CREATED         STATE     SCALE     ENDPOINT                                        WEIGHT                               DETAIL
default/autoscale:v0   strongmonkey1992/autoscale:v0   4 minutes ago   active    6         https://autoscale-v0-default.8axlxl.on-rio.io   =============================> 100

# Run a workload that can be scaled to zero
$ rio run -p 8080/http --name autoscale-zero --scale=0-20 strongmonkey1992/autoscale:v0

# Wait for a couple of minutes. The workload is scaled to zero.
NAME                        IMAGE                           CREATED         STATE     SCALE       ENDPOINT   WEIGHT                               DETAIL
default/autoscale-zero:v0   strongmonkey1992/autoscale:v0   4 minutes ago   pending   (0/0/1)/0              =============================> 100  

# Access the workload. Once there is an active request, workload can be re-scaled to active.
$ rio ps 
NAME                     ENDPOINT                                          SCALE           WEIGHT
default/autoscale-zero   https://autoscale-zero-default.8axlxl.on-rio.io   v0/(0/0/0)/1    v0/100%
$ curl -s https://autoscale-zero-default.8axlxl.on-rio.io
Hi there, I am StrongMonkey:v13

# Workload is re-scaled to 1
$ rio revision default/autoscale-zero
NAME                        IMAGE                           CREATED          STATE     SCALE     ENDPOINT                                             WEIGHT                               DETAIL
default/autoscale-zero:v0   strongmonkey1992/autoscale:v0   18 minutes ago   active    1         https://autoscale-zero-v0-default.8axlxl.on-rio.io   =============================> 100  
```

## Source code to Deployment

Rio supports configure a git-based source code repository to deploy the actual workload. It can be as easy
as giving Rio a valid git repository repo. 

```bash
# Run a workload from a git repo. We assume the repo has a Dockerfile at root directory to build the image
$ rio run -p 8080/http -n build https://github.com/StrongMonkey/demo.git
default/build

# Waiting for the image to be built. Note the image column is empty. Once the image is ready service will be active
$ rio revision
NAME               IMAGE     CREATED          STATE      SCALE     ENDPOINT                                    WEIGHT                               DETAIL
default/build:v0             27 seconds ago   inactive   1         https://build-v0-default.8axlxl.on-rio.io   =============================> 100   

# Image is ready. Noted that we deploy the default docker registry into the cluster. 
# The image name has the format of ${registry-domain}/${namespace}/${name}:${commit} 
$ rio revision
NAME               IMAGE                                                                                         CREATED         STATE     SCALE     ENDPOINT                                    WEIGHT                               DETAIL
default/build:v0   registry-rio-system.8axlxl.on-rio.io/default/build:34512dddba18781fb6909c303eb206a73d41d9ba   2 minutes ago   active    1         https://build-v0-default.8axlxl.on-rio.io   =============================> 100   

# Show the endpoint of your workload
$ rio ps 
NAME            ENDPOINT                                 SCALE     WEIGHT
default/build   https://build-default.8axlxl.on-rio.io   v0/1      v0/100%

# Access the endpoint
$ curl -s https://build-default.8axlxl.on-rio.io
Hi there, I am StrongMonkey:v1
```

When you point your workload to a git repo, Rio will automatically watch any commit or tag pushed to
a specific branch(default is master). By default Rio will pull and check the branch at a certain interval, but this
can be configured to use a webhook.

```bash
# edit the code, change v1 to v3, push the code
$ vim main.go | git add -u | git commit -m "change to v3" | git push $remote

# A new revision has been automatically created. Noticed that once the new revision is created, the traffic will
# automatically shifted from old revision to new revision.
$ rio revision default/build
NAME                  IMAGE                                                                                                       CREATED          STATE     SCALE     ENDPOINT                                       WEIGHT                   DETAIL
default/build:v0      registry-rio-system.8axlxl.on-rio.io/default/build:34512dddba18781fb6909c303eb206a73d41d9ba                 20 minutes ago   active    1         https://build-v0-default.8axlxl.on-rio.io      ==================> 65   
default/build:25a0a   registry-rio-system.8axlxl.on-rio.io/default/build-e46cfb4-08a3b:25a0acda54812619f8063c121f6ed5ed2bfb968f   50 seconds ago   active    1         https://build-25a0a-default.8axlxl.on-rio.io   =========> 35    

# Access the endpoint
$ curl https://build-default.8axlxl.on-rio.io
Hi there, I am StrongMonkey:v1
$ curl https://build-default.8axlxl.on-rio.io
Hi there, I am StrongMonkey:v3

# Wait for all the traffic are shifted to the new revision, 
$ rio revision default/build
NAME                  IMAGE                                                                                                       CREATED          STATE     SCALE     ENDPOINT                                       WEIGHT                               DETAIL
default/build:v0      registry-rio-system.8axlxl.on-rio.io/default/build:34512dddba18781fb6909c303eb206a73d41d9ba                 24 minutes ago   active    1         https://build-v0-default.8axlxl.on-rio.io                                           
default/build:25a0a   registry-rio-system.8axlxl.on-rio.io/default/build-e46cfb4-08a3b:25a0acda54812619f8063c121f6ed5ed2bfb968f   4 minutes ago    active    1         https://build-25a0a-default.8axlxl.on-rio.io   =============================> 100

# Access the workload. Noted that all the traffic are routed to the new revision
$ curl https://build-default.8axlxl.on-rio.io
Hi there, I am StrongMonkey:v3
```

## Monitoring

By default Rio will deploy [grafana](https://grafana.com/) and [kiali](https://www.kiali.io/) to give user abilities to watch all metrics corresponding to service mesh.

```bash
# Monitoring services are deployed into rio-system namespace
$ rio --system ps 
NAME                                ENDPOINT                                       SCALE     WEIGHT
rio-system/autoscaler                                                              v0/1      v0/100%
rio-system/build-controller                                                        v0/1      v0/100%
rio-system/buildkit                                                                v0/1      v0/100%
rio-system/cert-manager                                                            v0/1      v0/100%
rio-system/grafana                  https://grafana-rio-system.8axlxl.on-rio.io    v0/1      v0/100%
rio-system/istio-citadel                                                           v0/1      v0/100%
rio-system/istio-gateway                                                           v0/1      v0/100%
rio-system/istio-pilot                                                             v0/1      v0/100%
rio-system/istio-telemetry                                                         v0/1      v0/100%
rio-system/kiali                    https://kiali-rio-system.8axlxl.on-rio.io      v0/1      v0/100%
rio-system/local-path-provisioner                                                  v0/1      v0/100%
rio-system/prometheus                                                              v0/1      v0/100%
rio-system/registry                 https://registry-rio-system.8axlxl.on-rio.io   v0/1      v0/100%
rio-system/webhook                  https://webhook-rio-system.8axlxl.on-rio.io    v0/1      v0/100%
```










## Roadmap

| Function | Implementation | Status |
|----------|----------------|--------|
| Container Runtime | containerd | included
| Orchestration | Kubernetes | included
| Networking | Flannel | included
| Service Mesh | Istio | included
| Monitoring | Prometheus | included
| Logging | Fluentd
| Storage | Longhorn |
| CI | Drone
| Registry | Docker Registry 2 | included
| Builder | Moby BuildKit | included
| TLS | Let's Encrypt | included
| Image Scanning | Clair
| Autoscaling | Knative

## License
Copyright (c) 2014 - 2019 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
