# Running workloads

### Deploying Container Into Rio

```bash
# Exposing the service requires passing the `-p` flag to expose ports from the container
$ rio run -p 80 --name demo nginx

# You will get an endpoint URL for your service
$ rio ps

# Access endpoint URL
curl https://demo-v0-default.xxxxx.on-rio.io
```

By default Rio will create a DNS record pointing to your cluster's gateway. Rio also uses [Let's Encrypt](https://letsencrypt.org/) to create
a certificate for the cluster domain so that all services support HTTPS by default.
For example, when you deploy your workload, you can access your workload in HTTPS. The domain always follows the format
of ${app}-${namespace}.\${cluster-domain}. You can see your cluster domain by running `rio info`.

Note: If linkerd feature is enabled, Rio will automatically inject linkerd-proxy into your workload. If you would like to disable that, run `rio run --no-mesh`.

### Expose your service

Note: In order to expose your service you have pass the flag `--port`(shorthand `-p`). The format is `[service_port:]container_port[/protocol]`
If you don't pass port the service will be private by default (only accessible inside cluster).

```bash
# To expose services through 80/http
$ rio run -p 80/http nginx

# To expose services through 80/http on port name web
$ rio run -p 80/http,web nginx

# Do not expose service publicly
$ rio run -p 80,web,expose=false nginx

# To expose services through hostport 80
$ rio run -p 8080:80,web,hostport=true nginx
```

### Examples

Notes: 
- none of these examples have the port specified, so there will not be an available app endpoint
- some of these examples need proper RBAC setup, for more information check [here](./rbac.md).  

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
$ rio run --env FOO=secret://certs/tls.crt nginx

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

### Split Traffic Between Revisions
Rio natively supports splitting traffic between revisions. Splitting Traffic can be quite useful in canary deployment, Blue/Green deployment and A/B testing. 

Each Rio service you deploy will have two unique label identifiers across current namespace: `app` and `version`.
Based on `app` and `version` user is allowed to assign weight between each revision to manage traffic.

To deploy a demo application with version v1

```bash
# Names follow the format of [namespace:]app[@version]. Defaults to default namespace and v0 version.
$ rio run --name demo@v1 -p 80 ibuildthecloud/demo:v1
```

To deploy another version with version v3

```bash
# Create a new service associated to the demo app, with different desired version and image, and give it a weight of zero
$ rio stage --image ibuildthecloud/demo:v3 demo@v1 v3 

# Manually stage using run
$ rio run --name demo@v3 --ports 80 --weight 0 ibuildthecloud/demo:v3 
```

Now you have defined two services with app `demo` and versions `v1` and `v3`. To access the global endpoint that serves
traffic from both versions:

```bash
# Endpoint URL always follows the format of `${app}-${namespace}.xxxxxx.on-rio.io`
$ rio endpoints
NAME              ENDPOINTS
demo              https://demo-default.xxxxxx.on-rio.io
```
Note: This endpoint will only return versions that have weight greater than 0%. Versions with a higher weight percentage will be returned more often.

#### Assign weight between each revision

Now assign weight 50% to demo@v3

```bash
# Weight is immediately assigned by default
$ rio weight demo@v3=50%

# Gradually increase weight over the specified duration (s=seconds, m=minutes, h=hours)
$ rio weight --duration 10m demo@v3=50%

# Promote v3 service (assigns weight=100% and sets all other versions to weight = 0%)
$ rio promote demo@v3
```

Note: services are discoverable inside cluster by their short DNS name. For example services demo@v1 and demo@v3 are discoverable through
`demo-v1` and `demo-v3`. `demo` is also discoverable to serve traffic from both versions.

### Running Stateful Application (experimental)

Rio supports running stateful applications by leveraging kubernetes [persistentvolume](https://kubernetes.io/docs/concepts/storage/persistent-volumes/).


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
$ rio run -v foo:/etc,hosttype=directoryorcreate nginx
``` 

Note: hostpath type can be found in [here](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath)

