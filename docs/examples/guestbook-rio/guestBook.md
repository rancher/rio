
### How to use Rio to deploy an application with arbitary YAML
_for Rio v0.6.0 and greater_

In this example we will see how to define both a normal Rio service and arbitary Kubernetse manifests and deploy both of these with Rio. Follow the quickstart to get Rio installed into your cluster and ensure the output of `rio info` looks similar to this: 
```
Rio Version:  >=0.6.0
Rio CLI Version: >=0.6.0
Cluster Domain: enu90s.on-rio.io
Cluster Domain IPs: <cluster domain ip>
System Namespace: rio-system
System Ready State: true
Wildcard certificates: true

System Components:
gateway-v2 status: Ready
rio-controller status: Ready
```

First, lets use a Riofile to define a a basic Rio service. 

```
configs:
  conf:
    index.html: |-
      <!DOCTYPE html>
      <html>
      <body>
      
      <h1>Hello World</h1>
      
      </body>
      </html>
services:
  nginx:
    image: nginx
    ports:
    - 80/http
    configs:
    - conf/index.html:/usr/share/nginx/html/index.html
```

Next, we can augment this service with the Kubernetes [sample guestbook](https://kubernetes.io/docs/tutorials/stateless-application/guestbook/)


```
configs:
    conf:
        index.html: |-
            <!DOCTYPE html>
            <html>
            <body>

            <h1>Hello World</h1>

            </body>
            </html>
services:
    nginx:
        image: nginx
        ports:
            - 80/http
        configs:
            - conf/index.html:/usr/share/nginx/html/index.html

kubernetes:
    manifest: |-
      apiVersion: v1
      kind: Service
      metadata:
        name: redis-master
        labels:
          app: redis
          tier: backend
          role: master
      spec:
        ports:
        - port: 6379
          targetPort: 6379
        selector:
          app: redis
          tier: backend
          role: master
---
      apiVersion: apps/v1 #  for k8s versions before 1.9.0 use apps/v1beta2  and before 1.8.0 use extensions/v1beta1
      kind: Deployment
      metadata:
        name: redis-master
      spec:
        selector:
          matchLabels:
            app: redis
            role: master
            tier: backend
        replicas: 1
        template:
          metadata:
            labels:
              app: redis
              role: master
              tier: backend
          spec:
            containers:
            - name: master
              image: k8s.gcr.io/redis:e2e  # or just image: redis
              resources:
                requests:
                  cpu: 100m
                  memory: 100Mi
              ports:
              - containerPort: 6379
---
      apiVersion: v1
      kind: Service
      metadata:
        name: redis-slave
        labels:
          app: redis
          tier: backend
          role: slave
      spec:
        ports:
        - port: 6379
        selector:
          app: redis
          tier: backend
          role: slave
---
      apiVersion: apps/v1 #  for k8s versions before 1.9.0 use apps/v1beta2  and before 1.8.0 use extensions/v1beta1
      kind: Deployment
      metadata:
        name: redis-slave
      spec:
        selector:
          matchLabels:
            app: redis
            role: slave
            tier: backend
        replicas: 2
        template:
          metadata:
            labels:
              app: redis
              role: slave
              tier: backend
          spec:
            containers:
            - name: slave
              image: gcr.io/google_samples/gb-redisslave:v1
              resources:
                requests:
                  cpu: 100m
                  memory: 100Mi
              env:
              - name: GET_HOSTS_FROM
                value: dns
                # If your cluster config does not include a dns service, then to
                # instead access an environment variable to find the master
                # service's host, comment out the 'value: dns' line above, and
                # uncomment the line below:
                # value: env
              ports:
              - containerPort: 6379
---
      apiVersion: v1
      kind: Service
      metadata:
        name: frontend
        labels:
          app: guestbook
          tier: frontend
      spec:
        # if your cluster supports it, uncomment the following to automatically create
        # an external load-balanced IP for the frontend service.
        # type: LoadBalancer
        ports:
        - port: 80
        selector:
          app: guestbook
          tier: frontend
---
      apiVersion: apps/v1 #  for k8s versions before 1.9.0 use apps/v1beta2  and before 1.8.0 use extensions/v1beta1
      kind: Deployment
      metadata:
        name: frontend
      spec:
        selector:
          matchLabels:
            app: guestbook
            tier: frontend
        replicas: 3
        template:
          metadata:
            labels:
              app: guestbook
              tier: frontend
          spec:
            containers:
            - name: php-redis
              image: gcr.io/google-samples/gb-frontend:v4
              resources:
                requests:
                  cpu: 100m
                  memory: 100Mi
              env:
              - name: GET_HOSTS_FROM
                value: dns
                # If your cluster config does not include a dns service, then to
                # instead access environment variables to find service host
                # info, comment out the 'value: dns' line above, and uncomment the
                # line below:
                # value: env
              ports:
              - containerPort: 80
```

Typically you would track your Riofile with some form of VCS but for now simply save it in a local directory.

Next, run `rio up` in that directory.

You can watch Rio service come up with `rio ps` and the kubernetes deployments with `kubectl get deployments -w`.

You can check the sample service came up by going to the endpoint given by `rio ps`
```
NAME      IMAGE     ENDPOINT                                          SCALE     APP       VERSION    WEIGHT    CREATED       DETAIL
nginx     nginx     https://nginx-2c21baa1-default.enu90s.on-rio.io   1         nginx     2c21baa1   100%      4 hours ago
```

We can use rio to expose the service and provision a LetsEncrypt certificate for it. 

` rio router add guestbook to frontend,port=80 `

This will create a route to the service and create an endpoint. 

```
rio endpoints
NAME        ENDPOINTS
nginx       https://nginx-default.enu90s.on-rio.io
guestbook   https://guestbook-default.enu90s.on-rio.io
```

We can now access this endpoint over encrypted https! 

