## PublicDomain

Rio allows you to add a vanity domain, such as `www.myproductionsite.com`, to your workloads.

For example:

* If you enabled rdns feature, Setup a CNAME record from your domain to the Rio cluster domain. You should be able to see the cluster domain when running `rio info`.

    www.myproduction.com -----> CNAME -------> xxxxxx.on-rio.io

* If you disabled rdns feature, you won't be able to get `xxxxxx.on-rio.io` domain. Instead, create A record to IP of API gateway
    
    Get the ip of the service loadbalancer that points to API gateway

```bash
 kubectl get svc gateway-proxy -n rio-system  -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

Create DNS record

www.myproduction.com -----> A -------> IP

Register the domain in rio

```bash
# Register a target service with the domain
$ rio domain register myproductionsite.com $target

# Access your domain 
$ rio endpoint
```

You can provide your own certificates instead of using letsencrypt. To create a secret that contains your tls.crt and tls.key:

```bash
kubectl -n rio-system create secret tls www.myproductionsite.com-tls --cert=/path/to/your.cert --key=/path/to/your.key
```

Register domain with secret you just created 

```bash
$ rio domain register --secret www.myproductionsite.com-tls www.myproductionsite.com $target
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
$ rio domain unregister publicdomain/myproductionsite.com 
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
