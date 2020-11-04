# Developing application using Rio

The Rio CLI provides an easy way to build and deploy your applications into a Kubernetes cluster. It takes advantage of [Riofile](./riofile.md) and [buildkit](https://github.com/moby/buildkit) to
build images and update deployments with them.

#### Using Riofile to develop applications

Let's use this [repo](https://github.com/ibuildthecloud/rancher-demo) as an example:

Clone this repo.

```bash
$ git clone https://github.com/ibuildthecloud/rancher-demo
``` 

Go into the repo

```bash
$ cd rancher-demo
```

Create a Riofile in the root directory

```yaml
services:
  demo:
    image: ./          # By giving image a relative path, it tells rio to use this as build context
    port: 8080/http    # defining ports to expose
```

Run

```bash
$ rio up
```

Check with `rio ps`. It should create a service with URL serving the content. It should be serving blue cows.

Open the `Dockerfile` and change `ENV COW_COLOR` from `blue` to `red`.

Re-run `rio up`. Once it is finished, it should already be updated with new images and start serving red cows.

By following the example above, you can now develop your code locally and run `rio up` to see your code changes automatically.

**Note**: This feature requires a Dockerfile. 

#### Manually build and run

You can also use the Rio CLI to build and run an image locally in your cluster.

Go to the root directory of repo and run

```bash
$ rio build
```

Wait for the image to be built. Once it is done, run

```bash
$ rio images
```

Run containers with images you just built

```bash
$ rio run -p 8080 localhost:5442/default/rancher-demo:latest
```
