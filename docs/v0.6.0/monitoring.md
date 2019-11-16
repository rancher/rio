# Monitoring 

Rio deploys [linkerd](https://linkerd.io/) to collect metrics from workloads. You can use Rio dashboard or linkerd dashboard to monitor traffic for your applications.

To run Rio dashboard, check [here](./dashboard.md).

To run linkerd dashboard:

```bash
$ rio linkerd

# if you have linkerd binary
$ linkerd dashboard
```