from random import randint
import util


def run_mem(stack, size, unit):

    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")
    cmd = (f'rio run -n {fullName} --memory {size}{unit} nginx')
    util.runwait(cmd, fullName)

    return name


def rio_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['memoryReservationBytes']


def kube_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", sname)

    container = obj['spec']['template']['spec']['containers'][0]
    return container['resources']['requests']['memory']


def test_mem(stack):
    size = "100000000"
    unit = ""
    service_name = run_mem(stack, size, unit)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == 100000000

    k8s_got = kube_chk(stack, service_name)
    assert k8s_got == "100M"


def test_mem2(stack):
    size = "100000000"
    unit = "b"
    service_name = run_mem(stack, size, unit)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == 100000000

    k8s_got = kube_chk(stack, service_name)
    assert k8s_got == "100M"


def test_mem3(stack):
    size = "100000"
    unit = "k"
    service_name = run_mem(stack, size, unit)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == 102400000

    k8s_got = kube_chk(stack, service_name)
    assert k8s_got == "102400k"


def test_mem4(stack):
    size = "10"
    unit = "m"
    service_name = run_mem(stack, size, unit)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == 10485760

    k8s_got = kube_chk(stack, service_name)
    assert k8s_got == "10485760"


def test_mem5(stack):
    size = "1"
    unit = "g"
    service_name = run_mem(stack, size, unit)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == 1073741824

    k8s_got = kube_chk(stack, service_name)
    assert k8s_got == "1073741824"
