# Run Validation test.  Use functions to test run and get outpu

import util


def riotest(stack, service):
    fullName = (f"{stack}/{service}")

    inspect = util.rioInspect(fullName)

    return inspect['state']


def kubetest(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    replicas = obj['status']['replicas']

    return replicas


def test_rio_status(stack, service):
    assert riotest(stack, service) == "active"


def test_kube_replicas(stack, service):
    assert kubetest(stack, service) == 1
