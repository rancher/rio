# Install Rio

Rio can be installed from both CLI or kubernetes manifest.

1. Download the latest release. 

```bash
$ curl -sfL https://get.rio.io | sh - 
```

Note: by default it will only download the latest release from github. To test a pre-release or alpha build, run.

```bash
$ curl -sfL https://get.rio.io | INSTALL_RIO_VERSION=${version} sh - 
```

2. Prepare a Kubernetes cluster, see [Clusters and Providers](#clusters-and-providers). Setup KUBECONFIG environment variable to point to your kubernetes cluster.

3. Run 

```bash
$ rio install
```

Note: to install rio from kubernetes manifest, run `rio install --yaml`. It will print out kubernetes manifest instead of 
installing rio directly, so that you can apply the manifest later.

#### Options

See the [CLI install reference docs](cli-reference.md#install) for complete list of options.

**--ip-address**

Manually specify IPAddress for API gateway services. The IP will be used to generate a record for the cluster domain. 
By default, if this flag is not specified, rio will use the IP of [Service Loadbalancer](https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/) that points to API gateway. 

Note: If service loadbalancer cannot be provisioned, [Nodeport](https://kubernetes.io/docs/concepts/services-networking/service/#nodeport) is used to expose API gateway.

**--disable-features**

Choose features to be disabled when starting rio control plane. Below are a list of available features

| Feature | Description |
|----------|----------------|
| autoscaling | Auto-scaling services based on in-flight requests
| build | Rio Build, from source code to deployment
| gloo | API gateway backed by gloo
| linkerd | Linkerd service mesh
| letsencrypt | Let's Encrypt
| rdns | Acquire DNS from public Rancher DNS service
| dashboard | Rio UI

Examples:

```bash
# disable autoscaling feature
$ rio install --disable-features autoscaling

# disable autoscaling, linkerd 
$ rio install --disable-features autoscaling --disable-features linkerd
```

**--enable-debug**

Enable debug logging in rio control plane

**--yaml**

Print out kubernetes manifests that are needed to install Rio

**--check**

Check if rio is installed in the current cluster without deploying rio controller. 
If rio has not been installed, this command might hang on `Waiting for rio controller to initialize`.


#### Clusters and Providers

A 1.15 or higher version of kubernetes is recommended.

See the kubernetes [getting started guide](https://kubernetes.io/docs/setup/) to help choose a cluster type.

**k3s**

When installing on k3s use the `--no-deploy traefik` flag.

**EKS**

Ensure you are running enough nodes with proper instance types to allow for the rio and kubernetes systems to run at least 45 pods.
See the docs to help determine proper sizes:

* Guide: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-eni.html#AvailableIpPerENI
* Total limits: https://github.com/awslabs/amazon-eks-ami/blob/master/files/eni-max-pods.txt

For example: you have a workload that requires 15 pods and you want to run 2 nodes.
15+45 is a 60 pod minimum requirement, so running 2 t3.large nodes (2*35=70) would be just enough.
