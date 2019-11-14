# Running workloads

### Deploying your container into Rio

```bash
# To expose service you have to pass -p flag to expose ports from container
$ rio run -p 80 --name demo nginx

# You will get an endpoint URL for your service
$ rio ps

# Access endpoint URL
curl https://demo-v0-default.xxxxx.on-rio.io
```

By default Rio will create a DNS record pointing to your cluster's gateway. Rio also uses Let's Encrypt to create
a certificate for the cluster domain so that all services support HTTPS by default.
For example, when you deploy your workload, you can access your workload in HTTPS. The domain always follows the format
of ${app}-${namespace}.\${cluster-domain}. You can see your cluster domain by running `rio info`.

Note: If linkerd feature is enabled, Rio will automatically inject linkerd-proxy into your workload. If you would like disable that, run `rio run --no-mesh`.

### Expose your service

Note: In order to expose your service you have pass `--port`(short flag `-p`). The format is `[service_port:]container_port[/protocol]`
If you don't pass port the service will be private by default.(only accessible inside cluster)

```bash
# To expose services through 80/http
$ rio run -p 80/http nginx

# To expose services through 80/http on port name web
$ rio run -p 80/http,web nginx

# Do not expose service publicly
$ rio run -p 80,web,expose=false nginx

# To expose services through hostport 80
$ rio run -p 80,web,hostport=true nginx
```

### Examples

```bash
# Running container with configMap mounted into containers, requires configMap to exist in the same namespace
$ rio run --config config/nginx.conf:/etc/nginx/nginx.conf nginx

# Running container with configMap mounted into containers, all keys from config will be mounted
$ rio run --config config:/etc/nginx nginx

# Running container with configMap as environment variable.
$ rio run --env FOO=config://data/content nginx # Use configMap data and key content as value of environment variable FOO

# Running container with secret mounted into containers, requires secret to exist in the same namespace 
$ rio run --secret certs/tls.crt:/etc/ssl/tls.key nginx

# Running container with secret mounted into container, all keys from secret will be mounted
$ rio run --secret certs:/etc/ssl/ nginx

# Running container with secret as environment variable.
$ rio run --env FOO=secret://certs/tls.crt

# Running container with no service mesh
$ rio run --no-mesh nginx

# Running container with privileged flag
$ rio run --privileged nginx

# Running container with environment variables
$ rio run --env FOO=BAR --env FOO1=BAR1 nginx

# Running container and attach to it
$ rio run -it nginx bash

# Running container with scale of 5
$ rio run --scale=5 nginx

# Running container with host networking
$ rio run --net=host nginx
```

For more examples, check [here](./cli-reference.md)

### Split Traffic between revisions
Rio support splitting traffic natively between revisions. Splitting Traffic can be quite useful in canary deployment, Blue/Green deployment and A/B testing. 

Each Rio service you deployed will have two unique label identifiers across current namespace: `app` and `version`.
Based on `app` and `version` user is allowed to assigned weight between each revision to manage traffic.

To deploy a demo application with version v1

```bash
# Name follow the format of [namespace:]app[@version]. Default to default namespace and v0 version.
$ rio run --name demo@v1 -p 80 ibuildthecloud/demo:v1
```

To deploy another version with version v3

```bash
# Stage copy and change target service spec, create a new service with desired version and give it weight of zero.
$ rio stage --image ibuildthecloud/demo:v3 demo@v1 v3 

# Manually stage using run
$ rio run --name demo@v3 ibuildthecloud/demo:v3 
```

Now that you have defined two service with app `demo` and version `v1` and `v3`. To access the global endpoint that serves
traffic from both version:

```bash
# Endpoint URL always follows the format of `${app}-${namespace}.xxxxxx.on-rio.io`
$ rio endpoints
NAME              ENDPOINTS
demo              https://demo-default.xxxxxx.on-rio.io
```

#### Assign weight between each revision

Now assign weight 50% to demo@v3

```bash
# Weight is immediately assigned by default
$ rio weight demo@v3=50%

# Gradually increase weight in specified period
$ rio weight --duration 10m demo@v3=50%

# Promote v3 service
$ rio promote demo@v3
```

Note: services are discoverable inside cluster by their short DNS name. For example services demo@v1 and demo@v3 are discoverable through
`demo-v1` and `demo-v3`. `demo` is also discoverable to serve traffic from both versions.

### Running stateful application (experimental)

Rio support running stateful application leveraging kubernetes [persistentvolume](https://kubernetes.io/docs/concepts/storage/persistent-volumes/).


To mount a volume into container (By default it will create [emptydir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume):

```bash
$ rio run -v foo:/data nginx
```

To mount a persistent volume into container (By default it will create persistent volume if cluster has default storageclass, otherwise it will use existing pvc with the same name):

```bash
$ rio run -v foo:/data,persistent=true nginx
```

To mount a hostpath volume into container

```bash
$ rio run -v rio run -v foo:/etc,hosttype=directoryorcreate nginx
``` 

Note: hostpath type can be found in [here](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath)

