import util
from os import system
from random import randint


def run_capdrop(stack, *capability):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = "%s/%s" % (stack, name)

    command = (f'rio run -n {fullName}')

    for c in capability:
        command += " --cap-drop " + c

    command += " nginx"
    print(command)
    system(command)
    system("rio wait %s" % fullName)

    return name


def rio_capdrop_chk(stack, service_name):
    fullName = "%s/%s" % (stack, service_name)
    got = util.rioInspect(fullName, "capDrop")
    print(got)

    return got


def kube_capdrop_chk(stack, service_name, *capabilities):
    fullName = "%s/%s" % (stack, service_name)

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", service_name)

    container = obj['spec']['template']['spec']['containers'][0]
    return container['securityContext']['capabilities']['drop']


def test_cap_1(stack):
    service_name = run_capdrop(stack, "AUDIT_CONTROL")

    r = rio_capdrop_chk(stack, service_name)
    assert r == "[AUDIT_CONTROL]"

    k = kube_capdrop_chk(stack, service_name, "AUDIT_CONTROL")
    assert k == ['AUDIT_CONTROL']


def test_cap_2(stack):
    service_name = run_capdrop(stack, "AUDIT_CONTROL", "SYSLOG")

    r = rio_capdrop_chk(stack, service_name)
    assert r == "[AUDIT_CONTROL SYSLOG]"

    k = kube_capdrop_chk(stack, service_name, "AUDIT_CONTROL", "SYSLOG")
    assert k == ['AUDIT_CONTROL', 'SYSLOG']
