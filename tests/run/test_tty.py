import util


def run_tty(stack, *value):

    name = util.rioRun(stack, ' '.join(value), 'nginx')

    return name


def rio_chk_tty(stack, sname):
    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    return inspect["tty"]


def kube_chk_tty(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]
    results = cnt['tty']

    return results


def test_tty1(stack):
    service_name = run_tty(stack, '--tty=true')

    rio_tty = rio_chk_tty(stack, service_name)
    assert rio_tty is True

    kube_tty = kube_chk_tty(stack, service_name)
    assert kube_tty is True


def test_tty2(stack):
    service_name = run_tty(stack, '--tty=false')

    rio_tty = rio_chk_tty(stack, service_name)
    assert rio_tty is False

#    kube_tty = kube_chk_tty(stack, service_name)
#    assert kube_tty is False
