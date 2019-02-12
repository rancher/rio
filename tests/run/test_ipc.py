import util


def run_ipc(stack, value, value2):
    name = util.rioRun(stack, value, value2, 'nginx')

    return name


def rio_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['ipc']


def kube_chk(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']
    results = cnt['hostIPC']

    return results


def test_ipc1(stack):
    value = "--ipc"
    value2 = "host"
    service_name = run_ipc(stack, value, value2)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == "host"

    kube_got = kube_chk(stack, service_name)
    assert kube_got is True


def test_ipc2(stack):
    value = ""
    value2 = ""
    service_name = run_ipc(stack, value, value2)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == "default"
