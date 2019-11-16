# RBAC

Rio can be configured and used in a multi-tenant environment while providing the best practices for security.

This section requires basic knowledge of kubernetes RBAC. To understand kubernetes RBAC, go [here](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

By default Rio creates four roles `rio-admin`, `rio-privileged`, `rio-readonly`, `rio-standard`. Detailed permissions can be found in [here](https://github.com/rancher/rio/blob/master/stacks/rio-bootstrap-stack.yaml).

---

To create a binding from one of the roles to your user and group in default namespace:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
# This role binding allows "jane" to have rio admin access in the "default" namespace.
kind: RoleBinding
metadata:
  name: rio-admin
  namespace: default
subjects:
- kind: User
  name: jane # Name is case sensitive
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole 
  name: rio-admin
  apiGroup: rbac.authorization.k8s.io
```

**Note:** Pre-defined roles are experimental and subject to change.

### Container security

Rio by default restricts users without admin permissions from deploying containers with insecure parameters. 
This is done by implementing a [validatingMutationWebhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/).

The following fields are protected by default:

- Enable and disable sidecar injection
- Privileged containers
- Hostpath 
- Hostport
- HostNetworking

To enable a user to have these permissions, the following verbs must be granted on a role that is bound with the user:

- rio-servicemesh
- rio-privileged
- rio-hostpath
- rio-hostport
- rio-hostnetwork

For example, to create a role with permission to launch privileged containers:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: default
  name: run-privileged
rules:
- apiGroups: ["rio.cattle.io"]
  resources: ["services"]
  verbs: ["create", "update", "delete", "rio-privileged"]
```

Then follow the previous example to bind the user to this role.
