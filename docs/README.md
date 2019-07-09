# Documentation

## Usage

- [Install Options](#install-options)
- [Running workloads](#running-workload)
- [Monitoring](#Monitoring)
- [AutoScaling based on QPS](#autoscaling)
- [Continuous Delivery](#continuous-deliverysource-code-to-deployment)

### Install Options
Rio by default expose 9443 and 9080 for https and http traffic. To change port, `rio run --http-port ${http_port} --https-port ${https_port}`.
More advanced options are also available by running `rio install --help`. To change install options, re-run `rio install ${args}`. 

### Running workload

- [Example](#Example)
- [Canary Deployment](#canary-deployment)
- [Automatic DNS and HTTPS](#automatic-dns-and-https)
- [Adding external services](#adding-external-services)
- [Adding Router](#adding-router)
- [Adding Public domain](#adding-public-domain)

##### Example
To run a workload, simple type `rio run -p ${port} $image`. For example

```bash
$ rio run -p 80/http --name svc ibuildthecloud/demo:v1
default/svc:v0

# See the endpoint of your workload
$ rio ps
Name          CREATED          ENDPOINT                                    REVISIONS   SCALE     WEIGHT    DETAIL
default/svc   53 seconds ago   https://svc-default.5yt5mw.on-rio.io:9443   v0          1         100%      

### Access your workload
$ curl https://svc-default.5yt5mw.on-rio.io:9443
Hello World
```

##### Canary Deployment
Rio allows you to easily configure canary deployment by staging services and shifting traffic between revisions.

```bash
# Create a new service
$ rio run -p 80/http --name demo1 --scale=3 ibuildthecloud/demo:v1
default/svc:v0

# Stage a new version, updating just the docker image and assigning it to "v3" version.
$ rio stage --image=ibuildthecloud/demo:v3 default/demo1:v3
default/svc:v3

# Or change the spec of the new service by adding --edit
$ rio stage --edit default/svc:v3

# Notice a new URL was created for your staged service
$ rio revision default/demo1
Name               IMAGE                    CREATED          SCALE     ENDPOINT                                         WEIGHT    DETAIL
default/demo1:v3   ibuildthecloud/demo:v3   19 seconds ago   3         https://demo1-v3-default.5yt5mw.on-rio.io:9443   0         
default/demo1:v0   ibuildthecloud/demo:v1   2 minutes ago    3         https://demo1-v0-default.5yt5mw.on-rio.io:9443   100   

# Access the current revision
$ curl -s https://demo1-v0-default.5yt5mw.on-rio.io:9443
Hello World

# Access the staged service under the new URL
$ curl -s https://demo1-v3-default.5yt5mw.on-rio.io:9443
Hello World v3

# Promote v3 service. The traffic will be shifted to v3 gradually. By default we apply a 5% shift every 5 seconds, but it can be configured
# using the flags `--rollout-increment` and `--rollout-interval`. To turn off rollout(the traffic percentage will be changed to
# the desired value immediately), run `--no-rollout`.
$ rio promote default/demo1:v3

Name               IMAGE                    CREATED              SCALE     ENDPOINT                                         WEIGHT    DETAIL
default/demo1:v3   ibuildthecloud/demo:v3   About a minute ago   3         https://demo1-v3-default.5yt5mw.on-rio.io:9443   5         
default/demo1:v0   ibuildthecloud/demo:v1   3 minutes ago        3         https://demo1-v0-default.5yt5mw.on-rio.io:9443   95   

# Access the app. You should be able to see traffic routing to the new revision
$ curl https://demo1-default.5yt5mw.on-rio.io:9443
Hello World

$ curl https://demo1-default.5yt5mw.on-rio.io:9443
Hello World v3

# Wait for v3 to be 100% weight. Access the app, all traffic should be routed to new revision right now.
$ rio revision default/svc
Name               IMAGE                    CREATED         SCALE     ENDPOINT                                         WEIGHT    DETAIL
default/demo1:v3   ibuildthecloud/demo:v3   4 minutes ago   3         https://demo1-v3-default.5yt5mw.on-rio.io:9443   100       
default/demo1:v0   ibuildthecloud/demo:v1   6 minutes ago   3         https://demo1-v0-default.5yt5mw.on-rio.io:9443   0         

$ curl https://demo1-default.5yt5mw.on-rio.io:9443
Hello World v3

# Manually adjusting weight between revisions
$ rio weight default/demo1:v0=5% default/demo1:v3=95%

$ rio ps
Name            CREATED             ENDPOINT                                      REVISIONS   SCALE     WEIGHT    DETAIL
default/demo1   7 minutes ago       https://demo1-default.5yt5mw.on-rio.io:9443   v0,v3       3,3       5%,95%    
```

##### Automatic DNS and HTTPS
By default Rio will create a DNS record pointing to your cluster. Rio also uses Let's Encrypt to create
a certificate for the cluster domain so that all services support HTTPS by default.
For example, when you deploy your workload, you can access your workload in HTTPS. The domain always follows the format
of ${app}-${namespace}.\${cluster-domain}. You can see your cluster domain by running `rio info`.

##### Adding external services
ExternalService is a service(databases, legacy apps) that is outside of your cluster, and can be added into service discovery.
It can be IPs, FQDN or service in another namespace. Once added, external service can be discovered by short name within the same namespace.

```bash
$ rio external create ${namespace/name} mydb.com

$ rio external create ${namespace/name} 8.8.8.8

$ rio external create ${namespace/name} ${another_svc/another_namespace}

```

##### Adding Router
Router is a set of L7 load-balancing rules that can route between your services. It can add Header-based, path-based routing, cookies
and other rules. For example, to add path-based routing,

```bash
$ rio run -p 80/http --name svc ibuildthecloud/demo:v1
$ rio run -p 80/http --name svc3 ibuildthecloud/demo:v3

# Create a route to point to svc:v0 and svc:v3
$ rio route add route1/to-svc-v0 to default/svc
$ rio route add route1/to-svc-v3 to default/svc3

# Access the route
$ rio route
Name             URL                                                      OPTS      ACTION    TARGET
default/route1   https://route1-default.5yt5mw.on-rio.io:9443/to-svc-v0             to        svc:v0,port=80
default/route1   https://route1-default.5yt5mw.on-rio.io:9443/to-svc-v3             to        svc:v3,port=80

$ curl -s https://route1-default.iazlia.on-rio.io:9443/to-svc-v0
Hello World

$ curl -s https://route1-default.iazlia.on-rio.io:9443/to-svc-v3
Hello World v3
```

##### Adding Public domain
Rio allows you to add a vanity domain to your workloads. For example, to add a domain `www.myproductionsite.com` to your workload,
run
```bash
# Create a domain that points to route1. You have to setup a cname record from your domain to cluster domain.
# For example, foo.bar -> CNAME -> iazlia.on-rio.io
$ rio domain add foo.bar default/route1
default/foo-bar

# Use your own certs by providing a secret that contain tls cert and key instead of provisioning by letsencrypts. The secret has to be created first in system namespace.
$ rio domain add --secret $name foo.bar default/route1
```

Note: By default Rio will automatically configure Letsencrypt HTTP-01 challenge to provision certs for your publicdomain. This needs you to install rio on standard ports.
Try `rio install --httpport 80 --httpsport 443`.

### Monitoring
By default, Rio will deploy [Grafana](https://grafana.com/) and [Kiali](https://www.kiali.io/) to give users the ability to watch all metrics of the service mesh.
You can find endpoints of both services by running `rio -s ps`.

```bash
Name                          CREATED       ENDPOINT                                           REVISIONS   SCALE     WEIGHT    DETAIL
rio-system/controller         7 hours ago                                                      v0          1         100%      
rio-system/activator          7 hours ago                                                      v0          1         100%      
rio-system/kiali              9 hours ago   https://kiali-rio-system.5yt5mw.on-rio.io:9443     v0          1         100%      
rio-system/cert-manager       9 hours ago                                                      v0          1         100%      
rio-system/istio-pilot        9 hours ago                                                      v0          1         100%      
rio-system/istio-gateway      9 hours ago                                                      v0          1         100%      
rio-system/istio-citadel      9 hours ago                                                      v0          1         100%      
rio-system/istio-telemetry    9 hours ago                                                      v0          1         100%      
rio-system/grafana            9 hours ago   https://grafana-rio-system.5yt5mw.on-rio.io:9443   v0          1         100%      
rio-system/registry           9 hours ago                                                      v0          1         100%      
rio-system/webhook            9 hours ago   https://webhook-rio-system.5yt5mw.on-rio.io:9443   v0          1         100%      
rio-system/autoscaler         9 hours ago                                                      v0          1         100%      
rio-system/build-controller   9 hours ago                                                      v0          1         100%      
rio-system/prometheus         9 hours ago                                                      v0          1         100%  
```

### Autoscaling
By default each workload is enabled with autoscaling(min scale 1, max scale 10), which means the workload can be scaled from 1 instance to 10 instances
depending on how much traffic it receives. To change the scale range, run `rio run --scale=$min-$max ${args}`. To disable autoscaling,
 run `rio run --scale=${num} ${args}`
 
```bash
# Run a workload, set the minimal and maximum scale
$ rio run -p 8080/http --name autoscale --scale=1-20 strongmonkey1992/autoscale:v0
default/autoscale:v0

# Put some load to the workload. We use [hey](https://github.com/rakyll/hey) to create traffic
$ hey -z 600s -c 60 http://autoscale-v0-default.5yt5mw.on-rio.io:9080

# Note that the service has been scaled to 6 instances
$ rio revision default/autoscale
Name                   IMAGE                           CREATED          SCALE     ENDPOINT                                             WEIGHT    DETAIL
default/autoscale:v0   strongmonkey1992/autoscale:v0   49 seconds ago   1         https://autoscale-v0-default.5yt5mw.on-rio.io:9443   100       

# Run a workload that can be scaled to zero
$ rio run -p 8080/http --name autoscale-zero --scale=0-20 strongmonkey1992/autoscale:v0
default/autoscale-zero:v0

# Wait a couple of minutes for the workload to scale to zero
$ rio revision default/autoscale-zero
Name                        IMAGE                           CREATED         SCALE     ENDPOINT                                                  WEIGHT    DETAIL
default/autoscale-zero:v0   strongmonkey1992/autoscale:v0   9 seconds ago   1         https://autoscale-zero-v0-default.5yt5mw.on-rio.io:9443   100       

# Access the workload. Once there is an active request, the workload will be re-scaled to active.
$ rio ps
Name                     CREATED          ENDPOINT                                               REVISIONS   SCALE     WEIGHT    DETAIL
default/autoscale-zero   13 minutes ago   https://autoscale-zero-default.5yt5mw.on-rio.io:9443   v0          0/1       100%     

$ curl -s https://autoscale-zero-v0-default.5yt5mw.on-rio.io:9443
Hi there, I am StrongMonkey:v13

# Verify that the workload has been re-scaled to 1
$ rio revision default/autoscale-zero
Name                        IMAGE                           CREATED         SCALE     ENDPOINT                                                  WEIGHT    DETAIL
default/autoscale-zero:v0   strongmonkey1992/autoscale:v0   9 seconds ago   1         https://autoscale-zero-v0-default.5yt5mw.on-rio.io:9443   100       
```

### Continuous Delivery(Source code to Deployment)
Rio supports configuration of a Git-based source code repository to deploy the actual workload. It can be as easy
as giving Rio a valid Git repository repo URL.

```bash
# Run a workload from a git repo. We assume the repo has a Dockerfile at root directory to build the image
$ rio run -n build https://github.com/StrongMonkey/demo.git
default/build:v0

# Waiting for the image to be built. Note that the image column is empty. Once the image is ready service will be active
$ rio revision
Name               IMAGE     CREATED         SCALE     ENDPOINT                                         WEIGHT    DETAIL
default/build:v0             6 seconds ago   0/1       https://build-v0-default.5yt5mw.on-rio.io:9443   100     

# The image is ready. Note that we deploy from the default docker registry into the cluster.
# The image name has the format of ${registry-domain}/${namespace}/${name}:${commit}
$ rio revision
Name               IMAGE                                                    CREATED              SCALE     ENDPOINT                                         WEIGHT    DETAIL
default/build:v0   default-build:ff564b7058e15c3e6813f06feb965af7787f0b28   About a minute ago   1         https://build-v0-default.5yt5mw.on-rio.io:9443   100  


# Show the endpoint of your workload
$ rio ps
Name            CREATED              ENDPOINT                                      REVISIONS   SCALE     WEIGHT    DETAIL
default/build   About a minute ago   https://build-default.5yt5mw.on-rio.io:9443   v0          1         100%      

# Access the endpoint
$ curl -s https://build-default.5yt5mw.on-rio.io:9443
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

To configure webhook secret, run `rio secret add --github`. Once webhook secret is configured, a webhook will be created to create new revision based on push and tag events.

To configure private git repo for basic auth, run `rio secret add --git`. After this, you can add private git repo to Rio.

To configure private docker registry, run `rio secret add --docker`. This configures docker push secret so you can push to a custom registry instead of built-in local registry.

To configure custom dockerFile path and buildContext, run `rio run --docker-file ${path} --build-context ${path}`

To configure custom registry and image name, run `rio run --build-registry ${registry} --build-image-name ${image_name}`

To enable build for pull-request, run `rio run --build-enable-pr ${args}`. 

To view logs from your builds
```bash
$ rio builds
NAME                                                                     SERVICE                   REVISION                                   CREATED        SUCCEED   REASON
default/fervent-swartz6-ee709-786b366d5d44de6b547939f51d467437e45c5ee1   default/fervent-swartz6   786b366d5d44de6b547939f51d467437e45c5ee1   23 hours ago   True    

$ rio logs -f default/fervent-swartz6-ee709-786b366d5d44de6b547939f51d467437e45c5ee1
```

## Concept

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

### Service Mesh

Rio has a built in service mesh, powered by Istio and Envoy. The service mesh provides all of the core communication
abilities for services to talk to each other, inbound traffic and outbound traffic. All traffic can be encrypted,
validated, and routed dynamically according to the configuration. Rio specifically does not require the user to
understand much about the underlying service mesh.

