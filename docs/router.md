# Configuring Router

Router is a set of L7 load-balancing rules that can route between your services. 
It can add header-based routing, path-based routing, cookies, and other rules.
Routing uses the Gloo feature under the hood, so ensure that is enabled before continuing! See [Install Docs](./install.md) to learn how to enable this feature. It will be enabled by default.

#### Create Routers

To create a router:
```bash
$ rio [-n $namespace] route add $name to $target
```

Note: $name will be the router name. $target can point to individual services or a group of services.
The target service has to be in the same namespace as the router.

For example, to point to app `demo` and version `v1`:

```bash
$ rio route add prod to demo@v1
```

To point to all weighted versions of app `demo`:

```bash
$ rio route add prod to demo
```
##

#### Add Path-Based Match

Creating a route with path-based matching supports: exact match: `foo`; prefix match: `foo*`; and regular expression match: `regexp(foo.*)`.

```bash
$ rio route add $name/path to $target

# to add prefix
$ rio route add $name/path'*' to $target

# to add regular expression match
$ rio route add $name/regexp'(foo[1-9])' to $target
```

Try to access the routes you just created to test if routes are configured properly

```bash
$ curl https://route.xxx.on-rio.io/path
$ curl https://route.xxx.on-rio.io/path-test
$ curl https://route.xxx.on-rio.io/foo1 
```
##

#### Point to Different Port

Create a route to a different port:
```bash
$ rio route add $name to $target,port=8080
```
##

#### Add Header-Based Route

Creating a route with header-based matching supports: exact match: `foo`; prefix match: `foo*`; and regular expression match: `regexp(foo.*)`. 
```bash
$ rio route add --header USER=test $name to $target
$ rio route add --header USER=test'*' $name to $target
$ rio route add --header USER=regexp'(test[1-9])' $name to $target
```

Access url to test if routes are configured properly

```bash
$ curl -H'USER: test' https://route.xxx.on-rio.io
$ curl -H'USER: test-foo' https://route.xxx.on-rio.io
$ curl -H'USER: test1' https://route.xxx.on-rio.io
```
##

#### Route Based on HTTP Method

Create route based on HTTP method:
```bash
$ rio route add --method POST $name to $target
```

Sending POST requests to get traffic go through. Instead, GET request will get 404.

```bash
$ curl -i -X POST https://route.xxx.on-rio.io
$ curl -i -X GET https://route.xxx.on-rio.io
```
##

#### Manipulate headers

Add, set, or remove headers:
```bash
$ rio route add --add-header FOO=BAR $name to $target
$ rio route add --set-header FOO=BAR $name to $target
$ rio route add --remove-header FOO=BAR $name to $target
```
##

#### Rewrite to host/path

Rewrite host header and path:
```bash
$ rio route add --rewrite-host foo $name to $target
$ rio route add --rewrite-path /path $name to $target
```
##

#### Redirect

Redirect to another service:
```bash
$ rio route add $name redirect $target_service/path
```

To add a https redirect 
```bash
$ rio route add --https-redirect $name to $target
```

If you want to https redirect on your public domain, register domain to this router

```bash
$ rio domain add foo.bar $name 
```
##

#### Timeout

Add timeout:
```bash
$ rio route add --timeout-seconds $value $name to $target
```
##

#### Fault injection

Add fault injection:
```bash
$ rio route add --fault-httpcode 502 --fault-delay-milli-seconds 1000 --fault-percentage 80 $name to $target
```
##

#### Retry logic

Add retry logic:
```bash
$ rio route add --retry-attempts 5 --retry-timeout-seconds 1 $name to $target
```
##

#### Split traffic in router

Create router to different revision and different weight:
```bash
$ rio route add $name to $service@v0,weight=50 $service@v1,weight=50
```
##

#### Insert Rules

Insert a router rule instead of append so that it will be evaluated first:
```bash
$ rio route add --insert $name to $target
```
