# FAQs 

Frequently asked questions for Rio
##

**Q:** What Kubernetes version do I need to run rio?

**A:** We recommend using [K3s](https://k3s.io/)! 
Rio should work with any Kubernetes version v1.15+. If using a managed Kubernetes instance, use the latest version available. 
We have internally tested major features of Rio with GKE v1.14.8-gke.12 and EKS v1.14.8-eks-b7174d.
##

**Q:** Why doesn't the first workload I create have a version but staged workloads do?

**A:** The first workload you create is always `v0` by default, and this is hidden in the CLI to avoid clutter.
##

**Q:** Why is the Rio CLI stuck when I run a command?

**A:** Verify you have used an available option and don't have any typos. 
If you are sure there are no typos in your command, feel free to submit an issue with the exact command and options you used.
##

**Q:** How can I delete multiple workloads at once?

**A:** Use a space to separate the services, routes, etc. Example: `rio rm svc1 svc2 router/route1 externalservice/foo`
##

**Q:** Why can't I remove services that were created with a Riofile?

**A:** If you wish to delete ALL resources associated to the stack, delete the stack itself:
1. Obtain the stack name: `rio stacks`
2. Delete it: `rio rm <stack name>`

If instead you wish to remove a specific resource, then remove it from the Riofile and reapply it: `rio up`
##

**Q:** How to configure custom domain instead of xxxxxx.on-rio.io?

**A:** Rio allows you to configure multiple cluster domains that are generated for endpoint URL. By default you get 
    
```text
${namespace}-${name}.xxxxxx.on-rio.io
```
     
if you have configured your own cluster domain you will get endpoint like:

```text
${namespace}-${name}.you.company.com
```

1. Run `rio info`. 

* 1a) If you enabled rdns feature, you can take the cluster domain and create a CNAME record from your own wildcard domain to rio cluster domain.

        *.your.company.com -----> CNAME ------> xxx.on-rio.io 

* 1b) If you disabled rdns feature, you won't be able to get `xxxxxx.on-rio.io` domain. Create A record from your domain to gateway IP.

        *.your.company.com -----> A ------> IP

     To get gateway IP run:

    ```bash
    kubectl get svc gateway-proxy -n rio-system  -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
    ```

    Also you are responsible for updating dns record if load balancer IP has changed.

2. Create a wildcard TLS secret for your own domain.

```bash
kubectl -n rio-system create secret tls your.company.com-tls --cert=/path/to/your.cert --key=/path/to/your.key
```

3. Create a clusterDomain.yaml and `kubectl apply -f ./clusterDomain.yaml`

```yaml
apiVersion: admin.rio.cattle.io/v1
kind: ClusterDomain
metadata:
  name: your.company.com
spec:
  secretName: your.company.com-tls 
  httpPort: 80
  httpsPort: 443
```

4. Rio will serve your own wildcard domain and your own certs now.
