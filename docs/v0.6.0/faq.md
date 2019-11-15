# FAQs 

Frequently asked questions about working with rio
##

**Q:** What Kubernetes version do I need to run rio?

**A:** We recommend using [k3s](https://k3s.io/)! 
Rio should work with any kubernetes version v1.15+. If using a managed kubernetes instance, use the latest version available. 
We have internally tested major features of rio with GKE v1.14.8-gke.12 and EKS v1.14.8-eks-b7174d.
##

**Q:** Why doesn't the first workload I create have a version but staged workloads do?

**A:** The first workload you create is always `v0` by default, and this is hidden in the CLI to avoid clutter.
##

**Q:** Why is the Rio cli stuck when I run a command?

**A:** Verify you have used an available option and don't have any typos. 
If you are sure there are no typos in your command, feel free to submit an issue with the exact command and options you used.
##

**Q:** How can I delete multiple workloads at once?

**A:** Use a space to separate the services, routes, etc. Example: `rio rm svc1 svc2 router/route1 externalservice/foo`
##

**Q:** Why can't I remove services that were created with a Riofile?

**A:** If you wish to delete ALL resources associated to the stack, delete the stack itself:
1. Obtain stack name: `rio stacks`
2. Delete it: `rio rm <stack name>`

If instead you wish to remove a specific resource, then remove it from the Riofile and reapply it: `rio up`
##
