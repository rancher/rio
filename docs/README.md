# Documentation

## Usage

- [Install Options](#install-options)
- [Running workloads](#running-workload)
- [Monitoring](#Monitoring)
- [AutoScaling based on QPS](#autoscaling)
- [Continuous Delivery](#continuous-deliverysource-code-to-deployment)
- [Using Riofile to build and develop](#using-riofile-to-build-and-develop-application)

### Install Options
Rio provides three install options for users. 
`ingress`: Rio will use existing ingress controller and ingress resource to expose gateway services. All the traffic will go through ingress then inside cluster. Starting v0.4.0 this is the default mode.
`svclb`: Rio will use service loadbalancer to expose gateway services. 
`hostport`: Rio will expose hostport for gateway services.

There are other install options:
`http-port`: Http port gateway service will listen. If install mode is ingress, it will 80.
`https-port`: Https port gateway service will listen. If install mode is ingress, it will 443.
`ip-address`: Manually specify worker IP addresses to generate DNS domain. By default Rio will detect based on install mode.
`service-cidr`: Manually specify service-cidr for service mesh to intercept traffic. By default Rio will try to detect.
`disable-features`: Specify feature to disable during install.
`httpproxy`: Specify HTTP_PROXY environment variable for control plane.
`lite`: install with lite mode.

### Running workload

- [Quick start](#quick-start)
- [Canary Deployment](#canary-deployment)
- [Automatic DNS and HTTPS](#automatic-dns-and-https)
- [Adding external services](#adding-external-services)
- [Adding Router](#adding-router)
- [Adding Public domain](#adding-public-domain)
- [Using Riofile](#using-riofile)


##### Quick start
To deploy workload to rio:
```bash
# ibuildthecloud/demo:v1 is a docker image that listens on 80 and print "hello world"
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

Rio provides a similar experience as Docker CLI when running a container. Run `rio run --help` to see more options.

##### Canary Deployment
Rio allows you to easily configure canary deployment by staging services and shifting traffic between revisions.

```bash
# Create a new service
$ rio run -p 80/http --name demo1 ibuildthecloud/demo:v1

# Stage a new version, updating just the docker image and assigning it to "v3" version. If you want to change options other than just image, run with --edit.
$ rio stage --image=ibuildthecloud/demo:v3 default/demo1:v3
$ rio stage --edit default/svc:v3

# Notice a new URL was created for your staged service. For each revision you will get a unique URL.
$ rio revision default/demo1
Name               IMAGE                    CREATED          SCALE     ENDPOINT                                         WEIGHT    DETAIL
default/demo1:v3   ibuildthecloud/demo:v3   19 seconds ago   1         https://demo1-v3-default.5yt5mw.on-rio.io:9443   0         
default/demo1:v0   ibuildthecloud/demo:v1   2 minutes ago    1         https://demo1-v0-default.5yt5mw.on-rio.io:9443   100   

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
default/demo1:v3   ibuildthecloud/demo:v3   About a minute ago   1         https://demo1-v3-default.5yt5mw.on-rio.io:9443   5         
default/demo1:v0   ibuildthecloud/demo:v1   3 minutes ago        1         https://demo1-v0-default.5yt5mw.on-rio.io:9443   95   

# Access the app. You should be able to see traffic routing to the new revision
$ curl https://demo1-default.5yt5mw.on-rio.io:9443
Hello World

$ curl https://demo1-default.5yt5mw.on-rio.io:9443
Hello World v3

# Wait for v3 to be 100% weight. Access the app, all traffic should be routed to new revision right now.
$ rio revision default/svc
Name               IMAGE                    CREATED         SCALE     ENDPOINT                                         WEIGHT    DETAIL
default/demo1:v3   ibuildthecloud/demo:v3   4 minutes ago   1         https://demo1-v3-default.5yt5mw.on-rio.io:9443   100       
default/demo1:v0   ibuildthecloud/demo:v1   6 minutes ago   1         https://demo1-v0-default.5yt5mw.on-rio.io:9443   0         

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
$ rio domain register www.myproductionsite.com default/route1
default/foo-bar

# Use your own certs by providing a secret that contain tls cert and key instead of provisioning by letsencrypts. The secret has to be created first in system namespace.
$ rio domain register --secret $name www.myproductionsite.com default/route1

# Access your domain 
```

Note: By default Rio will automatically configure Letsencrypt HTTP-01 challenge to provision certs for your publicdomain. This needs you to install rio on standard ports.
If you are install rio with svclb or hostport mode, try `rio install --httpport 80 --httpsport 443`.

##### Using Riofile

###### Riofile example

Rio allows you to define a file called `Riofile`. `Riofile` allows you define rio services and configmap is a friendly way with `docker-compose` syntax.
For example, to define a nginx application with conf

```yaml
configs:
  conf:
    index.html: |-
      <!DOCTYPE html>
      <html>
      <body>
      
      <h1>Hello World</h1>
      
      </body>
      </html>
services:
  nginx:
    image: nginx
    ports:
    - 80/http
    configs:
    - conf/index.html:/usr/share/nginx/html/index.html
```

```yaml
services:
  demo:
    build:
      repo: https://github.com/rancher/rio-demo
      branch: master
    ports:
    - 80/http
    scale: 1-20
```

Once you have defined `Riofile`, simply run `rio up`. Any change you made for `Riofile`, re-run `rio up` to pick the change.

More complicated examples are available at [here](../stacks).

###### Riofile reference
`arg`: Arguments to the entrypoint

`command`: Entrypoint array

`scale`: Scale of the service. Can be specifed as `min-max`(1-10) to enable autoscaling.

`ports`: Container ports. Format: `$(servicePort:)containerPort/protocol`. 

`build`: Build arguments.  

```yaml
build:
    buildArgs: # build arguments
    - foo=bar
    dockerFile: Dockerfile # the name of Dockerfile to look for
    dockerFilePath: ./ # the path of Dockerfile to look for
    buildContext: ./  # build context
    noCache: true # build without cache
    push: true
    buildImageName: foo/bar # specify custom image name
    pushRegistry: docker.io # specify push registry
```

`configs`: Specify configmap to mount. Format: `$name/$key:/path/to/file`.

`secrets`: Specify secret to mount. Format: `$name/$key:/path/to/file`.

`pullPolicy`: Specify image pull policy. Options: `always/never/ifNotProsent`.

`disableServiceMesh`: Disable service mesh sidecar.

`global_permissions`: Specify the global permission of workload

Example: 
```yaml
global_permissions:
- 'create,get,list certmanager.k8s.io/*'
```
this will give workload abilities to **create, get, list** **all** resources in api group **certmanager.k8s.io**.

If you want to hook up with an existing role:

```yaml
global_permissions:
- 'role=cluster-admin'
```

`permisions`: Specify current namespace permission of workload

Example: 
```yaml
permissions:
- 'create,get,list certmanager.k8s.io/*'
```

this will give workload abilities to **create, get, list** **all** resources in api group **certmanager.k8s.io** in **current** namespace.

`labels`: Specify labels

`annotations`: Specify annotations

`containers`: Specify multiple containers.

`env`: Specify environment variables. You can use the following syntax.

```yaml
"self/name":           "metadata.name",
"self/namespace":      "metadata.namespace",
"self/labels":         "metadata.labels",
"self/annotations":    "metadata.annotations",
"self/node":           "spec.nodeName",
"self/serviceAccount": "spec.serviceAccountName",
"self/hostIp":         "status.hostIP",
"self/nodeIp":         "status.hostIP",
"self/ip":             "status.podIP",
```
For example, to set an environment name to its own name
```yaml
env:
- POD_NAME=$(self/name)
```

###### Watching Riofile
You can setup github repository to watch Riofile changes and re-apply Riofile changes. Here is the example:
```bash
$ rio up https://github.com/username/repo
```
If you want to setup webhook to watch, go to [here](#setup-github-webhook-experimental)


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

#### Setup credential for private repository and github webhook
1. Set up git basic auth.(Currently ssh key is not supported and will be added soon). Here is an exmaple of adding a github repo.
```bash
$ rio secret add --git-basic-auth
Select namespace[default]: $(put the same namespace with your workload)
git url: https://github.com/username
username: $username
password: $password
```
2. Run your workload and point it to your private git repo. It will automatically use the secret you just configured.

#### Setup Github webhook (experimental)
By default, rio will automatically pull git repo and check if repo code has changed. You can also configure a webhook to automatically push any events to Rio to trigger the build.

1. Set up Github webhook token.
```bash
$ rio secret add --github-webhook
Select namespace[default]: $(put the same namespace with your workload)
accessToken: $(github_accesstoken)
```
$(github_accesstoken) has to be able create webhook in your github repo.
2. Create workload and point to your repo.
3. Go to your Github repo, it should have webhook configured to point to one of our webhook service.

#### Set Custom build arguments and docker registry
You can also push to your own registry for images that rio has built.
1. Setup docker registry auth. Here is an example of how to setup docker registry.
```bash
$ rio secret add --docker
Select namespace[default]: $(put the same namespace with your workload)
Registry url[]: https://index.docker.io/v1/
username[]: $(your_docker_hub_username)
password[]: $(password)
```
2. Create your workload. Set the correct push registry.
```bash
$ rio run --build-registry docker.io --build-image-name $(username)/yourimagename $(repo)
```
`docker.io/$(username)/yourimagename` will be pushed into dockerhub registry. 

#### Enable Pull request (experimental)
Rio also allows you to configure pull request builds. This needs you to configure github webhook token correctly.
1. Set up github webhook token in the previous session
2. Run workload with pull-request enabled.
```bash
$ rio run --build-enable-pr $(repo)
```

After this, if there is any pull request, Rio will create a deployment based on this pull request, and you will get a unique link
to see the change this pull request introduced in the actual deployment.

#### View build logs
To view logs from your builds
```bash
$ rio builds
NAME                                                                     SERVICE                   REVISION                                   CREATED        SUCCEED   REASON
default/fervent-swartz6-ee709-786b366d5d44de6b547939f51d467437e45c5ee1   default/fervent-swartz6   786b366d5d44de6b547939f51d467437e45c5ee1   23 hours ago   True    

$ rio logs -f default/fervent-swartz6-ee709-786b366d5d44de6b547939f51d467437e45c5ee1

# restart any builds that failed
$ rio build restart default/fervent-swartz6-ee709-786b366d5d44de6b547939f51d467437e45c5ee1
```

### Using Riofile to build and develop application

Rio allows developer to build and develop applications from local source code. Rio will by default use buildkit to build application.

Requirements:
1. Local repo must have `Dockerfile` and `Riofile`.
2. Developer have rio installed in a **single-node k3s** cluster. We will support minikube later, but as today buildkit is not supported in minikube.(https://github.com/kubernetes/minikube/issues/4143 )

Use cases:
1. `git clone https://github.com/StrongMonkey/riofile-demo.git`
2. `cd riofile-demo`
3. `rio up`. It will build the project and bring up services.
4. `rio ps`. 
5. `vim main.go && change "Hi there, I am demoing Riofile" to "Hi there, I am demoing something"`
6. Re-run `rio up`. It will rebuild. After it is done, revisit service endpoint to see if content is changed.

If you want more complicated build arguments, rio supports the following format
```yaml
services:
  demo:
   ports:
   - 8080/http
   build:
    buildArgs:
    - foo=bar
    dockerFile: Dockerfile
    dockerFilePath: ./
    buildContext: ./
    noCache: true
    push: true
    buildImageName: docker.io/foo/bar
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

## FAQ

* How can I upgrade rio?
```
Upgrading rio just needs the latest release of rio binary. Re-run `rio install` with your install options.
```

* How can I swap out letsencrypt certificate with my own certs?
```
Create a TLS secret in `rio-system` namespace that contains your tls cert and key. Edit cluster domain by running `k edit clusterdomain cluster-domain -n rio-system`.
Change spec.secretRef.name to the name of your TLS secret.
```

* How can I use my own DNS domain?
```
Disable rdns and letsencrypt features by running `rio install --disable-features rdns,letsencrypt`. Edit cluster domain by running `k edit clusterdomain cluster-domain -n rio-system`.
Change status.domain to your own wildcard doamin. You are responsible to manage your dns record to gateway IP or worker nodes.
```

