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

2. Prepare a Kubernetes cluster. Setup KUBECONFIG environment variable to point to your kubernetes cluster. To choose a kubernetes cluster, check [here](https://kubernetes.io/docs/setup/). 

Note: 1.15 or higher version of kubernetes is recommended.

3. Run 

```bash
$ rio install
```

Note: to install rio from kubernetes manifest, run `rio install --yaml`. It will print out kubernetes manifest instead of 
installing rio directly, so that you can apply the manifest later.

#### Options

| Option | Type | Description |
|------|----| -------------|
| `--ip-address` | string array | Manually specify IP addresses to generate rdns domain, supports comma separated values | 
| `--disable-features` | string array | Manually specify features to disable, supports comma separated values |
| `--enable-debug` | boolean | Enable debug logging in rio-controller pod |
| `--yaml` | boolean | Only print out k8s yaml manifest |
| `--check` | boolean | Only check status, don't deploy controller |

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

