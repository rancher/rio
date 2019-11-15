# CLI Reference

############################
TEMPLATE - REMOVE THIS

##### Usage
```
```

##### Options

| flag       | aliases | description                                  | default |
|------------|---------|----------------------------------------------|---------|
| 

##### Examples

```shell script
```

############################











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
# see previous builds from stacks or workloads
rio build history

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
| --format    |         | 'json' or 'yaml' or Custom format: '{{.Name}} {{.Obj.Name}}' [$FORMAT] |         |
| --all       | -a      | print all resources, including router and externalservice              |         |
| --workloads | -w      | include apps/v1 Deployments and DaemonSets in output                   |         |


##### Examples

```shell script
# show services and workloads
rio ps -w

# output json
rio ps --format json

# display name and weight in custom format
rio ps --format "{{.Obj.Name}} -> {{.Data.Weight}}" 
```

## image

List images built from local registry

##### Usage
```
rio image
```


## run
todo

## rm

Delete resources

##### Usage
```
rio rm [TYPE/]RESOURCE_NAME
```


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


## attach
todo

## logs

Print logs from services or containers

##### Usage
```
rio logs [OPTIONS] SERVICE/BUILD
```

##### Options

```| flag              | aliases  | description                                                                                       | default |
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
```

##### Examples

```shell script
# get logs from a service
rio logs demo

# Get logs from a build
rio build history
rio logs taskrun/affectionate-mirzakhani-mfp5q-ee709-4e40c

# get 1 previous log line for the linkerd-proxy in demo service
rio logs --tail 1 --container linkerd-proxy -a demo

# ignore init-containers and filter to waiting or terminated pods, include timestamps
rio logs --container-state "terminated,waiting" --init-containers=false --timestamps demo

# target terminated pods of all kinds, format as json
rio logs -p -a  --output json demo
```


## install
todo

## uninstall
todo

## Stage

Stage a new revision of a service


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
rio stage --env abc=xyz demo v2
```


## weight

Set the percentage of traffic to allocate to a given service version. See also promote. 

Defaults to an immediate rollout, set duration to perform a gradual rollout

##### Usage
```
rio weight [OPTIONS] SERVICE_NAME=PERCENTAGE
```

##### Options

| flag       | aliases | description                                  | default |
|------------|---------|----------------------------------------------|---------|
| --duration | none    | How long the rollout should take             | 0s      |
| --pause    | none    | Whether to pause all rollouts on current app | false   |

##### Examples

```shell script
# immediately shift 100% of traffic to app n@v0
rio weight n=100 

# shift n@v2 to 50% of traffic gradually over 5m
rio weight --duration=5m n@v2=50 

# Pause last command at current state, will pause all rollouts on versions in app
rio weight --pause=true n@v2=50
```


## Promote

Send 100% of traffic to an app version and scale down other versions. See also weight. 

##### Usage

```
rio promote [OPTIONS] SERVICE_NAME
```

##### Options

| flag       | aliases | description                                  | default |
|------------|---------|----------------------------------------------|---------|
| --duration | none    | How long the rollout should take             | 0s      |
| --pause    | none    | Whether to pause all rollouts on current app | false   |

##### Examples

```shell script
# promote n@v2 
rio promote n@v2

# promote n@v2 over 1 hour 
rio promote --duration=1h n@v2

# pause last command
rio promote --pause=true n@v2
```


## systemlogs

Print the logs from Rio management plane

##### Usage
```
rio systemlogs
```

## up

Apply a Riofile

##### Usage
```
rio up [OPTIONS]
```

##### Options

| flag                         | aliases                                                                  | description | default |
|------------------------------|--------------------------------------------------------------------------|-------------|---------|
| --name value                 | Set stack name, defaults to current directory name                       |             |         |
| --answers value              | Set answer file                                                          |             |         |
| --file value                 | -f value        Set rio file                                             |             |         |
| --parallel                   | -p                Run builds in parallel                                 |             |         |
| --branch value               | Set branch when pointing stack to git repo (default: "master")           |             |         |
| --revision value             | Use a specific commit hash                                               |             |         |
| --build-webhook-secret value | Set GitHub webhook secret name                                           |             |         |
| --build-clone-secret value   | Set name of secret to use with git clone                                 |             |         |
| --permission value           | Permissions to grant to container's service account in current namespace |             |         |


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
rio kill pod/test-v042dxp-5fb7d8f677-f9xgn
```


## info

Show system info

##### Usage
```
rio info
```
