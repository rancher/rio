import util


def run_interactive(stack, value):
    name = util.rioRun(stack, value, 'nginx')

    return name


def rio_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['stdinOpen']


def kube_chk(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]
    results = cnt['stdin']

    return results


def test_interactive1(stack):
    value = "--interactive"
    service_name = run_interactive(stack, value)

    rio_got = rio_chk(stack, service_name)
    assert rio_got is True

    kube_got = kube_chk(stack, service_name)
    assert kube_got is True


def test_interactive2(stack):
    value = ""
    service_name = run_interactive(stack, value)

    rio_got = rio_chk(stack, service_name)
    assert rio_got is False
