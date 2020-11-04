# CLI Reference


## Table of Contents

- [attach](#attach)
- [build](#build)
- [build-history](#build-history)
- [cat](#cat)
- [dashboard](#dashboard)
- [edit](#edit)
- [exec](#exec)
- [export](#export)
- [image](#image)
- [info](#info)
- [inspect](#inspect)
- [install](#install)
- [kill](#kill)
- [logs](#logs)
- [promote](#promote)
- [ps](#ps)
- [router](#router)
- [run](#run)
- [rm](#rm)
- [scale](#scale)
- [stage](#stage)
- [system-logs](#system-logs)
- [system-features](#system-feature)
- [uninstall](#uninstall)
- [up](#up)
- [weight](#weight)



## attach

Attach to a process running in a container

##### Usage
```
rio attach [OPTIONS] CONTAINER
```

##### Options
| flag            | aliases | description                                                  | default |
|-----------------|---------|--------------------------------------------------------------|---------|
| --timeout value |         | Timeout waiting for the container to be created to attach to | 1m      |
| --pod value     |         | Specify pod, default is first pod found                      |         |

##### Examples

```shell script
rio attach demo

rio attach --timeout 30s --pod mydemopod demo
```

---

## build

Build a docker image using buildkitd

##### Usage

```
rio build command [command options] [arguments...]
```

##### Options

| flag         | aliases | description                                                             | default |
|--------------|---------|-------------------------------------------------------------------------|---------|
| --file value | -f      | Name of the file to look for build, support both Riofile and Dockerfile |         |
| --tag value  | -t      | Name and optionally a tag in the 'name:tag' format                      |         |
| --build-arg  |         | Set build-time variables                                                |         |
| --no-cache   |         | Do not use cache when building the image                                |         |
| --help       | -h      | show help                                                               |         |


##### Examples

```shell script

# Navigate to directory with Dockerfile and build it into local registry
rio build -t test:v1

# See image that was build
rio image

# Build from riofile insted
rio build -t test:v1 --no-cache -f Riofile.yaml

# Now run the image
rio run -n test -p 8080 localhost:5442/default/test:v1

# Build the image again with new tag
rio build -t test:v2

# Now stage the 2nd image
rio stage --image localhost:5442/default/test:v2 test v2
```

---

## build-history

Show previous builds

##### Usage

```
rio build-history [command options] [arguments...]
```

##### Options

| flag           | aliases | description                                                  | default |
|----------------|---------|--------------------------------------------------------------|---------|
| --quiet        | -q      | Only display Names                                           |         |
| --format value |         | 'json' or 'yaml' or Custom format: {{ "'{{.Obj.Name}}'" }} [$FORMAT] |         |


##### Examples

```shell script
# see previous builds from stacks or workloads
rio build-history

# custom output format
rio build-history --format {{ "{{.Obj.Name}}" }}
```

---

## cat

Print the contents of a config

##### Usage
```
rio cat [OPTIONS] [NAME...]
```

##### Options

| flag  | aliases | description             | default |
|-------|---------|-------------------------|---------|
| --key | -k      | The values which to cat |         |

##### Examples

```shell script
# cat a configmap
rio cat configmap/config-foo

# cat a key from a configmap
rio cat --key=a configmap/config-foo
```

---

## dashboard

Open the dashboard in a browser

##### Usage
```
rio dashboard [OPTIONS]
```

##### Options

| flag          | aliases | description          | default |
|---------------|---------|----------------------|---------|
| --reset-admin |         | Reset admin password |         |


##### Examples

```shell script
# reset admin pw
rio dashboard --reset-admin 
```

---

## edit

Edit resources

##### Usage
```
rio edit [TYPE/]RESOURCE_NAME
```

##### Options

| flag  | aliases | description                                           | default |
|-------|---------|-------------------------------------------------------|---------|
| --raw |         | Edit the raw API object, not the pretty formatted one |         |


##### Examples

```shell script
rio edit demo@v4

rio edit router/myrouter
```

---

## exec

Run a command in a running container

##### Usage
```
rio exec [OPTIONS] CONTAINER COMMAND [ARG...]
```

##### Options

| flag              | aliases  | description                                          | default |
|-------------------|----------|------------------------------------------------------|---------|
| --stdin           | -i       | Pass stdin to the container                          |         |
| --tty             | -t       | Stdin is a TTY                                       |         |
| --container value | -c value | Specify container in pod, default is first container |         |
| --pod value       |          | Specify pod, default is first pod found              |         |

##### Examples

```shell script
# ssh into running container
rio exec -it demo sh

# this is equivalent of doing
rio exec --tty --stdin demo sh

# choose pod and container
rio exec -it --pod mypod --container server demo sh
```

---

## export

Export a namespace or service

##### Usage
```
rio export [TYPE/]NAMESPACE_OR_SERVICE
```

##### Options

| flag           | aliases | description                                        | default |
|----------------|---------|----------------------------------------------------|---------|
| --format value |         | Specify output format, yaml/json. Defaults to yaml | yaml    |
| --riofile      |         | Export riofile format                              |         |


##### Examples

```shell script
# export a service
rio export demo

# export a namespace in riofile format
rio export --riofile namespace/default
```

---

## image

List images built from the local registry

##### Usage
```
rio image
```

---

## info

Show system info

##### Usage
```
rio info
```

---

## inspect

Inspect resources

##### Usage
```
rio inspect [TYPE/][NAMESPACE/]SERVICE_NAME
```

##### Options

| flag     | aliases | description                                           | default |
|----------|---------|-------------------------------------------------------|---------|
| --format |         | Edit the raw API object, not the pretty formatted one |         |


##### Examples

```shell script
rio inspect svc@v2

# inspect a build
rio inspect taskrun/affectionate-mirzakhani-mfp5q-ee709-4e40c
```

---

## install

Install the Rio management plane

See the [install docs](install.md) for more info.

##### Usage
```
rio install [OPTIONS]
```

##### Options

| flag                     | aliases | description                                                                            | default                         |
|--------------------------|---------|----------------------------------------------------------------------------------------|---------------------------------|
| --check                  |         | Only check status, don't deploy controller                                             |                                 |
| --disable-features value |         | Manually specify features to disable, supports comma separated values                  |                                 |
| --enable-debug           |         | Enable debug logging in controller                                                     |                                 |
| --ip-address value       |         | Manually specify IP addresses to generate rdns domain, supports comma separated values |                                 |
| --yaml                   |         | Only print out k8s yaml manifest                                                       |                                 |
| --rdns-url               |         | Specify Rdns server url to use                                                         | https://api.on-rio.io/v1        |


**--check**

Check if Rio is installed in the current cluster without deploying the Rio controller.
If Rio has not been installed, this command might hang on `Waiting for rio controller to initialize`.

**--disable-features**

Choose features to be disabled when starting the Rio control plane. Below are a list of available features

| Feature     | Description                                       |
|-------------|---------------------------------------------------|
| autoscaling | Auto-scaling services based on in-flight requests |
| build       | Rio Build, from source code to deployment         |
| gloo        | API gateway backed by gloo                        |
| linkerd     | Linkerd service mesh                              |
| letsencrypt | Let's Encrypt                                     |
| rdns        | Acquire DNS from public Rancher DNS service       |
| dashboard   | Rio UI                                            |

**--ip-address**

Manually specify IPAddress for API gateway services. The IP will be used to generate a record for the cluster domain.
By default, if this flag is not specified, Rio will use the IP of [Service Loadbalancer](https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/) that points to API gateway.

Note: If service loadbalancer cannot be provisioned, [Nodeport](https://kubernetes.io/docs/concepts/services-networking/service/#nodeport) is used to expose API gateway.

##### Examples

```shell script
# basic install
rio install

# install with debug and disable some features
rio install --enable-debug --disable-features linkerd,gloo

# print yaml to run manually, with custom ip-address
rio install --yaml --ip-address 127.0.0.1
```


---

## kill

Kill pods individually or all pods belonging to a service

##### Usage
```
rio kill [SERVICE_NAME/POD_NAME]
```

##### Examples

```shell script
# kill a service
rio kill demo

# kill individual pods
rio pods # first get pod name
rio kill pod/demo-v042dxp-5fb7d8f677-f9xgn
```

---


## logs

Print logs from services or containers

##### Usage
```
rio logs [OPTIONS] SERVICE/BUILD
```

##### Options

| flag              | aliases  | description                                                                                       | default |
|-------------------|----------|---------------------------------------------------------------------------------------------------|---------|
| --since value     | -s value | Logs since a certain time, either duration (5s, 2m, 3h) or RFC3339                                | "24h"   |
| --timestamps      | -t       | Print the logs with timestamp                                                                     |         |
| --tail value      | -n value | Number of recent lines to print, -1 for all                                                       | 200     |
| --container value | -c value | Print the logs of a specific container, use -a for system containers                              |         |
| --previous        | -p       | Print the logs for the previous instance of the container in a pod if it exists, excludes running |         |
| --init-containers |          | Include or exclude init containers                                                                |         |
| --all             | -a       | Include hidden or systems logs when logging                                                       |         |
| --no-color        | --nc     | Dont show color when logging                                                                      |         |
| --output value    | -o value | Output format: [default, raw, json]                                                               | default |


##### Examples

```shell script
# get logs from a service
rio logs demo

# Get logs from a build
rio build-history
rio logs taskrun/affectionate-mirzakhani-mfp5q-ee709-4e40c

# get 1 previous log line for the linkerd-proxy in demo service
rio logs --tail 1 --container linkerd-proxy -a demo

# ignore init-containers and filter to waiting or terminated pods, include timestamps
rio logs --container-state "terminated,waiting" --init-containers=false --timestamps demo

# target terminated pods of all kinds, format as json
rio logs -p -a  --output json demo
```

---

## Promote

Send 100% of traffic to an app version and scale down other versions. See also weight. 

##### Usage

```
rio promote [OPTIONS] SERVICE_NAME
```

##### Options

| flag       | aliases | description                                                                   | default |
|------------|---------|-------------------------------------------------------------------------------|---------|
| --duration | none    | How long the rollout should take. An approximation, actual time may fluctuate | 0s      |
| --pause    | none    | Whether to pause all rollouts on current app                                  | false   |

##### Examples

```shell script
# promote n@v2 
rio promote n@v2

# promote n@v2 over 1 hour 
rio promote --duration=1h n@v2

# pause last command
rio promote --pause=true n@v2
```


---

## ps

List services

##### Usage
```
rio ps [OPTIONS]
```

##### Options

| flag        | aliases | description                                                            | default |
|-------------|---------|------------------------------------------------------------------------|---------|
| --quiet     | -q      | Only display Names                                                     |         |
| --format    |         | 'json' or 'yaml' or Custom format: {{ "'{{.Name}} {{.Obj.Name}}'" }} [$FORMAT] |         |
| --all       | -a      | print all resources, including router and externalservice              |         |
| --workloads | -w      | include apps/v1 Deployments and DaemonSets in output                   |         |


##### Examples

```shell script
# show services and workloads
rio ps -w

# output json
rio ps --format json

# display name and weight in custom format
rio ps --format {{ "{{.Obj.Name}} -> {{.Data.Weight}}" }}
```

---

## router

Route traffic across the mesh

##### Usage
```
rio routers command [command options] [arguments...]
```

##### Options

| flag        | aliases | description                                                            | default |
|-------------|---------|------------------------------------------------------------------------|---------|
| --quiet     | -q      | Only display Names                                                     |         |
| --format    |         | 'json' or 'yaml' or Custom format: {{ "'{{.Name}} {{.Obj.Name}}'" }} [$FORMAT] |         |


##### Examples

```shell script
# show existing routers
rio route
```

#### add/create

Create a router. By default appends at the end.

Services specified without a version are assumed to be apps. For example `rio route add x to svc` would target the svc app endpoint,
not the `svc@v0` version.

##### Usage
```
rio router create/add MATCH ACTION [TARGET...]
```

##### Options

| flag                              | aliases | description                                          | default |
|-----------------------------------|---------|------------------------------------------------------|---------|
| --insert                          |         | Insert the rule at the beginning instead of the end  |         |
| --header value                    |         | Match HTTP header (format key=value, value optional) |         |
| --fault-percentage value          |         | Percentage of matching requests to fault             | 0       |
| --fault-delay-milli-seconds value |         | Inject a delay for fault in milliseconds             | 0       |
| --fault-httpcode value            |         | HTTP code to send for fault injection                | 0       |
| --add-header value                |         | Add HTTP header to request (format key=value)        |         |
| --set-header value                |         | Override HTTP header to request (format key=value)   |         |
| --remove-header value             |         | Remove HTTP header to request (format key=value)     |         |
| --retry-attempts value            |         | How many times to retry                              | 0       |
| --retry-timeout-seconds value     |         | Timeout per retry in seconds                         | 0       |
| --timeout-seconds value           |         | Timeout in seconds for all requests                  | 0       |
| --method value                    |         | Match HTTP method, support comma-separated values    |         |

##### Examples

```shell script

# route to the demo app endpoint
rio route add myroute to demo

# route a specific path to the demo app's version 0, and insert into first slot
rio route add --insert myroute/name.html to demo@v0
```

See the [routers readme](router.md) for advanced example usage.


---

## run

Create and run a new service

##### Usage
```
rio run [OPTIONS] IMAGE [COMMAND] [ARG...]
```

##### Options

| flag                             | aliases  | description                                                                                                                                   | default                |
|----------------------------------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------|------------------------|
| --add-host value                 |          | Add a custom host-to-IP mapping (host=ip)                                                                                                     |                        |
| --annotations value              |          | Annotations to attach to this service                                                                                                         |                        |
| --build-branch value             |          | Build repository branch                                                                                                                       | master                 |
| --build-dockerfile value         |          | Set Dockerfile name                                                                                                                           | defaults to Dockerfile |
| --build-context value            |          | Set build context                                                                                                                             | .                      |
| --build-webhook-secret value     |          | Set GitHub webhook secret name                                                                                                                |                        |
| --build-docker-push-secret value |          | Set docker push secret name                                                                                                                   |                        |
| --build-clone-secret value       |          | Set git clone secret name                                                                                                                     |                        |
| --build-image-name value         |          | Specify custom image name to push                                                                                                             |                        |
| --build-registry value           |          | Specify to push image to                                                                                                                      |                        |
| --build-revision value           |          | Build git commit or tag                                                                                                                       |                        |
| --build-pr                       |          | Enable builds on new pull requests                                                                                                            |                        |
| --build-tag                      |          | Enable builds on any new tags instead of new commits on a branch, requires webhook, does not support polling                                  |                        |
| --build-tag-include              |          | Pattern that tags must match                                                                                                                  |                        |
| --build-tag-exclude              |          | Pattern that excludes tags                                                                                                                    |                        |
| --build-timeout value            |          | Timeout for build, ( (ms/s/m/h))                                                                                                              | 10m                    |
| --command value                  |          | Overwrite the default ENTRYPOINT of the image                                                                                                 |                        |
| --config value                   |          | Configs to expose to the service (format: name[/key]:target)                                                                                  |                        |
| --concurrency value              |          | The maximum concurrent request a container can handle (autoscaling)                                                                           | 10                     |
| --cpus value                     |          | Number of CPUs                                                                                                                                |                        |
| --dns value                      |          | Set custom DNS servers                                                                                                                        |                        |
| --dnsoption value                |          | Set DNS options (format: key:value or key)                                                                                                    |                        |
| --dnssearch value                |          | Set custom DNS search domains                                                                                                                 |                        |
| --env value                      | -e value | Set environment variables                                                                                                                     |                        |
| --env-file value                 |          | Read in a file of environment variables                                                                                                       |                        |
| --global-permission value        |          | Permissions to grant to container's service account for all namespaces                                                                        |                        |
| --group value                    |          | The GID to run the entrypoint of the container process                                                                                        |                        |
| --health-cmd value               |          | Command to run to check health                                                                                                                |                        |
| --health-failure-threshold value |          | Consecutive failures needed to report unhealthy                                                                                               | 0                      |
| --health-header value            |          | HTTP Headers to send in GET request for healthcheck                                                                                           |                        |
| --health-initial-delay value     |          | Start period for the container to initialize before starting healthchecks ( (ms/s/m/h))                                                       | "0s"                   |
| --health-interval value          |          | Time between running the check ( (ms/s/m/h))                                                                                                  | "0s"                   |
| --health-success-threshold value |          | Consecutive successes needed to report healthy                                                                                                | 0                      |
| --health-timeout value           |          | Maximum time to allow one check to run ( (ms/s/m/h))                                                                                          | "0s"                   |
| --health-url value               |          | URL to hit to check health (example: http://:8080/ping)                                                                                       |                        |
| --host-dns                       |          | Use the host level DNS and not the cluster level DNS                                                                                          |                        |
| --hostname value                 |          | Container host name                                                                                                                           |                        |
| --image-pull-policy value        |          | Behavior determining when to pull the image (never/always/not-present)                                                                        | "not-present"          |
| --image-pull-secrets value       |          | Specify image pull secrets                                                                                                                    |                        |
| --interactive                    | -i       | Keep STDIN open even if not attached                                                                                                          |                        |
| --label-file value               |          | Read in a line delimited file of labels                                                                                                       |                        |
| --label value                    | -l value | Set meta data on a container                                                                                                                  |                        |
| --memory value                   | -m value | Memory reservation (format: <number>[<unit>], where unit = b, k, m or g)                                                                      |                        |
| --name value                     | -n value | Assign a name to the container. Use format [namespace:]name[@version]                                                                         |                        |
| --net value                      |          | Set network mode (host)                                                                                                                       |                        |
| --no-mesh                        |          | Disable service mesh                                                                                                                          |                        |
| --permission value               |          | Permissions to grant to container's service account in current namespace                                                                      |                        |
| --ports value                    | -p value | Publish a container's port(s) (format: svcport:containerport/protocol)                                                                        |                        |
| --privileged                     |          | Run container with privilege                                                                                                                  |                        |
| --read-only                      |          | Mount the container's root filesystem as read only                                                                                            |                        |
| --rollout-duration value         |          | How long the rollout should take. An approximation, actual time may fluctuate. Affects template services, but not weight or promote commands. | "0s"                   |
| --request-timeout-seconds value  |          | Set request timeout in seconds                                                                                                                | 0                      |
| --scale value                    |          | The number of replicas to run or a range for autoscaling (example 1-10)                                                                       |                        |
| --secret value                   |          | Secrets to inject to the service (format: name[/key]:target)                                                                                  |                        |
| --stage-only                     |          | Only stage service when generating new services. Can only be used when template is true                                                       |                        |
| --template                       |          | If true new version is created per git commit. If false update in-place                                                                       |                        |
| --tty                            | -t       | Allocate a pseudo-TTY                                                                                                                         |                        |
| --user value                     | -u value | UID[:GID] Sets the UID used and optionally GID for entrypoint process (format: <uid>[:<gid>])                                                 |                        |
| --volume value                   | -v value | Specify volumes for for services                                                                                                              |                        |
| --weight value                   |          | Specify the weight for the services                                                                                                           | 0                      |
| --workdir value                  | -w value | Working directory inside the container                                                                                                        |                        |

##### Examples

```shell script
# basic run
rio run -p 80 nginx

# run a named service with set scale, concurrency and ports. Build an image from a github repo
rio run -n mysvc --scale 5-10 --concurrency 5 -p 80:8080/http https://github.com/rancher/rio-demo

# add a version to service
rio run --weight 50 -n mysvc@v2 -p 80 nginx

# set custom readiness probe
rio run --health-url http://:8080/status --health-initial-delay 10s --health-interval 5s --health-failure-threshold 5 --health-timeout 5s -p 8080 cbron/mybusybox:dev

# set permission for containers. By setting permissions, rio will assign a serviceaccount to the pod which will have the corresponding permissions. Global permission means permissions across all namespaces.
rio run --global-permission "create,update,delete services" --permission "* apps/deployments" nginx

# set host:ip entry in container
rio run --add-host db=1.2.3.4 nginx

# set build parameters
rio run --build-branch dev --build-dockerfile Dockerfile.production --build-context . --build-webhook-secret webhook https://github.com/example/exmaple 

# run a service that deploy on any new tag matching '^v' and not match 'alpha'
rio run -p 8080 -n tag-demo --build-webhook-secret=githubtoken --build-tag=true --build-tag-include="^v" --build-tag-exclude="alpha" https://github.com/rancher/rio-demo

```

---

## rm

Delete resources

##### Usage
```
rio rm [TYPE/]RESOURCE_NAME
```

##### Examples

```shell script
# delete service foo
rio rm foo

# delete multiple resources of different types
rio rm svc1 svc2 router/route1 externalservice/foo
```
---

## scale

Scale a service to a desired number, or set autoscaling params

##### Usage
```
rio scale [SERVICE=NUMBER_OR_MIN-MAX...]
```

##### Examples

```shell script
rio scale foo=5

# autoscaling
rio scale foo=1-5
```

---

## stage

Stage a new revision of a service

Note that when using `--edit` certain values (like `spec.weight`) will be overwritten, and other flags (like `--env`) won't take effect.

##### Usage
```
rio stage [OPTIONS] SERVICE NEW_REVISION
```

##### Options

| flag             | aliases  | description                                        | default |
|------------------|----------|----------------------------------------------------|---------|
| --image value    |          | Runtime image (Docker image/OCI image)             |         |
| --edit           |          | Edit the config to change the spec in new revision |         |
| --env value      | -e value | Set environment variables                          |         |
| --env-file value |          | Read in a file of environment variables            |         |

##### Examples

```shell script
# stage an image (tag v3) to the 2nd version of the demo service
rio stage --image ibuildthecloud/demo:v3 demo v2

# stage the same image with different env variables
rio stage -e abc=xyz demo v2

# stage but edit first
rio stage --edit demo v2
```

---

## system logs

Print the logs from the Rio management plane

##### Usage
```
rio system logs
```

---

## system feature

View/Edit system feature/configuration

##### Uasge
```bash
# view system feature
rio system feature

# edit system feature/configuration
rio system feature --edit
```

---


## uninstall

Uninstall rio

##### Usage
```
rio uninstall [OPTIONS]
```

##### Options

| flag              | aliases | description                           | default      |
|-------------------|---------|---------------------------------------|--------------|
| --namespace value |         | namespace to install system resources | "rio-system" |

##### Examples

```shell script
rio uninstall

rio uninstall --namespace alt-namespace
```

---


## up

Apply a Riofile

##### Usage
```
rio up [OPTIONS]
```

##### Options

| flag                         | aliases  | description                                                                                                  | default |
|------------------------------|----------|--------------------------------------------------------------------------------------------------------------|---------|
| --name value                 | -n       | Set stack name, defaults to current directory name                                                           |         |
| --answers value              |          | Set answer file                                                                                              |         |
| --file value                 | -f value | Set rio file                                                                                                 |         |
| --parallel                   | -p       | Run builds in parallel                                                                                       |         |
| --branch value               |          | Set branch when pointing stack to git repo                                                                   | master  |
| --revision value             |          | Use a specific commit hash                                                                                   |         |
| --build-webhook-secret value |          | Set GitHub webhook secret name                                                                               |         |
| --build-tag                  |          | Enable builds on any new tags instead of new commits on a branch, requires webhook, does not support polling |         |
| --build-tag-include          |          | Pattern that tags must match                                                                                 |         |
| --build-tag-exclude          |          | Pattern that excludes tags                                                                                   |         |
| --build-clone-secret value   |          | Set name of secret to use with git clone                                                                     |         |
| --push-registry-secret value |          | Set secret for pushing to custom registry                                                                    |         |
| --permission value           |          | Permissions to grant to container's service account in current namespace                                     |         |

##### Examples

```shell script
# apply a file named 'Riofile' in current directory
rio up

# apply stack.yaml as a stack named mystack as 2nd revision
rio up --name mystack -f stack.yaml -p

# apply a riofile from git repo, from a specific branch and commit, using a secret, and setup webhook.
rio up --branch branchname --build-webhook-secret=githubtoken --build-clone-secret=mysecret --revision {commit_sha}  https://github.com/exmaple/example

# Set custom permissions to give the stack, and supply answers to riofile questions
rio up  --permissions '* configmaps' --answers answerfile.yaml
```

---

## weight

Set the percentage of traffic to allocate to a given service version. See also promote. 

Defaults to an immediate rollout, set duration to perform a gradual rollout

Note that once a service version is set to 100% of weight, you must assign weight to other services in order to route traffic to them.
For instance if you have svc-a and svc-b, and you set svc-a=100% and then svc-a=50%, svc-b will still have 0% weight and svc-a will still have 100%. You must set svc-b=50% to give it weight.

##### Usage
```
rio weight [OPTIONS] SERVICE_NAME=PERCENTAGE
```

##### Options

| flag       | aliases | description                                                                   | default |
|------------|---------|-------------------------------------------------------------------------------|---------|
| --duration |         | How long the rollout should take. An approximation, actual time may fluctuate | 0s      |
| --pause    |         | Whether to pause all rollouts on current app                                  | false   |

##### Examples

```shell script
# immediately shift 100% of traffic to app n@v0
rio weight n=100 

# shift n@v2 to 50% of traffic gradually over 5m
rio weight --duration=5m n@v2=50 

# Pause last command at current state, will pause all rollouts on versions in app
rio weight --pause=true n@v2=50
```
