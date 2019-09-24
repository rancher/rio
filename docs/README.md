# Documentation

## Table of Contents

- [Install Options](#install-options)
- [Concepts](#concepts)
- [Running workloads](#running-workload)
  - [Canary Deployment](#canary-deployment)
  - [Automatic DNS and HTTPS](#automatic-dns-and-https)
  - [Adding external services](#adding-external-services)
  - [Adding Router](#adding-router)
  - [Adding Public domain](#adding-public-domain)
  - [Using Riofile](#using-riofile)
- [Monitoring](#Monitoring)
- [AutoScaling based on QPS](#autoscaling)
- [Continuous Delivery](#continuous-deliverysource-code-to-deployment)
  - [Example](#example)
  - [Setting Private repository](#setup-credential-for-private-repository)
  - [Setting github webhook](#setup-github-webhook-experimental)
  - [Setting private registry](#set-custom-build-arguments-and-docker-registry)
  - [Setting Pull Request Feature](#enable-pull-request-experimental)
  - [View Build logs](#view-build-logs)
- [Local Developer Setup](#local-developer-setup)
- [FAQ](#faq)

## Install Options
Rio provides a number of options when installed using `rio install`.

* `mode`: How Rio exposes the service mesh gateway. All HTTP requests to Rio services go through the gateway. There are three options:

| Mode | Description |
|------|-------------|
| `ingress` | Rio will use existing ingress controller and ingress resource to expose gateway services. All the traffic will go through ingress. Starting v0.4.0 this is the default mode. |
| `svclb` | Rio will use service loadbalancer to expose gateway services. | 
| `hostport` | Rio will expose hostport for gateway services. |

* `http-port`: HTTP port gateway service will listen. You can only set the HTTP port if install mode is svclb or hostport. Default HTTP port for svclb and hostport mode is 9080. If install mode is ingress, HTTP port is determined by ingress and cannot be changed by Rio installer. Ingress controllers typically expose HTTP port 80, although some ingress controllers allow you to specify custom HTTP ports.
* `https-port`: HTTPS port gateway service will listen. You can only set the HTTPS port if install mode is svclb or hostport. Default HTTPS port for svclb and hostport mode is 9443. If install mode is ingress, HTTPS port is determined by ingress and cannot be changed by Rio installer. Ingress controllers typically expose HTTPS port 443, although some ingress controllers allow you to specify custom HTTPS ports.
* `ip-address`: Rio generates DNS domains that map to IP address of the gateway. Rio will attempt to detect gateway IP addresses automatically, and you can override by manually specify comma-separated IP addresses.
* `service-cidr`: Rio will attempt to detect service CIDR for service mesh intercept traffic. You can override by manually specify a service CIDR.
* `disable-features`: Specify feature to disable during install. Here are the available feature list.

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

* `httpproxy`: Specify HTTP_PROXY environment variable for Rio control plane. This is useful when Rio is installed behind an HTTP firewall.
* `lite`: Disable all monitoring features including prometheus, mixer, grafana and kiali.

## Concepts

Rio introduces a small number of new concepts: Service, Apps, Router, External Service, and Domain. In addition, it reuses two existing Kubernetes resources: ConfigMaps and Secrets.

The power of Rio resides in its ability to utilize the power of Kubernetes, Istio service mesh, Knative, and Tekton CI/CD through a simple set of concepts.

### Service

Service is the core concept in Rio. Services are a scalable set of identical containers.
When you run containers in Rio you create a Service. `rio run` and `rio create` will
create a service. You can scale that service with `rio scale`. Services are assigned a DNS name so that it can be discovered and accessed from other services.

### Apps

An App contains multiple services, and each service can have multiple revisions. Each service in Rio is assigned an app and a version. Services that have the same app but different versions are referred to as revisions.
An application named `foo` will be given a DNS name like `foo.clusterdomain.on-rio.io` and each version is assigned it's own DNS name. If the app was
`foo` and the version is `v2` the assigned DNS name for that revision would be similar to `foo-v2.clusterdomain.on-rio.io`. `rio ps` and `rio revision` will
list the assigned DNS names.

### Router

Router is a resource that manages load balancing and traffic routing rules. Routing rules can route based
on hostname, path, HTTP headers, protocol, and source.

### External Service

External Service provides a way to register external IPs or hostnames in the service mesh so they can be accessed by Rio services.

### Public Domain

Public Domain can be configured to assign a service or router a vanity domain like www.myproductionsite.com.

### Configs

ConfigMaps are a standard Kubernetes resource and can be referenced by Rio services. It is a piece of configuration which can be mounted into pods so that configuration data can be separated from image artifacts.

### Secrets

Secrets are a standard Kubernetes resource and can be referenced by rio services. It contains sensitive data which can be mounted into pods. 

## Running workload

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

### Canary Deployment
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

### Automatic DNS and HTTPS
By default Rio will create a DNS record pointing to your cluster's gateway. Rio also uses Let's Encrypt to create
a certificate for the cluster domain so that all services support HTTPS by default.
For example, when you deploy your workload, you can access your workload in HTTPS. The domain always follows the format
of ${app}-${namespace}.\${cluster-domain}. You can see your cluster domain by running `rio info`.

Some DNS servers provide protection against DNS rebinding attacks, and may disallow endpoint name resolution for `on-rio.io`. If that happens you need to configure whitelisting for `on-rio.io` on your DNS server.

### Adding external services
ExternalService is a service like databases and legacy apps outside of your cluster. 
ExternalService can be IP addresses, FQDN or a Rio service in another namespace. Once added, external service can be discovered by short name within the same namespace.

```bash
$ rio external create ${namespace/name} mydb.com

$ rio external create ${namespace/name} 8.8.8.8

$ rio external create ${namespace/name} ${another_svc/another_namespace}

```

### Adding Router
Router is a set of L7 load-balancing rules that can route between your services. It can add Header-based, path-based routing, cookies
and other rules.

Create router in a different namespace:
```bash
$ rio route add $name.$namespace to $target_namespace/target_service
```

Insert a router rule
```bash
$ rio route insert $name.$namespace to $target_namespace/target_service
```

Create a route based path match
```bash
$ rio route add $name.$namespace/path to $target_namespace/target_service
```

Create a route to a different port:
```bash
$ rio route add $name.$namespace to $target_namespace/target_service,port=8080
```

Create router based on header (supports exact match: `foo`, prefix match: `foo*`, and regular expression match: `regexp(foo.*)`)
```bash
$ rio route add --header USER=$format $name.$namespace to $target_namespace/target_service
```

Create router based on cookies (supports exact match: `foo`, prefix match: `foo*`, and regular expression match: `regexp(foo.*)`)
```bash
$ rio route add --cookie USER=$format $name.$namespace to $target_namespace/target_service
```

Create route based on HTTP method (supports exact match: `foo`, prefix match: `foo*`, and regular expression match: `regexp(foo.*)`)
```bash
$ rio route add --method GET $name.$namespace to $target_namespace/target_service
```

Add, set or remove headers:
```bash
$ rio route add --add-header FOO=BAR $name.$namespace to $target_namespace/target_service
$ rio route add --set-header FOO=BAR $name.$namespace to $target_namespace/target_service
$ rio route add --remove-header FOO=BAR $name.$namespace to $target_namespace/target_service
```

Mirror traffic:
```bash
$ rio route add $name.$namespace mirror $target_namespace/target_service
```

Rewrite host header and path
```bash
$ rio route add $name.$namespace rewrite $target_namespace/target_service
```

Redirect to another service
```bash
$ rio route add $name.$namespace redirect $target_namespace/target_service/path
```

Add timeout
```bash
$ rio route add --timeout $name.$namespace to $target_namespace/target_service
```

Add fault injection
```bash
$ rio route add --fault-httpcode 502 --fault-delay 1s --fault-percentage 80 $name.$namespace to $target_namespace/target_service
```

Add retry logic
```bash
$ rio route add --retry-attempts 5 --retry-timeout 1s $name.$namespace to $target_namespace/target_service
```

Create router to different revision and different weight
```bash
$ rio route add $name.$namespace to $service:v0,weight=50 $service:v1,weight=50
```

### Adding Public domain
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

Note: By default Rio will automatically configure Letsencrypt HTTP-01 challenge to provision certs for your public domain. This requires you to install rio on standard ports.
If you are install rio with svclb or hostport mode, try `rio install --http-port 80 --https-port 443`.

### Using Riofile

Rio works with standard Kubernetes YAML files. Rio additionally supports a more user-friendly `docker-compose`-style config file called `Riofile`. `Riofile` allows you define rio services, apps, routes, external services, configmap, and secrets.

For example, this is an example of an nginx application:

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

Once you have defined `Riofile`, simply run `rio up`. Any change you made for `Riofile`, re-run `rio up` to pick the change.

#### Riofile reference
```yaml
# Configmap
configs:          
  config-foo:     # specify name in the section 
    key1: |-      # specify key and data in the section 
      {{ config1 }}
    key2: |-
      {{ config2 }}
      
# Service
services:
  service-foo:
    disableServiceMesh: true # Disable service mesh side injection for service
    
    # Scale setting
    scale: 2 # Specify scale of the service. If you pass range `1-10`, it will enable autoscaling which can be scale from 1 to 10. Default to 1 if omitted
    updateBatchSize: 1 # Specify the update batch size. If not set, defaults to 1.
    
    # Revision setting
    app: my-app # Specify app name. Defaults to service name. This is used to aggregate services that belongs to the same app.
    version: v0 # Specify revision name. Defaults to v0.
    weight: 80 # Weight assigned to this revision. Value: 0-100. Defaults to 100.
    
    # Autoscaling setting. Only required if autoscaling is set
    concurrency: 10 # specify concurrent request each pod can handle(soft limit, used to scale service)
    
    # Traffic rollout config. Optional
    rollout: true # whether rollout traffic gradually
    rolloutIncrement: 5 # traffic percentage increment(%) for each interval. Will not work if rollout is false. Required if rollout is true
    rolloutInterval: 2 # traffic increment interval(seconds). Will not work if rollout is false. Required if rollout is true 
    
    # Permission for service
    # 
    #   global_permissions:
    #   - 'create,get,list certmanager.k8s.io/*'
    #  
    #   this will give workload abilities to **create, get, list** **all** resources in api group **certmanager.k8s.io**.
    #
    #   If you want to hook up with an existing role:
    #
    #   
    #   global_permissions:
    #   - 'role=cluster-admin'
    #   
    #
    #   - `permisions`: Specify current namespace permission of workload
    #
    #   Example: 
    #   
    #   permissions:
    #   - 'create,get,list certmanager.k8s.io/*'
    #  
    #
    #   This will give workload abilities to **create, get, list** **all** resources in api group **certmanager.k8s.io** in **current** namespace. 
    #   
    #   Example: 
    #   
    #   permissions:
    #   - 'create,get,list /node/proxy'
    #   
    #    This will give subresource for node/proxy
    
    # Optional, will created and mount serviceAccountToken into pods with corresponding permissions 
    global_permissions:
    - 'create,get,list certmanager.k8s.io/*'
    permissions:
    - 'create,get,list certmanager.k8s.io/*'
    
    # Container configuration
    image: # Container image. Required if not setting build
    imagePullPolicy: # Image pull policy. Options: (always/never/ifNotProsent), defaults to ifNotProsent. 
    build: # Setting build parameters. Set if you want to build image for source
      repo: https://github.com/rancher/rio # Git repository to build. Required
      branch: master # Git repository branch. Required
      revision: v0.1.0 # Revision digest to build. If set, image will be built based on this revision. Otherwise it will take head revision in repo. Also if revision is not set, it will be served as the base revision to watch any change in repo and create new revision based changes from repo.
      buildArgs: # Build arguments to pass to buildkit https://docs.docker.com/engine/reference/builder/#understand-how-arg-and-from-interact. Optional
      - foo=bar
      dockerFile: Dockerfile # The name of Dockerfile to look for. Defaults to Dockerfile
      dockerFilePath: ./ # The path of Dockerfile to look for. Defaults to ./
      buildContext: ./  # Docker build context. Defaults to ./
      noCache: true # Build without cache. Defaults to false.
      buildImageName: myname/image:tag # Specify custom image name(excluding registry name). Default name: $namespace/name:$revision_digest
      pushRegistry: docker.io # Specify push registry. Example: docker.io, gcr.io. Defaults to localhost registry.
      stageOnly: true # If set, newly created revision will get any traffic. Defaults to false.
      githubSecretName: secretGithub # specify github webhook secretName to setup github webhook. Defaults to global secret that is configured by rio cli. [link](#setup-github-webhook-experimental)
      gitSecretName: secretGit # Specify git secret name for private git repository. Defaults to global secret that is configured by rio cli. [link](#setup-credential-for-private-repository)
      pushRegistrySecretName: secretDocker # Specify secret name for pushing to docker registry. [link](#set-custom-build-arguments-and-docker-registry)
      enablePr: true # Enable pull request feature. Defaults to false
    command: # Container entrypoint, not executed within a shell. The docker image's ENTRYPOINT is used if this is not provided.
    - echo
    args: # Arguments to the entrypoint. The docker image's CMD is used if this is not provided.
    - "hello world"
    workingDir: /home # Container working directory
    ports: # Container ports, format: `$(servicePort:)containerPort/protocol`. Required if user wants to expose service through gateway
    - 8080:80/http,web # Service port 8080 will be mapped to container port 80 with protocol http, named `web`
    - 8080/http,admin,internal=true # Service port 8080 will be mapped to container port 8080 with protocol http, named `admin`, internal port(will not be exposed through gateway) 
    env: # Specify environment variable
    - POD_NAME=$(self/name) # Mapped to "metadata.name" 
    # 
    # "self/name":           "metadata.name",
    # "self/namespace":      "metadata.namespace",
    # "self/labels":         "metadata.labels",
    # "self/annotations":    "metadata.annotations",
    # "self/node":           "spec.nodeName",
    # "self/serviceAccount": "spec.serviceAccountName",
    # "self/hostIp":         "status.hostIP",
    # "self/nodeIp":         "status.hostIP",
    # "self/ip":             "status.podIP",
    # 
    cpus: 100m # Cpu request, format 0.5 or 500m. 500m = 0.5 core. If not set, cpu request will not be set. https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
    memory: 100 mi # Memory request. 100mi, available options. If not set, memory request will not be set. https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
    secrets: # Specify secret to mount. Format: `$name/$key:/path/to/file`. Secret has to be pre-created in the same namespace
    - foo/bar:/my/password
    configs: # Specify configmap to mount. Format: `$name/$key:/path/to/file`. 
    - foo/bar:/my/config
    livenessProbe: # LivenessProbe setting. https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/
      httpGet:
        path: /ping
        port: 9997
      initialDelaySeconds: 10
    readinessProbe: # ReadinessProbe https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/
      failureThreshold: 7
      httpGet:
        path: /ready
        port: 9997
    stdin: true # Whether this container should allocate a buffer for stdin in the container runtime
    stdinOnce: true # Whether the container runtime should close the stdin channel after it has been opened by a single attach. When stdin is true the stdin stream will remain open across multiple attach sessions.
    tty: true # Whether this container should allocate a TTY for itself
    user: 1000 # The UID to run the entrypoint of the container process.
    group: 1000 # The GID to run the entrypoint of the container process
    readOnly: true # Whether this container has a read-only root filesystem
    
    nodeAffinity: # Describes node affinity scheduling rules for the pod.
    podAffinity:  # Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)).
    podAntiAffinity: # Describes pod anti-affinity scheduling rules (e.g. avoid putting this pod in the same node, zone, etc. as some other pod(s)).
    
    addHost: # Hostname alias
    net: host # Host networking
    imagePullSecrets: # Image pull secret https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
    - secret1
    - secret2 
    
    containers: # Specify sidecars
    - init: true # Init container
      image: ubuntu
      args:
      - "echo"
      - "hello world"
      # Other options are available in container section above
 
# Router
routers:
  foo:
    routes:
    - matches: # Match rules, the first rule matching an incoming request is used
      - path: # Match path, can specify regxp, prefix or exact match 
          regxp: /bar.*
          # prefix: /bar
          # exact: /bar
        scheme:
          regxp: /bar.*
          # prefix: /bar
          # exact: /bar
        method:
          regxp: /bar.*
          #prefix: /bar
          #exact: /bar
        headers:
          FOO:
            regxp: /bar.*
            #prefix: /bar
            #exact: /bar
        cookie:
          USER:
            regxp: /bar.*
            #prefix: /bar
            #exact: /bar
      to:  # Specify destination
      - service: service-foo
        revision: v0
        namespace: default
        weight: 50
      - service: service-foo
        revision: v1
        namespace: default
        weight: 50
      redirect: # Specify redirect rule
        host: www.foo.bar
        path: /redirect
      rewrite:
        host: www.foo.bar
        path: /rewrite
      headers: # Header operations
        add:
          foo: bar
        set:
          foo: bar
        remove:
        - foo
      fault:
        percentage: 80 # Inject fault percentage(%)
        delayMillis: 100 # Adding delay before injecting fault (millseconds)
        abort:
          httpStatus: 502 # Injecting http code
      mirror:   # Sending mirror traffic
        service: mirror-foo
        revision: v0
        namespace: default
      timeoutMillis: 100 # Setting request timeout (milli-seconds)
      retry:
        attempts: 10 # Retry attempts
        timeoutMillis: 100 # Retry timeout (milli-seconds)
        
# Externalservices
externalservices:
  foo:
    ipAddresses: # Pointing to external IP addresses
    - 1.1.1.1
    - 2.2.2.2
    fqdn: www.foo.bar # Pointing to fqdn
    service: $namespace/$name # Pointing to services in another namespace
``` 

#### Watching Riofile
You can setup Rio to watch for Riofile changes in a Github repository and deploy Riofile changes automatically. For example:
```bash
$ rio up https://github.com/username/repo
```
If you want to setup webhook to watch, go to [here](#setup-github-webhook-experimental)


## Monitoring
By default, Rio will deploy [Grafana](https://grafana.com/) and [Kiali](https://www.kiali.io/) to give users the ability to watch all metrics of the service mesh.
You can find endpoints of both services by running `rio -s ps`, and then access these services through their endpoint URLs.

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

## Autoscaling
By default each workload is enabled with autoscaling (min scale 1, max scale 10), which means the workload can be scaled from 1 instance to 10 instances
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

## Continuous Delivery (Source code to Deployment)

Rio supports continous delivery from git-based source code repository to deploy the actual workload. Rio will watch for changes in the git repository, automatically build Docker images, and deploy new versions of the application.

To utilize the continous delivery feature, give Rio a valid git repository URL.

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
a specific branch (default is master). By default, Rio will pull and check the branch at a 15 second interval, but
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

### Setup credential for private repository
1. Set up git basic auth. (Currently ssh key is not supported and will be added soon.) Here is an exmaple of adding Github repo.
```bash
$ rio secret add --git-basic-auth
Select namespace[default]: $(put the same namespace with your workload)
git url: https://github.com/username
username: $username
password: $password
```
2. Run your workload and point it to your private git repo. It will automatically use the secret you just configured.

### Setup Github webhook (experimental)
By default, rio will automatically pull git repo and check if repo code has changed. You can also configure a webhook to automatically push any events to Rio to trigger the build.

1. Set up Github webhook token.
```bash
$ rio secret add --github-webhook
Select namespace[default]: $(put the same namespace with your workload)
accessToken: $(github_accesstoken) # the token has to be able create webhook in your github repo.
```

2. Create workload and point to your repo.

3. Go to your Github repo, it should have webhook configured to point to one of our webhook service.

### Set Custom build arguments and docker registry
You can also push to your own registry for images that rio has built.

1. Setup docker registry auth. Here is an example of how to setup docker registry.
```bash
$ rio secret add --docker
Select namespace[default]: $(put the same namespace with your workload)
Registry url[]: https://index.docker.io/v1/
username[]: $(your_docker_hub_username)
password[]: $(password)
```

To confirm this worked you can decode the contents of generated secret:

```bash
kubectl get secret dockerconfig-pull --output="jsonpath={.data.\.dockerconfigjson}" | base64 --decode
```

2. Create your workload. Set the correct push registry.

```bash
$ rio run --build-registry docker.io --build-image-name $(username)/yourimagename $(repo)
```
`docker.io/$(username)/yourimagename` will be pushed into dockerhub registry.

If you'd like to pull your image from a private repository, rather than push an image, setup your docker secret as above then pull and run your image with:

```bash
$ rio run --image-pull-secrets=dockerconfig-pull $(repo)
```

### Enable Pull request (experimental)
Rio also allows you to configure pull request builds. This needs you to configure github webhook token correctly.

1. Set up github webhook token in the previous session

2. Run workload with pull-request enabled.

```bash
$ rio run --build-enable-pr $(repo)
```

After this, if there is any pull request, Rio will create a deployment based on this pull request, and you will get a unique link
to see the change this pull request introduced in the actual deployment.

### View build logs
To view logs from your builds
```bash
$ rio builds
NAME                                                                     SERVICE                   REVISION                                   CREATED        SUCCEED   REASON
default/fervent-swartz6-ee709-786b366d5d44de6b547939f51d467437e45c5ee1   default/fervent-swartz6   786b366d5d44de6b547939f51d467437e45c5ee1   23 hours ago   True    

$ rio logs -f default/fervent-swartz6-ee709-786b366d5d44de6b547939f51d467437e45c5ee1

# restart any builds that failed
$ rio build restart default/fervent-swartz6-ee709-786b366d5d44de6b547939f51d467437e45c5ee1
```

## Local Developer Setup

Rio supports a local developer setup where source code is stored locally. Rio by default uses buildkit to build application.

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

If you want more complex build arguments, rio supports the following format
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

* How can I reference persist volume?
```
Rio only supports stateless workloads at this point.
```

* How to manually specify IP addresses?
```
Rio will automatically detect work node ip addresses based on install mode. If your host has multiple IP addresses, you can manually specify which IP address Rio should use for creating external DNS records with the `--ip-address` flag. 
For instance to advertise the external IP of an AWS instance: `rio install --ip-address $(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)`
By doing this, you lose the ability to dynamic updating IP addresses to DNS.
```

