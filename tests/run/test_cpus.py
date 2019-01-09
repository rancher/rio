# from os import unlink
from random import randint
import util
from time import sleep
# import tempfile


def run_cpus(stack, value):

    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")
    cmd = (f'rio run -n {fullName}')
    value = str(value)

    cmd += " --cpus " + value + " nginx"

    print(cmd)
    util.run(cmd)
#    util.run(f"rio wait {fullName}")
    sleep(5)
    print(name)

    return name


def rio_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['nanoCpus']


def kube_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", sname)

    container = obj['spec']['template']['spec']['containers'][0]
    return container['resources']['requests']['cpu']


def test_cpus(stack):
    cpus = 0
    service_name = run_cpus(stack, cpus)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == (f"{cpus}")

#    k8s_got = kube_chk(stack, service_name)
#    assert k8s_got == (f"{cpus}")


def test_cpus2(stack):
    cpus = 1
    service_name = run_cpus(stack, cpus)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == (f"{cpus}")

#    k8s_got = kube_chk(stack, service_name)
#    assert k8s_got == (f"{cpus}")


def test_cpus3(stack):
    cpus = 10
    service_name = run_cpus(stack, cpus)

    rio_got = rio_chk(stack, service_name)
    assert rio_got == (f"{cpus}")

#    k8s_got = kube_chk(stack, service_name)
#    assert k8s_got == (f"{cpus}")
