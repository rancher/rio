from random import randint
import util


def run_memlimit(stack, size, unit):

    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")
    cmd = (f'rio run -n {fullName} --memory-limit {size}{unit} nginx')
    util.run(cmd)
    util.run(f"rio wait {fullName}")

    return name


def rio_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['memoryLimitBytes']


def kube_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", sname)

    container = obj['spec']['template']['spec']['containers'][0]
    return container['resources']['limits']['memory']


def test_memlimit(stack):
    size = "100000000"
    unit = ""
    service_name = run_memlimit(stack, size, unit)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == 100000000

    k8s_got = kube_chk(stack, service_name)
    assert k8s_got == "100M"


def test_memlimit2(stack):
    size = "100000000"
    unit = "b"
    service_name = run_memlimit(stack, size, unit)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == 100000000

    k8s_got = kube_chk(stack, service_name)
    assert k8s_got == "100M"


def test_memlimit3(stack):
    size = "100000"
    unit = "k"
    service_name = run_memlimit(stack, size, unit)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == 102400000

    k8s_got = kube_chk(stack, service_name)
    assert k8s_got == "102400k"


def test_memlimit4(stack):
    size = "10"
    unit = "m"
    service_name = run_memlimit(stack, size, unit)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == 10485760

    k8s_got = kube_chk(stack, service_name)
    assert k8s_got == "10485760"


def test_memlimit5(stack):
    size = "1"
    unit = "g"
    service_name = run_memlimit(stack, size, unit)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == 1073741824

    k8s_got = kube_chk(stack, service_name)
    assert k8s_got == "1073741824"
