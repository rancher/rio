## PublicDomain

Rio allows you to add a vanity domain to your workloads. For example, to add a domain `www.myproductionsite.com` to your workload.

For example:

1. Setup CNAME record for your domain to rio cluster domain. You should be able to see cluster domain when running `rio info`.

    myproduction.com -----> CNAME -------> xxxxxx.on-rio.io

2. Register domain
```bash
$ rio domain register myproductionsite.com $target

# Use your own certs by providing a secret that contain tls cert and key instead of provisioning by letsencrypts. The secret has to be created first in system namespace.
$ rio domain register --secret $secret_name www.myproductionsite.com $target

# Access your domain 
$ rio endpoint
```

$target can get a service(`app@version`), a group of services(`app`) or a router(`router`)

For example, to register domain with app `demo` and version `v1`:

```bash
$ rio domain register myproductionsite.com demo@v1
```

To register domain with app `demo`:

```bash
$ rio domain register  myproductionsite.com demo
```

To register domain with router `prod`:

```bash
$ rio domain register  myproductionsite.com prod
```

To unregister domain:

```bash
$ rio unregister myproductionsite.com 
```

## ExternalService

ExternalService is a service like databases and legacy apps outside of your cluster. 
ExternalService can be IP addresses, FQDN or a Rio service in another namespace. Once added, external service can be discovered by short name within the same namespace.

To add externalservice `ext` pointing to FQDN

```bash
$ rio external create ext mydb.address
```

To add externalservice ext pointing to IP addresses

```bash
$ rio external create ext 1.1.1.1 2.2.2.2
```

To add externalservice pointing to app `demo` and version `v1`

```bash
$ rio external create ext demo@v1

# to change the namespace
$ rio external create ext foo:demo@v1
```

To add externalservice pointing to app `demo` 

```bash
$ rio external create `ext` demo
```

To add externalservice pointing to route `prod`

```bash
$ rio external create `ext` router/prod
```