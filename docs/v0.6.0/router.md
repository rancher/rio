# Configuring Router

Router is a set of L7 load-balancing rules that can route between your services. It can add Header-based, path-based routing, cookies
and other rules.

#### Create routers

To create routers:
```bash
$ rio [-n $namespace] route add $name to $target
```

Note: $name will be the router name. $target can point to individual services or a group of services.
Target service has to be in the same namespace with router

For example, to point to app `demo` and version `v1`:

```bash
$ rio route add prod to demo@v1
```

To point to all versions of app `demo`

```bash
$ rio route add prod to demo
```

#### Add path-based match

Create a route based path match (supports exact match: `foo`, prefix match: `foo*`, and regular expression match: `regexp(foo.*)`)

```bash
$ rio route add $name/path to $target

# to add prefix
$ rio route add $name/path* to $target

# to add regular expression match
$ rio route add $name/regexp(foo.*) to $target
```

#### Point to different port

Create a route to a different port:
```bash
$ rio route add $name to $target,port=8080
```

#### Set header and HTTP method based match

Create router based on header (supports exact match: `foo`, prefix match: `foo*`, and regular expression match: `regexp(foo.*)`) 
```bash
$ rio route add --header USER=VALUE $name to $target
```

Create route based on HTTP method
```bash
$ rio route add --method GET $name to $target
```

#### Manipulate headers

Add, set or remove headers:
```bash
$ rio route add --add-header FOO=BAR $name to $target
$ rio route add --set-header FOO=BAR $name to $target
$ rio route add --remove-header FOO=BAR $name to $target
```

#### Mirror traffic

Mirror traffic:
```bash
$ rio route add $name mirror $target
```

#### Rewrite to host/path

Rewrite host header and path
```bash
$ rio route add $name rewrite $rewrite_host/$rewrite_path
```

#### Redirect

Redirect to another service
```bash
$ rio route add $name redirect $target_service/path
```

#### Timeout

Add timeout
```bash
$ rio route add --timeout-seconds $value $name to $target
```

#### Fault injection

Add fault injection
```bash
$ rio route add --fault-httpcode 502 --fault-delay-milli-seconds 1000 --fault-percentage 80 $name to $target
```

#### Retry logic

Add retry logic
```bash
$ rio route add --retry-attempts 5 --retry-timeout-seconds 1s $name to $target
```

#### Split traffic in router

Create router to different revision and different weight
```bash
$ rio route add $name to $service@v0,weight=50 $service@v1,weight=50
```

#### Insert Rules

Insert a router rule instead of append(so it will be evaluated first)
```bash
$ rio route add --insert $name to $target
```