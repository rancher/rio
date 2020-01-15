## Concepts

Rio introduces a small number of new concepts: Service, Router, External Service, and Domain. In addition, it reuses two existing Kubernetes resources: ConfigMaps and Secrets.

The power of Rio resides in its ability to utilize the power of Kubernetes, Linkerd service mesh, and Tekton CI/CD through a simple set of concepts.

### Service

Service is the core concept in Rio. Services are a scalable set of identical containers.
Running containers in Rio (`rio run`) creates a Service. That can be manually scaled with `rio scale`. 
Services are assigned a DNS name so that they can be discovered and accessed by other services.

### Stack

Stack is an abstraction of `Riofile`. Riofile is a file that is similar to docker-compose. 
In it, multiple Rio resources (services, routers, etc.) can be defined in a single file with simplified syntax that is easy to memorize.

### Router

Router is a resource that manages load balancing and traffic routing rules. Routing rules can route based
on hostname, path, HTTP headers, protocol, and source.

### External Service

External Service provides a way to register external IPs or hostnames in the service mesh so they can be accessed by Rio services.
Also, Rio services running in one namespace can be declared as ExternalServices in another namespace.

### Public Domain

Public Domain can be configured to assign a service or router a vanity domain like www.myproductionsite.com.

### Cluster Domain

Cluster Domain is a DNS name that can be configured globally on each service so that each service will have an endpoint like $service_name-$namespace.xxx.clusterdomain.
By default Rio will create and assign the `on-rio.io` DNS name to each service, so it will have a DNS name like `$name-$namespace.xxx.on-rio.io`.

### Configs

ConfigMaps are a standard Kubernetes resource and can be referenced by Rio services. 
They are pieces of configuration that can be mounted into pods so that configuration data can be separated from image artifacts.

### Secrets

Secrets are a standard Kubernetes resource and can be referenced by Rio services. They contain sensitive data that can be mounted into pods. 
