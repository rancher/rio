from random import randint
import util


def run_scale(stack, value):

    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")
    cmd = (f'rio run -n {fullName} --scale {value} nginx')

    util.run(cmd)
    util.run(f"rio wait {fullName}")

    return name


def rio_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['scale']


def kube_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", sname)

    container = obj['spec']['replicas']
    return container


def test_scale(stack):
    value = "1"
    service_name = run_scale(stack, value)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == 1

    k8s_got = kube_chk(stack, service_name)
    assert k8s_got == 1


def test_scale2(stack):
    value = "3"
    service_name = run_scale(stack, value)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == 3

    k8s_got = kube_chk(stack, service_name)
    assert k8s_got == 3


def test_scale3(stack):
    value = "10"
    service_name = run_scale(stack, value)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == 10

    k8s_got = kube_chk(stack, service_name)
    assert k8s_got == 10
