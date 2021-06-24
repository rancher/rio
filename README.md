# Thanks for all your interest in this project. At present time, this project is on hold and the repo will be archived until we're ready to pick it up again!

# Rio [Beta]

[![Build Status](https://drone-publish.rancher.io/api/badges/rancher/rio/status.svg?branch=master)](https://drone-publish.rancher.io/rancher/rio)
[![Go Report Card](https://goreportcard.com/badge/github.com/rancher/rio)](https://goreportcard.com/report/github.com/rancher/rio)

Rio is an Application Deployment Engine for Kubernetes that can be layered on top of any standard Kubernetes cluster. Consisting of a few Kubernetes custom resources and a CLI to enhance the user experience, users can easily deploy services to Kubernetes and automatically get continuous delivery, DNS, HTTPS, routing, monitoring, autoscaling, canary deployments, git-triggered builds, and much more. All it takes to get going is an existing Kubernetes cluster and the rio CLI.

Rio is currently in Beta. 

Connect with us on the #rio channel on the [rancher slack](https://slack.rancher.io/)

## Documentation
See [here](https://rio.rancher.io) for detailed documentation and guides.

## Quick Start

1. Have a Kubernetes 1.15 or newer cluster running.

   [k3s](https://k3s.io/), [RKE](https://github.com/rancher/rke), [Minikube](https://kubernetes.io/docs/setup/minikube/), [Docker For Mac Edge](https://docs.docker.com/docker-for-mac/edge-release-notes/), [GKE](https://cloud.google.com/kubernetes-engine/), [AKS](https://docs.microsoft.com/en-us/azure/aks/), [EKS](https://aws.amazon.com/eks/), see the [install docs](/docs/install.md) for info and requirements.

2. Run

```bash
# Download the CLI (available for macOS, Windows, Linux)
$ curl -sfL https://get.rio.io | sh -   # or manually from https://github.com/rancher/rio/releases

# Setup your cluster for Rio
$ rio install

# Make sure all the pods are up and running. These may take several minutes.
$ rio -n rio-system pods

# Run a sample service
$ rio run -p 80:8080 https://github.com/rancher/rio-demo

# Check the status
$ rio ps
$ rio info
```

## License

Copyright (c) 2014 - 2020 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
