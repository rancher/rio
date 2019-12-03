# Install Rio

Rio can be installed from both the CLI or a Kubernetes manifest.

1. Download the latest release. 

```bash
$ curl -sfL https://get.rio.io | sh - 
```

Note: by default the script will only download the latest release from GitHub. To test a pre-release or alpha build, run.

```bash
$ curl -sfL https://get.rio.io | INSTALL_RIO_VERSION=${version} sh - 
```

2. Prepare a Kubernetes cluster, see [Clusters and Providers](#clusters-and-providers). Set the KUBECONFIG environment variable to point to your Kubernetes cluster.

3. Run 

```bash
$ rio install
```

Note: to install Rio a Kubernetes manifest, run `rio install --yaml`. It will print out the Kubernetes manifest instead of 
installing Rio directly, so that you can apply the manifest later.

## Options

See the [CLI install reference docs](cli-reference.md#install) for complete list of options.

To lower footprint of rio, use the `disable-features` flag. Just be aware that you are disabling features of Rio. Check [here](#features).

## Clusters and Providers

A 1.15 or higher version of Kubernetes is recommended.

See the Kubernetes [getting started guide](https://kubernetes.io/docs/setup/) to help choose a cluster type.

**Internal vs external IP address**

Some cloud providers will use an internal IP address by default and Rio will use that for its cluster domain.
You can override this during cluster setup by configuring external IP's, or by using a cloud-provider to setup a load balancer which will have an external IP.

For an example of installing a cluster with a cloud provider on Linode using Rancher, see [step #8 on this doc](https://www.linode.com/docs/kubernetes/how-to-deploy-kubernetes-on-linode-with-rancher-2-x/#provision-a-cluster).

**K3s**

When installing a K3s cluster, use the `--no-deploy traefik` flag.

**EKS**

Ensure you are running enough nodes with proper instance types to allow for the Rio and Kubernetes systems to run at least 45 pods.
See the docs to help determine proper sizes:

* Guide: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-eni.html#AvailableIpPerENI
* Total limits: https://github.com/awslabs/amazon-eks-ami/blob/master/files/eni-max-pods.txt

For example: you have a workload that requires 15 pods and you want to run 2 nodes.
15+45 is a 60 pod minimum requirement, so running 2 t3.large nodes (2*35=70) would be just enough.

## Features

Rio contains several independent components that are responsible for different features. By default It will install all components which give you full functionality including service mesh, source code to deployment, autoscaling and advanced API gateway.
If you have limited capacities in your cluster and don't want enable some of features, you can disable features to make footprint of Rio smaller.

Here is the list of features:

| Feature     | Description                                       |
|-------------|---------------------------------------------------|
| autoscaling | Auto-scaling services based on in-flight requests |
| build       | Rio Build, from source code to deployment         |
| gloo        | API gateway backed by gloo                        |
| linkerd     | Linkerd service mesh                              |
| letsencrypt | Let's Encrypt                                     |
| rdns        | Acquire DNS from public Rancher DNS service       |
| dashboard   | Rio UI                                            |

The core functionality of Rio includes Rio CLI UX and Riofile UX, which is built-in and doesn't rely on any of these features. Beside core functionality, here is a list of advanced features.

**Autoscaling**

Manages autoscaling services in Rio. It deploys a pod inside cluster, watches the autoscaling config for each service, and scrapes metrics from pods. It will then make decisions to scale pods up and down based on those real-time metrics and autoscaling configs.

**Build**

Build is responsible for building images from source code, which allows you to run deployment directly from source code with a Dockerfile. Build is also responsible for setting up webhook to watch your source code change and trigger automatic builds.

**Gloo**

Gloo is the default API gateway Rio deploys to manage ingress traffic outside of cluster. It uses [Envoy](https://www.envoyproxy.io/) under the hood to provide rate limiting, circuit breaking, retries and other advanced routing features. Rio router programs Gloo virtualservices.

**Linkerd**

Linkerd is a service mesh for Kubernetes and other frameworks. It makes running services easier and safer by giving you runtime debugging, observability, reliability, and security—all without requiring any changes to your code. By default Rio ships with Linkerd and enables its service mesh within Rio services

**Letsencrypt**

Letsencrypt is default CA that Rio uses to provision and sign certificates and use that to provide TLS traffic for your workloads. You easily swap certificates to your own certs. Under the hood it deploys [cert-manager](https://github.com/rancher/cert-manager/tree/rdns-provider) to manage certificate provisioning.

**Rdns**

Rdns is public DNS service hosted by Rancher. Each cluster with Rio installed will get a unique cluster domain (xxxxxx.on-rio.io). Rio registers this domain to Loadbalancer IP of API gateway service, which will eventually point to your service. Rdns will keep DNS record up-to-date if loadbalancer IP has changed.
If you disabled this feature, you are responsible for updating your own DNS record if the loadbalancer IP changes.

**Dashboard**

Dashboard is a built-in web UI for Rio that runs within the cluster. It can be used as an alternative to the Rio CLI. 

#

**Note**: If You have limited cpu and memory in your k8s cluster, we recommend you to disable features that you don't need in order to keep footprint of Rio smaller.

For example, To disable autoscaling:

```bash
$ rio install --disable-features autoscaling 
```

To disable linkerd service mesh:

```bash
$ rio install --disable-features linkerd 
```

If you would like to use your own cluster domain and certificates instead of rdns and letsencrypt, you can disable these two features:

```bash
$ rio install --disable-features rdns,letsencrypt 
```

And follow docs in [here](./faq.md) to see how to bring your own cluster domain and certs.

### Flipping features after installation

Rio manages features in a configMap called `rio-config` in `rio-system` namespace. It follows the format: 

```json
{
  "features": {
    "autoscaling": {
      "enabled": true
    },
    "build": {
      "enabled": true
    },
    "dashboard": {
      "enabled": true
    }
  },
  "letsEncrypt": {},
  "gateway": {
    "serviceName": "gateway-proxy-v2",
    "serviceNamespace": "rio-system"
  }
}
```

To edit the configMap, run 

```bash
rio -s edit configmap/rio-config 
```