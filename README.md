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

On start up, you can specify which feature to disable when installing rio. Simply run `rio install --disable-features autoscaling,build --disable-features letsencrypt`

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

## Documentation
Detailed documentation can be found in [here](/docs/README.md).

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
