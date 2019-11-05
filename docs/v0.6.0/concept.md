## Concepts

Rio introduces a small number of new concepts: Service, Router, External Service, and Domain. In addition, it reuses two existing Kubernetes resources: ConfigMaps and Secrets.

The power of Rio resides in its ability to utilize the power of Kubernetes, Linkerd service mesh, and Tekton CI/CD through a simple set of concepts.

### Service

Service is the core concept in Rio. Services are a scalable set of identical containers.
When you run containers in Rio you create a Service. `rio run` will
create a service. You can scale that service with `rio scale`. Services are assigned a DNS name so that it can be discovered and accessed from other services.

### Router

Router is a resource that manages load balancing and traffic routing rules. Routing rules can route based
on hostname, path, HTTP headers, protocol, and source.

### External Service

External Service provides a way to register external IPs or hostnames in the service mesh so they can be accessed by Rio services.

### Public Domain

Public Domain can be configured to assign a service or router a vanity domain like www.myproductionsite.com.

### Cluster Domain

Cluster Domain is a dns name that can be configured globally on each service so that each service will have endpoint like $service_name-$namespace.xxx.clusterdomain.
By default Rio will create and assign `on-rio.io` dns name on each service, so you will get a dns name like `$name-$namespace.xxx.on-rio.io`.

### Configs

ConfigMaps are a standard Kubernetes resource and can be referenced by Rio services. It is a piece of configuration which can be mounted into pods so that configuration data can be separated from image artifacts.

### Secrets

Secrets are a standard Kubernetes resource and can be referenced by rio services. It contains sensitive data which can be mounted into pods. 