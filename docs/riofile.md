# Riofile

## Table of Contents

- [Intro](#intro)
  - [Watching a Riofile](#watching-a-riofile)
- [Reference](#riofile-reference)
- [Templating](#templating)
  - [Using answer file](#using-answer-file)
  - [Using environment substitution](#using-environment-substitution)
- [Examples](#examples)
  - [How to use Rio to deploy an application with arbitary YAML](#How-to-use-Rio-to-deploy-an-application-with-arbitary-YAML)
  - [How to watch a repo](#How-to-watch-a-repo)


## Intro


Rio works with standard Kubernetes YAML files. Rio additionally supports a more user-friendly `docker-compose`-style config file called `Riofile`.
This allows you define rio services, routes, external services, configs, and secrets.

For example, here is an nginx application:

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

Once you have defined `Riofile`, simply run `rio up`.
If you made any change to the `Riofile`, re-run `rio up` to pick up the change.
To use a file not named "Riofile" use `rio up -f nginx.yaml`.

#### Watching a Riofile

You can setup Rio to watch for Riofile changes in a Github repository and deploy Riofile changes automatically. For example:
```bash
$ rio up https://github.com/username/repo
```

By default, Rio will poll the branch in 30 second intervals, but this can be configured to use a webhook instead. See [Webhook docs](./webhooks.md) for info.

## Riofile Reference

```yaml
# Configmap
configs:
  config-foo:     # specify name in the section
    key1: |-      # specify key and data in the section
      {{ config1 }}
    key2: |-
      {{ config2 }}

# Externalservices
externalservices:
  foo:
    ipAddresses: # Pointing to external IP addresses
    - 1.1.1.1
    - 2.2.2.2
    fqdn: www.foo.bar # Pointing to fqdn
    service: $namespace/$name # Pointing to services in another namespace

# Service
services:
  service-foo:
    app: my-app # Specify app name. Defaults to service name. This is used to aggregate services that belongs to the same app.
    version: v0 # Specify the version of app this service represents. Defaults to v0. Displayed as app@version, unless version is v0 where it will be omitted.
    scale: 2 # Specify scale of the service, defaults to 1. Use this to have service come up with the specified number of pods.
    template: false # Set this service as a template to build service versions from instead of overwriting on each build, false by default. See https://github.com/rancher/rio/blob/master/docs/continuous-deployment.md#automatic-versioning
    weight: 80 # Percentage of weight assigned to this revision. Defaults to 100.

    # To enable autoscaling:
    autoscale:
      concurrency: 10 # specify concurrent request each pod can handle(soft limit, used to scale service)
      maxReplicas: 10
      minReplicas: 1

    # Traffic rollout config. Optional
    rollout:
      increment: 5 # traffic percentage increment(%) for each interval.
      interval: 2 # traffic increment interval(seconds).
      pause: false # whether to perform rollout or not

    # Container configuration
    image: nginx # Container image. Required if not setting build
    imagePullPolicy: always # Image pull policy. Options: (always/never/ifNotProsent), defaults to ifNotProsent.
    build: # Setting build parameters. Set if you want to build image from source
      repo: https://github.com/rancher/rio # Git repository to build. Required
      branch: master # Git repository branch. Required
      revision: v0.1.0 # Revision digest to build. If set, image will be built based on this revision. Otherwise it will take head revision in repo. Also if revision is not set, it will be served as the base revision to watch any change in repo and create new revision based changes from repo.
      args: # Build arguments to pass to buildkit https://docs.docker.com/engine/reference/builder/#understand-how-arg-and-from-interact. Optional
      - foo=bar
      dockerfile: Dockerfile # The name of Dockerfile to look for.  This is the full path relative to the repo root. Defaults to `Dockerfile`.
      context: ./  # Docker build context. Defaults to .
      noCache: true # Build without cache. Defaults to false.
      imageName: myname/image:tag # Specify custom image name(excluding registry name). Default name: $namespace/name:$revision_digest
      pushRegistry: docker.io # Specify push registry. Example: docker.io, gcr.io. Defaults to localhost registry.
      pushRegistrySecretName: secretDocker # Specify secret name for pushing to docker registry. [link](#set-custom-build-arguments-and-docker-registry)
      stageOnly: true # If set, newly created revision will not get any traffic. Defaults to false.
      webhookSecretName: secretGithub # Specify the github secret name. Used to create Github webhook, the secret key has to be `accessToken`
      cloneSecretName: secretGit # Specify secret name for checking our git resources
      pr: true # Enable pull request feature. Defaults to false
      tag: false # Optionally enable to build off every tag release in the repo
      tagInclude: ^v # If tag is true, only use tags matching this pattern
      tagExclude: rc # If tag is true, exclude any tags with this pattern
      timeout: 10 # build timeout setting in seconds
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
    memory: 100Mi # Memory request. 100Mi, available options. If not set, memory request will not be set. https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
    secrets: # Specify secret to mount. Format: `$name/$key:/path/to/file`. Secret has to be pre-created in the same namespace
    - foo/bar:/my/password
    configs: # Specify configmap to mount. Format: `$name/$key:/path/to/file`.
    - foo/bar:/my/config
    livenessProbe: # LivenessProbe setting. https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/
      httpGet:
        path: /ping
        port: "9997" # port must be string
      initialDelaySeconds: 10
    readinessProbe: # ReadinessProbe https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/
      failureThreshold: 7
      httpGet:
        path: /ready
        port: "9997" # port must be string
    stdin: true # Whether this container should allocate a buffer for stdin in the container runtime
    stdinOnce: true # Whether the container runtime should close the stdin channel after it has been opened by a single attach. When stdin is true the stdin stream will remain open across multiple attach sessions.
    tty: true # Whether this container should allocate a TTY for itself
    runAsUser: 1000 # The UID to run the entrypoint of the container process.
    runAsGroup: 1000 # The GID to run the entrypoint of the container process
    readOnlyRootFilesystem: true # Whether this container has a read-only root filesystem
    privileged: true # Run container in privileged mode.

    nodeAffinity: # Describes node affinity scheduling rules for the pod.
    podAffinity:  # Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)).
    podAntiAffinity: # Describes pod anti-affinity scheduling rules (e.g. avoid putting this pod in the same node, zone, etc. as some other pod(s)).

    hostAliases: # Hostname alias
      - ip: 127.0.0.1
        hostnames:
        - example.com
    hostNetwork: true # Use host networking, defaults to False. If this option is set, the ports that will be used must be specified.
    imagePullSecrets: # Image pull secret https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
    - secret1
    - secret2

    # Containers: Specify sidecars. Other options are available in container section above, this is limited example
    containers:
    - init: true # Init container
      name: my-init
      image: ubuntu
      args:
      - "echo"
      - "hello world"


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

# Router
routers:
  foo:
    routes:
      - match: # Match rules, the first rule matching an incoming request is used
          path: # Match path, can specify regxp, prefix or exact match
            exact: /v0
          # prefix: /bar
          # regxp: /bar.*
          methods:
          - GET
          headers:
            - name: FOO
              value:
                regxp: /bar.*
                #prefix: /bar
                #exact: /bar
        to:  # Specify destination
          - app: myapp
            version: v1
            port: 80
            namespace: default
            weight: 50
          - app: myapp
            version: v2
            port: 80
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
            - name: foo
              value: bar
          set:
            - name: foo
              value: bar
          remove:
          - foo
        fault:
          percentage: 80 # Inject fault percentage(%)
          delayMillis: 100 # Adding delay before injecting fault (millseconds)
          abortHTTPStatus: 502 # Injecting http code
        mirror:   # Sending mirror traffic
          app: q
          version: v0
          namespace: default
        timeoutSeconds: 1 # Setting request timeout (milli-seconds)
        retry:
          attempts: 10 # Retry attempts
          timeoutSeconds: 1 # Retry timeout (milli-seconds)

# Use Riofile's answer/question templating
# When you define NAMESPACE and REVISION variables in questions section rio will automatically inject their values.
template:
  goTemplate: true # use go templating
  envSubst: true # use ENV vars during templating
  questions:  # now make some questions that we provide answers too
  - variable: FOO # This will be available with go templates in the field `.Values.FOO`
    description: "Some question we want an answer to."

# Supply arbitrary kubernetes manifest yaml
kubernetes:
  manifest: |-
    apiVersion: apps/v1
    kind: Deployment
    ....

```

## Templating

#### Using answer file

Rio allows the user to leverage an answer file to customize `Riofile`.
Go template and [envSubst](https://github.com/drone/envsubst) can be used to apply answers. By default, the `NAMESPACE` and `REVISION` variables are available when defined in the template questions.

Answer file is a yaml manifest with key-value pairs:

```yaml
FOO: BAR
```

For example, to use go templating to apply a service when provided with the above answers file and running in the `test` namespace:

1. Create Riofile
```yaml
{{- if (and (eq .Values.NAMESPACE "test") (eq .Values.FOO "BAR")) }}
services:
  demo:
    image: ibuildthecloud/demo:v1
    ports:
    - 80
{{- end}}

template:
  goTemplate: true # use go templating
  envSubst: true # use ENV vars during templating
  questions:  # now make some questions that we provide answers too
  - variable: FOO
    description: "My custom thing"
  - variable: NAMESPACE
    description: "The namespace"
```
2. `kubectl create namespace test`
3. `cd /path/to/Riofile && rio -n test up --answers answers.yaml`

Rio also supports a bash style envsubst replacement, with the following format:

```yaml
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

#### Using environment substitution

```yaml
services:
  demo:
    image: ibuilthecloud/demo:v1
    env:
    - FOO=${FOO}
    - MYNAMESPACE=${NAMESPACE}
  
template:
  goTemplate: true # use go templating
  envSubst: true # use ENV vars during templating  
  questions:  # now make some questions that we provide answers too
  - variable: FOO
    description: "My custom thing"
```


## Examples

#### How to use Rio to deploy an application with arbitary YAML

In this example we will see how to define both a normal Rio service and arbitary Kubernetse manifests and deploy both of these with Rio. Follow the quickstart to get Rio installed into your cluster and ensure the output of `rio info` looks similar to this:
```
Rio Version:  >=0.6.0
Rio CLI Version: >=0.6.0
Cluster Domain: enu90s.on-rio.io
Cluster Domain IPs: <cluster domain ip>
System Namespace: rio-system
System Ready State: true
Wildcard certificates: true

System Components:
gateway-v2 status: Ready
rio-controller status: Ready
```

First, lets use a Riofile to define a a basic Rio service.

```
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

Next, we can augment this service with the Kubernetes [sample guestbook](https://kubernetes.io/docs/tutorials/stateless-application/guestbook/)


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

kubernetes:
    manifest: |-
      apiVersion: v1
      kind: Service
      metadata:
        name: redis-master
        labels:
          app: redis
          tier: backend
          role: master
      spec:
        ports:
        - port: 6379
          targetPort: 6379
        selector:
          app: redis
          tier: backend
          role: master
      ---
      apiVersion: apps/v1 #  for k8s versions before 1.9.0 use apps/v1beta2  and before 1.8.0 use extensions/v1beta1
      kind: Deployment
      metadata:
        name: redis-master
      spec:
        selector:
          matchLabels:
            app: redis
            role: master
            tier: backend
        replicas: 1
        template:
          metadata:
            labels:
              app: redis
              role: master
              tier: backend
          spec:
            containers:
            - name: master
              image: k8s.gcr.io/redis:e2e  # or just image: redis
              resources:
                requests:
                  cpu: 100m
                  memory: 100Mi
              ports:
              - containerPort: 6379
      ---
      apiVersion: v1
      kind: Service
      metadata:
        name: redis-slave
        labels:
          app: redis
          tier: backend
          role: slave
      spec:
        ports:
        - port: 6379
        selector:
          app: redis
          tier: backend
          role: slave
      ---
      apiVersion: apps/v1 #  for k8s versions before 1.9.0 use apps/v1beta2  and before 1.8.0 use extensions/v1beta1
      kind: Deployment
      metadata:
        name: redis-slave
      spec:
        selector:
          matchLabels:
            app: redis
            role: slave
            tier: backend
        replicas: 2
        template:
          metadata:
            labels:
              app: redis
              role: slave
              tier: backend
          spec:
            containers:
            - name: slave
              image: gcr.io/google_samples/gb-redisslave:v1
              resources:
                requests:
                  cpu: 100m
                  memory: 100Mi
              env:
              - name: GET_HOSTS_FROM
                value: dns
                # If your cluster config does not include a dns service, then to
                # instead access an environment variable to find the master
                # service's host, comment out the 'value: dns' line above, and
                # uncomment the line below:
                # value: env
              ports:
              - containerPort: 6379
      ---
      apiVersion: v1
      kind: Service
      metadata:
        name: frontend
        labels:
          app: guestbook
          tier: frontend
      spec:
        # if your cluster supports it, uncomment the following to automatically create
        # an external load-balanced IP for the frontend service.
        # type: LoadBalancer
        ports:
        - port: 80
        selector:
          app: guestbook
          tier: frontend
      ---
      apiVersion: apps/v1 #  for k8s versions before 1.9.0 use apps/v1beta2  and before 1.8.0 use extensions/v1beta1
      kind: Deployment
      metadata:
        name: frontend
      spec:
        selector:
          matchLabels:
            app: guestbook
            tier: frontend
        replicas: 3
        template:
          metadata:
            labels:
              app: guestbook
              tier: frontend
          spec:
            containers:
            - name: php-redis
              image: gcr.io/google-samples/gb-frontend:v4
              resources:
                requests:
                  cpu: 100m
                  memory: 100Mi
              env:
              - name: GET_HOSTS_FROM
                value: dns
                # If your cluster config does not include a dns service, then to
                # instead access environment variables to find service host
                # info, comment out the 'value: dns' line above, and uncomment the
                # line below:
                # value: env
              ports:
              - containerPort: 80
```

Typically you would track your Riofile with some form of VCS but for now simply save it in a local directory.

Next, run `rio up` in that directory.

You can watch Rio service come up with `rio ps` and the Kubernetes deployments with `kubectl get deployments -w`.

You can check that the sample service came up by going to the endpoint given by `rio ps`
```
NAME      IMAGE     ENDPOINT                                          SCALE     APP       VERSION    WEIGHT    CREATED       DETAIL
nginx     nginx     https://nginx-2c21baa1-default.enu90s.on-rio.io   1         nginx     2c21baa1   100%      4 hours ago
```

We can use Rio to expose the service and provision a LetsEncrypt certificate for it.

` rio router add guestbook to frontend,port=80 `

This will create a route to the service and create an endpoint.

```
rio endpoints
NAME        ENDPOINTS
nginx       https://nginx-default.enu90s.on-rio.io
guestbook   https://guestbook-default.enu90s.on-rio.io
```

We can now access this endpoint over encrypted https!


#### How to watch a repo

```bash
namespace=something
kubectl create namespace ${namespace}
# based on your situation, you may have application level secrets to be created
# kubectl apply -f deployments/secrets
# if you have a private git repo, docker registry, you can use the above approach or you can create them using rio itself
# if you have build instructions and want to push to a private registry
# rio secrets create --docker
# since you wish to watch a repo for changes on the stack definition, rio must have the ability to create a webhook in your repository
# rio secrets create --github-webhook 
# try rio secrets create --help to see other options.
# make sure you pass --build-clone-secret gitcredential if you are having a private git repo
# see a nodejs example here https://github.com/lucidprogrammer/rio-samples.git

rio -n ${namespace} up --name somename  --push-registry-secret dockerconfig \
    --file deployments/my-stack.yaml --build-webhook-secret githubtoken\
    --answers deployments/values-my-stack.json https://github.com/lucidprogrammer/rio-samples.git

# you should be able to see the autoscaling in action with the following

hey -z 600s -c 60 endpointurlofweb1service
```

```yaml
template:
  goTemplate: true
  envSubst: true
  questions:
    - variable: NAMESPACE
      description: "namespace to deploy to"
    - variable: REVISION
      description: "Current commit"
    # make sure you respect the type of the variable used.
    - variable: MAX_SCALE
      type: "int"
      description: "maximum scale number"
services:
  web1:
    version: v0
    weight: 100
    stageOnly: false
    autoscale:
      concurrency: 10
      maxReplicas: ${MAX_SCALE}
      minReplicas: 1
    ports:
    - 80:3000/http
    env:
    - NAMESPACE=${NAMESPACE}
    # say you want to have a way to show the version/commit in your application.
    - REVISION=${REVISION}
    # let's say you want to access a super secret as env, you may do as follows
    # - REDIS_PASSWORD=secret://spec/REDIS_PASSWORD
    build:
      branch: master
      context: ./src/web1
      pushRegistry: docker.io
      # everytime you do a push to the repo, REVISION will match the specific commit hash and an image is created and pushed.
      imageName: lucidprogrammer/web1:${REVISION}
  
  # an example to use init containers
  web2:
    version: v0
    weight: 100
    ports:
    - 80:80/http
    image: nginx:alpine
    volumes:
    - name: html
      path: /usr/share/nginx/html
    containers:
    - init: true
      name: web2-init
      image: busybox
      volumes:
      - name: html
        path: /work-dir
      command:
        - wget
        - "-O"
        - "/work-dir/index.html"
        - http://kubernetes.io
```
