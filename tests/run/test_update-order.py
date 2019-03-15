import util


def run_update_order(stack, *value):

    name = util.rioRun(stack, ' '.join(value), 'nginx')

    return name


def rio_chk_update_order(stack, sname):
    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    return inspect["updateOrder"]


def kube_chk_maxSurge(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['strategy']['rollingUpdate']
    results = cnt['maxSurge']

    return results


def kube_chk_maxUnavailable(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['strategy']['rollingUpdate']
    results = cnt['maxUnavailable']

    return results


def test_updateOrder1(stack):
    service_name = run_update_order(stack, '--update-order', 'start-first')

    rio_update_order = rio_chk_update_order(stack, service_name)
    assert rio_update_order == "start-first"

    kube_maxSurge = kube_chk_maxSurge(stack, service_name)
    assert kube_maxSurge == 1

    kube_maxUnavailable = kube_chk_maxUnavailable(stack, service_name)
    assert kube_maxUnavailable == 0


def test_updateOrder2(stack):
    service_name = run_update_order(stack, '--update-order', 'stop-first')

    rio_update_order = rio_chk_update_order(stack, service_name)
    assert rio_update_order == "stop-first"

    kube_maxSurge = kube_chk_maxSurge(stack, service_name)
    assert kube_maxSurge == 0

    kube_maxUnavailable = kube_chk_maxUnavailable(stack, service_name)
    assert kube_maxUnavailable == 1
