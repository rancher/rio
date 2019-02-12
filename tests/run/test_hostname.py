import util


def run_hostname(stack, value, value2):
    name = util.rioRun(stack, value, value2, 'nginx')

    return name


def rio_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['hostname']


def kube_chk(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']
    results = cnt['hostname']

    return results


def test_hostname1(stack):
    value = "--hostname"
    value2 = "chost"

    service_name = run_hostname(stack, value, value2)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == "chost"

    kube_got = kube_chk(stack, service_name)
    assert kube_got == "chost"
