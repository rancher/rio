import util


def run_init(stack, value):
    name = util.rioRun(stack, value, 'nginx')

    return name


def rio_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['init']


def kube_chk(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['initContainers'][1]
    results = cnt['image']

    return results


def test_init1(stack):
    value = "--init"
    service_name = run_init(stack, value)

    rio_got = rio_chk(stack, service_name)
    assert rio_got is True

#    kube_got = kube_chk(stack, service_name)
#    assert kube_got == "nginx"


def test_init2(stack):
    value = ""
    service_name = run_init(stack, value)

    rio_got = rio_chk(stack, service_name)
    assert rio_got is False
