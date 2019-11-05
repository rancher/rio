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

Note: In order to expose your service you have pass `--port`(short flag `-p`). The format is `[service_port:]container_port[/protocol]`
If you don't pass port the service will be private by default.(only accessible inside cluster)


### Deploying canary deployment
Rio allows you to easily configure canary deployment by staging services and shifting traffic between revisions.

Each service you deployed will have two unique label identifier across current namespace: `app` and `version`.
Based on `app` and `version` user is allowed to assigned weight between each revision to manage traffic.

// todo: add more examples 