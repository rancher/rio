# Quick start

To run a container with docker images

```bash
# To expose service you have to pass -p flag to expose ports from container
$ rio run -p 80 --name demo nginx

# You will get an endpoint URL for your service
$ rio ps

# Access endpoint URL
curl https://demo-v0-default.xxxxx.on-rio.io
```

Rio allows user to run a container directly from source code. Here is an example of running a container from github repository.

```bash
# by pointing to git repository, Rio will clone and build source code to docker image, and deploy it into cluster.
Also Rio will watch sequential change from git and automatically update deployment.

$ rio run -p 8080 --name demo-2 https://github.com/rancher/rio-demo
```

For more advanced use cases, check [Running workload in Rio](./workloads.md)