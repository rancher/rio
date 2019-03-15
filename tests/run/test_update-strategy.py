import util


def run_updateStrategy(stack, *value):

    name = util.rioRun(stack, ' '.join(value), 'nginx')

    return name


def rio_chk_updateStrategy(stack, sname):
    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    return inspect["updateStrategy"]


def kube_chk_updateStrategy(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['strategy']
    results = cnt['type']

    return results


def test_updateStrategy1(stack):
    service_name = run_updateStrategy(stack, '--update-strategy', 'rolling')

    rio_updateStrategy = rio_chk_updateStrategy(stack, service_name)
    assert rio_updateStrategy == "rolling"

    kube_updateStrategy = kube_chk_updateStrategy(stack, service_name)
    assert kube_updateStrategy == "RollingUpdate"


# bug 162
# def test_updateStrategy2(stack):
#    service_name = run_updateStrategy(stack, '--update-strategy', 'on-delete')
#
#    rio_updateStrategy = rio_chk_updateStrategy(stack, service_name)
#    assert rio_updateStrategy == "on-delete"
#
#    kube_updateStrategy = kube_chk_updateStrategy(stack, service_name)
#    assert kube_updateStrategy == "on-delete"
