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

### Prerequisites

1. A supported Virtual Machine provider
    1. [VirtualBox](https://www.virtualbox.org/wiki/Downloads) is recommended and default
    1. [VMware Fusion](https://www.vmware.com/products/fusion.html) for Mac users performs better (in theory)
1. [Vagrant 1.6+](https://www.vagrantup.com/downloads.html)

Checkout a copy of Rio:
```
$ git clone https://github.com/rancher/rio.git
$ cd rio
```

Configure [vagrant.yaml](./vagrant.yaml) for your desired VM provider, if not VirtualBox. You may also change the number of nodes and in your Rio cluster and the resources allocated to each node.

Run `vagrant up` from the project root directory. You may be asked to select a bridged network interface; select the interface being used to connect to the internet. Depending on your host OS and directory permissions, you may be asked to authenticate during `rio` client installation on the host.

Run `rio -v` to ensure rio is installed on your host machine. Run `rio ps` to ensure the client is authenticated with the Rio cluster.
```
$ rio -v
rio version v0.0.3
$ rio ps
NAME      IMAGE     CREATED   SCALE     STATE     ENDPOINT   DETAIL
```

Done! Now try [an example](./README.md#rio-stage-options-service_id_name).

## Installation

Download: [Linux, Mac, Windows](https://github.com/rancher/rio/releases)

Rio will run in two different modes:

**Rio Standalone**: In this mode Rio comes with all the container tech you need built in. All you need are modern Linux
 servers.  Rio does not need Docker, Kubernetes or anything else installed on the host.  This mode is good if you want
 to run containers but don't want to be a Kubernetes expert.  Rio will ensure that you have the most secure setup and
 keep all the components up to date.  This is by far the easiest way to run clusters.

**Run on Kubernetes**: In this mode Rio will use an existing Kubernetes cluster.  The advantages of this approach is
that you get more flexibility in terms of networking, storage, and other components at the cost of greatly increased
complexity.  For the time being this mode is also good for your laptop as Minikube and Docker For Mac/Windows both
provide a simple way to run Kubernetes on your laptop.  In the future Rio will have a mode that is simpler and does not
require Docker For Mac/Windows or Minikube.

### Standalone

Standalone requires Linux 4.x+ that support overlay, squashfs, and containers in general.
This will be most current distributions.

Rio forms a cluster.  To create the cluster you need one or more servers and one or more agents.  Right now HA is in the works still
so only one server is supported.  To start a server run

    sudo rio server

That will start the server and register the current host as a node in the cluster.  At this point you have a full single node
cluster.  If you don't wish to use the current server as a node then run

    rio server --disable-agent

This mode does have the benefit of not requiring root privileges.  On startup the server will print something similar as below

```

INFO[0005] To use CLI: rio login -s https://10.20.0.3:7443 -t R108527fc31eb165d69e4ebb048168769d97734707dc22bd197b5ae2fcab27d9e64::admin:fb5ef140c22562de2789168ac6973bda 
INFO[0005] To join node to cluster: rio agent -s https://10.20.0.3:7443 -t R108527fc31eb165d69e4ebb048168769d97734707dc22bd197b5ae2fcab27d9e64::node:9cb35d8ae4a4621abdacfa6d8d1ea1b6 

```

Use those two command to either access the server from the CLI or add another node to the cluster.  If you are root
on the host that is running the Rio server, `rio login` is not required.

The state of the server will be in `/var/lib/rancher/rio/server` or `${HOME}/.rancher/rio/server` if running as non-root.
For more robust HA setups that state can be moved to MySQL or etcd (this is still in the works).  The state of the agent
will be in `/var/lib/rancher/rio/agent`.

### On Kubernetes

If you wish to run on an existing Kubernetes cluster all that is requires is that you have a working `kubectl` setup.  Then
just run

    rio login

Follow the onscreen prompts and Rio will try to install itself into the current `kubectl` cluster.  Please note `cluster-admin`
privileges are required for Rio.  This will probably changes, but for now we need the world.

Using Rio
=========

## Concepts

### Service

The main unit that is being dealt with in Rio are services.  Services are just a collection of containers that provide a
similar function.  When you run containers in Rio you are really creating a Scalable Service.  `rio run` and `rio create` will
create a service.  You can later scale that service with `rio scale`.  Services are assigned a DNS name so that group
of containers can be accessed from other services.

### Stack

A stack is a group of services and their related resources, such as configuration files, volumes and routes.  A stack
ends up typically representing one application.  All the names of services are unique within a stack, but not globally
unique.  This means a stack creates a scope for service discovery.  Under the hood a stack will use a Kubernetes
namespace.

### Project

A project is a collection of stacks, and other resources such as secrets. The `rio` command line runs commands within
a single project.  Using `rio --project PROJECT` you can point to a different project.  Stack names are unique
within a project only.  As the permissions model of Rio matures the project will be the primary unit that is used
for collaboration.  Users are invited to and given access to projects.

### Service Mesh

Rio has a built in service mesh, powered by Istio and Envoy.  The service mesh provides all of the core communication
abilities for services to talk to each other, inbound traffic and outbound traffic.  All traffic can be encrypted,
validated, and routed dynamically according to the configuration.  Rio specifically does not require the user to
understand much about the underlying service mesh.  Just know that all communication is going through the service mesh.

## Basics

For each of these command you can run `rio cmd --help` to get all the available options.

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

## Stack Files

Stacks in Rio can be imported, exported and dynamically edited.  The syntax of the stack files
is an extension of the docker-compose format.  We wish to be backwards compatible with
docker-compose where feasible.  This means Rio should be able to run a docker-compose file, but
a Rio stack file will not run in docker-compose as we are only backwards compatible.  Below is an
example of more complex stack file that is used to deploy istio

```yaml
configs:
  mesh:
    content: |-
      disablePolicyChecks: true
      ingressControllerMode: "OFF"
      authPolicy: NONE
      rdsRefreshDelay: 10s
      outboundTrafficPolicy:
        mode: ALLOW_ANY
      defaultConfig:
        discoveryRefreshDelay: 10s
        connectTimeout: 30s
        configPath: "/etc/istio/proxy"
        binaryPath: "/usr/local/bin/envoy"
        serviceCluster: istio-proxy
        drainDuration: 45s
        parentShutdownDuration: 1m0s
        interceptionMode: REDIRECT
        proxyAdminPort: 15000
        controlPlaneAuthPolicy: NONE
        discoveryAddress: istio-pilot.${NAMESPACE}:15007

services:
  istio-pilot:
    command: discovery
    configs:
    - mesh:/etc/istio/config/mesh
    environment:
    - POD_NAME=$(self/name)
    - POD_NAMESPACE=$(self/namespace)
    - PILOT_THROTTLE=500
    - PILOT_CACHE_SQUASH=5
    global_permissions:
    - '* config.istio.io/*'
    - '* networking.istio.io/*'
    - '* authentication.istio.io/*'
    - '* apiextensions.k8s.io/customresourcedefinitions'
    - '* extensions/thirdpartyresources'
    - '* extensions/thirdpartyresources.extensions'
    - '* extensions/ingresses'
    - '* extensions/ingresses/status'
    - create,get,list,watch,update configmaps
    - endpoints
    - pods
    - services
    - namespaces
    - nodes
    - secrets
    image: istio/pilot:0.8.0
    secrets: identity:/etc/certs
    sidekicks:
      istio-proxy:
        expose:
        - 15007/http
        - 15010/grpc
        image: istio/proxyv2:0.8.0
        command:
        - proxy
        - --serviceCluster
        - istio-pilot
        - --templateFile
        - /etc/istio/proxy/envoy_pilot.yaml.tmpl
        - --controlPlaneAuthPolicy
        - NONE
        environment:
        - POD_NAME=$(self/name)
        - POD_NAMESPACE=$(self/namespace)
        - INSTANCE_IP=$(self/ip)
        secrets: identity:/etc/certs

  istio-citadel:
    image: "istio/citadel:0.8.0"
    command:
    - --append-dns-names=true
    - --grpc-port=8060
    - --grpc-hostname=citadel
    - --self-signed-ca=true
    - --citadel-storage-namespace=istio-system
    global_permissions:
    - write secrets
    - serviceaccounts
    - services
    permissions:
    - get,delete deployments
    - get,delete serviceaccounts
    - get,delete services
    - get,delete deployments
    - get,list,update,delete extensions/deployments
    - get,list,update,delete extensions/replicasets
    secrets: identity:/etc/certs

  istio-gateway:
    labels:
      "gateway": "external"
    image: "istio/proxyv2:0.8.0"
    net: host
    dns: cluster
    command:
    - proxy
    - router
    - -v
    - "2"
    - --discoveryRefreshDelay
    - '1s' #discoveryRefreshDelay
    - --drainDuration
    - '45s' #drainDuration
    - --parentShutdownDuration
    - '1m0s' #parentShutdownDuration
    - --connectTimeout
    - '10s' #connectTimeout
    - --serviceCluster
    - istio-proxy
    - --zipkinAddress
    - ""
    - --statsdUdpAddress
    - ""
    - --proxyAdminPort
    - "15000"
    - --controlPlaneAuthPolicy
    - NONE
    - --discoveryAddress
    - istio-pilot:15007
    env:
    - POD_NAME=$(self/name)
    - POD_NAMESPACE=$(self/namespace)
    - INSTANCE_IP=$(self/ip)
    - ISTIO_META_POD_NAME=$(self/name)
    secrets: identity:/etc/certs
    global_permissions:
    - "get,watch,list,update extensions/thirdpartyresources"
    - "get,watch,list,update */virtualservices"
    - "get,watch,list,update */destinationrules"
    - "get,watch,list,update */gateways"

```

### rio export STACK_ID_OR_NAME

Export a specific stack.  This will print the stack to standard out.  You can pipe the out
of the export command to a file using the shell, for example `rio export mystack > stack.yml`

### rio up [OPTIONS] [[STACK_NAME] FILE|-]

Import a stack from file or standard in.  The `rio up` can be ran in different forms

```bash
# Create stack foo from standard input
cat stack.yml | rio up foo -

# Create stack foo from file
rio up foo stack.yml

# Run up for all files in the current directory matching *-stack.yml.  The portion before
# -stack.yml will be used as the stack name
rio up
```

### rio edit ID_OR_NAME

Edit a specific stack and run `rio up` with the new contents.

### Questions

When running `up` stack files can prompt the user for questions.  To define a question add questions
to your stack file as follows

```yaml
services:
  foo:
    environment:
      VAR: ${BLAH}
    image: nginx
    
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
$ rio run -p 80/http --name test/svc --scale=3 ibuildthecloud/demo:v1

# Ensure service is running and determine public URL
$ rio ps
NAME       IMAGE                    CREATED          SCALE     STATE     ENDPOINT                                  DETAIL
test/svc   ibuildthecloud/demo:v1   17 seconds ago   3         active    http://svc.test.8gr18g.lb.rancher.cloud   

# Stage new version, updating just the docker image and assigning it to "v3" version.
$ rio stage --image=ibuildthecloud/demo:v3 test/svc:v3

# Notice a new URL was created for your staged service
$ rio ps
NAME          IMAGE                    CREATED        SCALE     STATE     ENDPOINT                                     DETAIL
test/svc      ibuildthecloud/demo:v1   10 hours ago   3         active    http://svc.test.8gr18g.lb.rancher.cloud      
test/svc:v3   ibuildthecloud/demo:v3   10 hours ago   3         active    http://svc-v3.test.8gr18g.lb.rancher.cloud   

# Access current service
$ curl -s http://svc.test.8gr18g.lb.rancher.cloud
Hello World

# Access staged service under new URL
$ curl -s http://svc-v3.test.8gr18g.lb.rancher.cloud
Hello World v3

# Export to see stack file format
$ rio export test
services:
  svc:
    image: ibuildthecloud/demo:v1
    ports:
    - 80/http
    revisions:
      v3:
        image: ibuildthecloud/demo:v3
        scale: 3
    scale: 3

# Send some production traffic to new version
$ rio weight test/svc:v3=50%

# See that 50% of traffic goes to new service
$ curl -s http://svc.test.8gr18g.lb.rancher.cloud
Hello World
$ curl -s http://svc.test.8gr18g.lb.rancher.cloud
Hello World v3

# Happy with the new version we promote the stage version to be the primary
$ rio promote test/svc:v3

# All new traffic is v3
$ curl -s http://svc.test.8gr18g.lb.rancher.cloud
Hello World v3
$ curl -s http://svc.test.8gr18g.lb.rancher.cloud
Hello World v3

```

## Roadmap

| Function | Implementation | Status |
|----------|----------------|--------|
| Container Runtime | containerd | included
| Orchestration | Kubernetes | included
| Networking | Flannel | included
| Service Mesh | Istio | included
| Monitoring | Prometheus
| Logging | Fluentd
| Storage | Longhorn |
| CI | Drone
| Registry | Docker Registry 2
| Builder | Moby BuildKit
| TLS | Let's Encrypt
| Image Scanning | Clair

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
