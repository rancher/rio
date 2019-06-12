# Rio

Rio is a MicroPaaS that can be layered on top of any standard Kubernetes cluster. Consisting of a few Kubernetes custom resources and a CLI to enhance the user experience, users can easily deploy services to Kubernetes and automatically get continuous delivery, DNS, HTTPS, routing, monitoring, autoscaling, canary deployments, git-triggered builds, and much more. All it takes to get going is an existing Kubernetes cluster and the rio CLI.

## Quick Start

1. Have a Kubernetes 1.13 or newer cluster running.

   [k3s](https://k3s.io/), [RKE](https://github.com/rancher/rke), [Minikube](https://kubernetes.io/docs/setup/minikube/), [Docker For Mac Edge](https://docs.docker.com/docker-for-mac/edge-release-notes/), [GKE](https://cloud.google.com/kubernetes-engine/), [AKS](https://docs.microsoft.com/en-us/azure/aks/), [EKS](https://aws.amazon.com/eks/),

   Please ensure you have at least 3GB of memory free in your cluster. We will attempt to reduce the memory footprint in a future release. Some of the components we are currently depending on are quite large.

2. Run

```bash
# Download the CLI (available for macOS, Windows, Linux)
$ curl -sfL https://get.rio.io | sh -   # or manually from https://github.com/rancher/rio/releases

# Setup your cluster for Rio
$ rio install

# Make sure all the pods are up and running. These takes several minutes.
$ kubectl get po -n rio-system

# Run a sample service
$ rio run https://github.com/rancher/rio-demo

# Check the status
$ rio ps
$ rio console
$ rio info
```

Note: Rio will use a [service loadbalancer](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer) to expose the service mesh gateway.
If your cluster doesn't support service load balancers, simply run `rio install --host-ports` to use host ports to expose gateway.

If your host has multiple IP addresses, you can specify which IP address Rio should use for creating external DNS records with the `--ipaddress` flag. For instance to advertise the external IP of an AWS instance: `rio install --ipaddress $(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)`

On start up, you can specify which feature to disable when installing rio. Simplely run `rio install --disable-features autoscaling,build --disable-features letsencrypt`


| Feature | Description |
|----------|----------------|
| autoscaling | Auto-scaling services based on QPS and requests load
| build | Rio Build, from source code to deployment
| grafana | Grafana Dashboard
| istio | Service routing using Istio
| kiali | Kiali Dashboard
| letsencrypt | Let's Encrypt
| mixer | Istio Mixer telemetry
| prometheus | Enable prometheus
| rdns | Assign cluster a hostname from public Rancher DNS service

# Using Rio

## Concepts

### Service

The main unit that is being dealt with in Rio are services. Services are just a scalable set of containers that provide a
similar function. When you run containers in Rio you are really creating a Service. `rio run` and `rio create` will
create a service. You can later scale that service with `rio scale`. Services are assigned a DNS name so that group
of containers can be accessed from other services.

### Apps

An App contains multiple service revisions. Each service in Rio is assigned an app and version. Services that have the same app but
different versions are reference to as revisions. The group of all revisions for an app is what is called an App or application in Rio.
An application named `foo` will be given a DNS name like `foo.clusterdomain.on-rio.io` and each version is assigned it's own DNS name. If the app was
`foo` and the version is `v2` the assigned DNS name for that revision would be similar to `foo-v2.clusterdomain.on-rio.io`. `rio ps` and `rio revision` will
list the assigned DNS names.

### Router

Router is a virtual service that load balances and routes traffic to other services. Routing rules can route based
on hostname, path, HTTP headers, protocol, and source.

### External Service

External Service provides a way to register external IPs or hostnames in the service mesh so they can be accessed by Rio services.

### Public Domain

Public Domain can be configured to assign a service or router a vanity domain like www.myproductionsite.com.

## Service Mesh

Rio has a built in service mesh, powered by Istio and Envoy. The service mesh provides all of the core communication
abilities for services to talk to each other, inbound traffic and outbound traffic. All traffic can be encrypted,
validated, and routed dynamically according to the configuration. Rio specifically does not require the user to
understand much about the underlying service mesh.

## Cluster Domain and TLS

By default Rio will create a DNS record pointing to your cluster. Rio also uses Let's Encrypt to create
a certificate for the cluster domain so that all services support HTTPS by default.
For example, when you deploy your workload, you can access your workload in HTTPS. The domain always follows the format
of ${app}-${namespace}.\${cluster-domain}. You can see your cluster domain by running `rio info`.

```bash
# Run your workload
$ rio run -p 80/http --name svc --scale=3 ibuildthecloud/demo:v1
default/svc:v0

# See the endpoint of your workload
$ rio ps
NAME          ENDPOINT                                    REVISIONS   SCALE     WEIGHT
default/svc   https://svc-default.iazlia.on-rio.io:9443   v0          0/3       100%

### Access your workload
$ curl https://svc-default.iazlia.on-rio.io:9443
Hello World
```

## Service Mesh

### Staging Versions

Services can have multiple versions deployed at once. The service mesh can then decide how much traffic
to route to each revision.

### rio stage [--image=IMAGE][--edit] SERVICE

The `rio stage` command will stage a new revision of an existing service.
the existing service, it will stage a new version of the service. For example, below is
scenario to do a canary deployment.

```bash

# Create a new service
$ rio run -p 80/http --name svc --scale=3 ibuildthecloud/demo:v1
default/svc:v0

# Ensure the service is running and determine its public URL
$ rio revision default/svc
NAME             IMAGE                    CREATED          STATE     SCALE     ENDPOINT                                       WEIGHT                               DETAIL
default/svc:v0   ibuildthecloud/demo:v1   14 seconds ago   active    3         https://svc-v0-default.iazlia.on-rio.io:9443   =============================> 100


# Stage a new version, updating just the docker image and assigning it to "v3" version.
$ rio stage --image=ibuildthecloud/demo:v3 default/svc:v3
default/svc:v3

# Change the spec of the new service
$ rio stage --edit default/svc:v3

# Notice a new URL was created for your staged service
$ rio revision default/svc
NAME             IMAGE                    CREATED              STATE     SCALE     ENDPOINT                                       WEIGHT                               DETAIL
default/svc:v0   ibuildthecloud/demo:v1   About a minute ago   active    3         https://svc-v0-default.iazlia.on-rio.io:9443   =============================> 100
default/svc:v3   ibuildthecloud/demo:v3   49 seconds ago       active    3         https://svc-v3-default.iazlia.on-rio.io:9443

# Access the current revision
$ curl -s https://svc-v0-default.iazlia.on-rio.io:9443
Hello World

# Access the staged service under the new URL
$ curl -s https://svc-v3-default.iazlia.on-rio.io:9443
Hello World v3

# Show the access url for all the revisions
$ rio ps
NAME          ENDPOINT                                    REVISIONS   SCALE     WEIGHT
default/svc   https://svc-default.iazlia.on-rio.io:9443   v0,v3       3,3       100%,0%

# Access the app(stands for all the revision). Note that right now there is no traffic to v3.
$ curl https://svc-default.iazlia.on-rio.io:9443
Hello World

# Promote v3 service. The traffic will be shifted to v3 gradually. By default we apply a 5% shift every 5 seconds, but it can be configured
# using the flags `--rollout-increment` and `--rollout-interval`. To turn off rollout(the traffic percentage will be changed to
# the desired value immediately), run `--no-rollout`.
$ rio promote default/svc:v3

$ rio revision default/svc
NAME             IMAGE                    CREATED         STATE     SCALE     ENDPOINT                                       WEIGHT                   DETAIL
default/svc:v0   ibuildthecloud/demo:v1   3 minutes ago   active    3         https://svc-v0-default.iazlia.on-rio.io:9443   ==================> 65
default/svc:v3   ibuildthecloud/demo:v3   2 minutes ago   active    3         https://svc-v3-default.iazlia.on-rio.io:9443   =========> 35

# Access the app. You should be able to see traffic routing to the new revision
$ curl https://svc-default.iazlia.on-rio.io:9443
Hello World

$ curl https://svc-default.iazlia.on-rio.io:9443
Hello World v3

# Wait for v3 to be 100% weight. Access the app, all traffic should be routed to new revision right now.
$ rio revision default/svc
NAME             IMAGE                    CREATED         STATE     SCALE     ENDPOINT                                       WEIGHT                               DETAIL
default/svc:v0   ibuildthecloud/demo:v1   5 minutes ago   active    3         https://svc-v0-default.iazlia.on-rio.io:9443
default/svc:v3   ibuildthecloud/demo:v3   4 minutes ago   active    3         https://svc-v3-default.iazlia.on-rio.io:9443   =============================> 100

$ curl https://svc-default.iazlia.on-rio.io:9443
Hello World v3

# Adjust weight
$ rio weight default/svc:v0=5% default/svc:v3=95%

$ rio ps
NAME          ENDPOINT                                    REVISIONS   SCALE     WEIGHT
default/svc   https://svc-default.iazlia.on-rio.io:9443   v0,v3       3,3       5%,95%
```

### rio route

`rio route` allows you to create a router that contains routing rules to different workloads.

```base
# Create a route to point to svc:v0 and svc:v3
$ rio route append route1/to-svc-v0 to default/svc:v0
$ rio route append route1/to-svc-v3 to default/svc:v3

# Access the route
$ rio route
NAME             URL                                                      OPTS         ACTION    TARGET
default/route1   https://route1-default.iazlia.on-rio.io:9443/to-svc-v0   timeout=0s   to        svc:v0,port=80
default/route1   https://route1-default.iazlia.on-rio.io:9443/to-svc-v3   timeout=0s   to        svc:v3,port=80

$ curl -s https://route1-default.iazlia.on-rio.io:9443/to-svc-v0
Hello World

$ curl -s https://route1-default.iazlia.on-rio.io:9443/to-svc-v3
Hello World v3
```

### rio externalservice

`rio externalservice` allows you to create dns record for external services that are outside the service mesh

```bash
# Create an external service pointing to an IP
$ rio externalservice create external 1.1.1.1

#  Create an external service pointing to an FQDN
$ rio externalservice create external-fqdn my.app.com

$ rio external
NAME                    CREATED         TARGET
default/external        3 minutes ago   1.1.1.1
default/external-fqdn   3 seconds ago   my.app.com
```

### rio domain

`rio domain` allows you to create your own domain pointing to a specific service or route

```bash
# Create a domain that points to route1. You have to setup a cname record from your domain to cluster domain.
# For example, foo.bar -> CNAME -> iazlia.on-rio.io
$ rio domain add foo.bar default/route1
default/foo-bar

# Use your own certs by providing a secret that contain tls cert and key instead of provisioning by letsencrypts. The secret has to be created first in system namespace.
$ rio domain add --secret $name foo.bar default/route1
```

Note: By default Rio will automatically configure Letsencrypt HTTP-01 challenge to provision certs for your publicdomain. This need you to install rio on standard ports.
Try `rio install --httpport 80 --httpsport 443`.

## Autoscaling

By default, Rio enables autoscaling for workloads. Depends on QPS and current active requests on your workload,
Rio scales the workload to the proper scale.

```bash
# Run a workload, set the minimal and maximum scale
$ rio run -p 8080/http --name autoscale --scale=1-20 strongmonkey1992/autoscale:v0
default/autoscale:v0

# Put some load to the workload. We use [hey](https://github.com/rakyll/hey) to create traffic
$ hey -z 600s -c 60 http://autoscale-default.iazlia.on-rio.io:9080

# Note that the service has been scaled to 6 instances
$ rio revision default/autoscale
NAME                   IMAGE                           CREATED          STATE     SCALE     ENDPOINT                                             WEIGHT                               DETAIL
default/autoscale:v0   strongmonkey1992/autoscale:v0   40 seconds ago   active    6         https://autoscale-v0-default.iazlia.on-rio.io:9443   =============================> 100


# Run a workload that can be scaled to zero
$ rio run -p 8080/http --name autoscale-zero --scale=0-20 strongmonkey1992/autoscale:v0
default/autoscale-zero:v0

# Wait a couple of minutes for the workload to scale to zero
$ rio revision default/autoscale-zero
NAME                        IMAGE                           CREATED              STATE     SCALE     ENDPOINT                                                  WEIGHT                               DETAIL
default/autoscale-zero:v0   strongmonkey1992/autoscale:v0   About a minute ago   active    0         https://autoscale-zero-v0-default.iazlia.on-rio.io:9443   =============================> 100

# Access the workload. Once there is an active request, the workload will be re-scaled to active.
$ rio ps
NAME                     ENDPOINT                                               REVISIONS   SCALE     WEIGHT
default/autoscale-zero   https://autoscale-zero-default.iazlia.on-rio.io:9443   v0          0/1       100%

$ curl -s https://autoscale-zero-default.iazlia.on-rio.io:9443
Hi there, I am StrongMonkey:v13

# Verify that the workload has been re-scaled to 1
$ rio revision default/autoscale-zero
NAME                     ENDPOINT                                               REVISIONS   SCALE     WEIGHT
default/autoscale-zero   https://autoscale-zero-default.iazlia.on-rio.io:9443   v0          1         100%
```

## Source code to Deployment

Rio supports configuration of a Git-based source code repository to deploy the actual workload. It can be as easy
as giving Rio a valid Git repository repo URL.

```bash
# Run a workload from a git repo. We assume the repo has a Dockerfile at root directory to build the image
$ rio run -p 8080/http -n build https://github.com/StrongMonkey/demo.git
default/build:v0

# Waiting for the image to be built. Note the image column is empty. Once the image is ready service will be active
$ rio revision
NAME               IMAGE     CREATED          STATE      SCALE     ENDPOINT                                         WEIGHT                               DETAIL
default/build:v0             29 seconds ago   inactive   1         https://build-v0-default.iazlia.on-rio.io:9443   =============================> 100

# The image is ready. Note that we deploy from the default docker registry into the cluster.
# The image name has the format of ${registry-domain}/${namespace}/${name}:${commit}
$ rio revision
NAME               IMAGE                                                                                         CREATED              STATE     SCALE     ENDPOINT                                         WEIGHT                               DETAIL
default/build:v0   registry-rio-system.iazlia.on-rio.io/default/build:32a4e453ca3bf0672ece9abf6901fa307d951add   About a minute ago   active    0/1       https://build-v0-default.iazlia.on-rio.io:9443   =============================> 100


# Show the endpoint of your workload
$ rio ps
NAME            ENDPOINT                                      REVISIONS   SCALE     WEIGHT
default/build   https://build-default.iazlia.on-rio.io:9443   v0          1         100%

# Access the endpoint
$ curl -s https://build-default.iazlia.on-rio.io:9443
Hi there, I am StrongMonkey:v1
```

When you point your workload to a git repo, Rio will automatically watch any commit or tag pushed to
a specific branch (default is master). By default, Rio will pull and check the branch at a certain interval, but
can be configured to use a webhook instead.

```bash
# edit the code, change v1 to v3, push the code
$ vim main.go | git add -u | git commit -m "change to v3" | git push $remote

# A new revision has been automatically created. Noticed that once the new revision is created, the traffic will
# be automatically shifted from the old revision to the new revision.
$ rio revision default/build
NAME                   IMAGE                                                                                                       CREATED          STATE     SCALE     ENDPOINT                                             WEIGHT                               DETAIL
default/build:v0       registry-rio-system.iazlia.on-rio.io/default/build:32a4e453ca3bf0672ece9abf6901fa307d951add                 11 minutes ago   active    1         https://build-v0-default.iazlia.on-rio.io:9443
default/build:vc6d4c   registry-rio-system.iazlia.on-rio.io/default/build-e46cfb4-1d207:c6d4c4452b064e476940de7b33c7a70ac0d9e153   22 seconds ago   active    1         https://build-vc6d4c-default.iazlia.on-rio.io:9443   =============================> 100

# Access the endpoint
$ curl https://build-default.8axlxl.on-rio.io
Hi there, I am StrongMonkey:v1
$ curl https://build-default.8axlxl.on-rio.io
Hi there, I am StrongMonkey:v3

# Wait until all traffic has been shifted to the new revision
$ rio revision default/build
NAME                  IMAGE                                                                                                       CREATED          STATE     SCALE     ENDPOINT                                       WEIGHT                               DETAIL
default/build:v0      registry-rio-system.8axlxl.on-rio.io/default/build:34512dddba18781fb6909c303eb206a73d41d9ba                 24 minutes ago   active    1         https://build-v0-default.8axlxl.on-rio.io
default/build:25a0a   registry-rio-system.8axlxl.on-rio.io/default/build-e46cfb4-08a3b:25a0acda54812619f8063c121f6ed5ed2bfb968f   4 minutes ago    active    1         https://build-25a0a-default.8axlxl.on-rio.io   =============================> 100

# Access the workload. Note that all the traffic is routed to the new revision
$ curl https://build-default.8axlxl.on-rio.io
Hi there, I am StrongMonkey:v3
```

To configure a webhook, run the following command

```bash
$ rio secret create -d accessToken=$(github_access_token) webhook
default/webhook

# Right now every commit and tag event will trigger a new revision
$ rio run -p 8080/http --build-secret webhook -n build-webhook https://github.com/StrongMonkey/demo.git
default/build-webhook
```

To view logs from your builds
```bash
$ rio builds
NAME                                                                     SERVICE                   REVISION                                   CREATED        SUCCEED   REASON
default/fervent-swartz6-ee709-786b366d5d44de6b547939f51d467437e45c5ee1   default/fervent-swartz6   786b366d5d44de6b547939f51d467437e45c5ee1   23 hours ago   True    

$ rio logs -f default/fervent-swartz6-ee709-786b366d5d44de6b547939f51d467437e45c5ee1
```

## Monitoring

By default, Rio will deploy [Grafana](https://grafana.com/) and [Kiali](https://www.kiali.io/) to give users the ability to watch all metrics of the service mesh.

```bash
# Monitoring services are deployed into rio-system namespace
$ rio --system ps
NAME                          ENDPOINT                                            REVISIONS   SCALE     WEIGHT
rio-system/autoscaler                                                             v0          1         100%
rio-system/build-controller                                                       v0          1         100%
rio-system/buildkit                                                               v0          1         100%
rio-system/cert-manager                                                           v0          1         100%
rio-system/grafana            https://grafana-rio-system.iazlia.on-rio.io:9443    v0          1         100%
rio-system/istio-citadel                                                          v0          1         100%
rio-system/istio-gateway                                                          v0          1         100%
rio-system/istio-pilot                                                            v0          1         100%
rio-system/istio-telemetry                                                        v0          1         100%
rio-system/kiali              https://kiali-rio-system.iazlia.on-rio.io:9443      v0          1         100%
rio-system/prometheus                                                             v0          1         100%
rio-system/registry           https://registry-rio-system.iazlia.on-rio.io:9443   v0          1         100%
rio-system/webhook            https://webhook-rio-system.iazlia.on-rio.io:9443    v0          1         100%
```

![Grafana](https://raw.githubusercontent.com/StrongMonkey/rio/refactor/grafana-example.png)

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
