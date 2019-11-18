## PublicDomain

Rio allows you to add a vanity domain, such as `www.myproductionsite.com`, to your workloads.

For example:

1. Setup a CNAME record from your domain to the Rio cluster domain. You should be able to see the cluster domain when running `rio info`.

    myproduction.com -----> CNAME -------> xxxxxx.on-rio.io

2. Register the domain in rio
```bash
# Register a target service with the domain
$ rio domain register myproductionsite.com $target

# Use your own certs by providing a secret that contains a tls cert and key instead of the default LetsEncrypt certs. The secret has to be created first in system namespace.
$ rio domain register --secret $secret_name www.myproductionsite.com $target

# Access your domain 
$ rio endpoint
```

In the above example, `$target` can be a service(`app@version`), a group of services(`app`) or a router(`router`)

For example, to register a domain with app `demo` and version `v1`:

```bash
$ rio domain register myproductionsite.com demo@v1
```

To register a domain with app `demo`:

```bash
$ rio domain register  myproductionsite.com demo
```

To register a domain with router `prod`:

```bash
$ rio domain register  myproductionsite.com router/prod
```

To unregister a domain:

```bash
$ rio unregister publicdomain/myproductionsite.com 
```

## ExternalService

ExternalService is a service such as a database or legacy app outside of your cluster. 
ExternalService can be specified with an IP addresses, FQDN, or a Rio service in another namespace. 
Once added, external services can be discovered by short name within the same namespace.

To add externalservice `ext` pointing to FQDN

```bash
$ rio external create ext mydb.address
```

To add externalservice `ext` pointing to IP addresses

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
$ rio external create ext demo
```

To add externalservice pointing to route `prod`

```bash
$ rio external create ext router/prod
```
