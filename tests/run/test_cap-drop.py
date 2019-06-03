import util
from os import system
from random import randint


def run_capdrop(stack, *value):

    name = util.rioRun(stack, ' '.join(value), 'nginx')

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
    service_name = run_capdrop(stack, "--cap-drop", "AUDIT_CONTROL")

    r = rio_capdrop_chk(stack, service_name)
    assert r == "[AUDIT_CONTROL]"

    k = kube_capdrop_chk(stack, service_name, "AUDIT_CONTROL")
    assert k == ['AUDIT_CONTROL']


def test_cap_2(stack):
    service_name = run_capdrop(stack, "--cap-drop", "AUDIT_CONTROL",
                               "--cap-drop", "SYSLOG")

    r = rio_capdrop_chk(stack, service_name)
    assert r == "[AUDIT_CONTROL SYSLOG]"

    k = kube_capdrop_chk(stack, service_name, "AUDIT_CONTROL", "SYSLOG")
    assert k == ['AUDIT_CONTROL', 'SYSLOG']
