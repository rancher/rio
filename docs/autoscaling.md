# Autoscaling 

Rio deploys a simple autoscaler to watch metrics from workloads and scale application based on current in-flight requests.

Note: Metric is scraped from linkerd-proxy sidecar, this requires your application to be injected with linkerd sidecar.
This will happen by default when running new workloads.

To enable autoscaling:

```bash
$ rio run --scale 1-10 -p 8080 strongmonkey1992/autoscale:v0

# to give a higher concurrency
$ rio run --scale 1-10 --concurrency 20 -p 8080 strongmonkey1992/autoscale:v0 

# to scale to zero
$ rio run --scale 0-10 -p 8080 strongmonkey1992/autoscale:v0
```

To put load the following example use [hey](https://github.com/rakyll/hey):

```bash
hey -z 3m -c 60 http://xxx-xx.xxxxxx.on-rio-io

# watch the scale of your service
$ watch rio ps
```

Note: `concurrency` means the maximum in-flight requests each pod can take. If your total in-flight request is 60 and concurrency 
is 10, Rio will scale workloads to 6 replicas.

Note: When scaling application to zero, the first request will take longer.