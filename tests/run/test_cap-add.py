import util
from os import system
from random import randint


def run_capadd(stack, *value):

    name = util.rioRun(stack, ' '.join(value), 'nginx')

    return name


def rio_capadd_chk(stack, service_name):
    fullName = "%s/%s" % (stack, service_name)
    got = util.rioInspect(fullName, "capAdd")
    print(got)

    return got


def kube_capadd_chk(stack, service_name, *capabilities):
    fullName = "%s/%s" % (stack, service_name)

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", service_name)

    container = obj['spec']['template']['spec']['containers'][0]
    return container['securityContext']['capabilities']['add']


def test_cap_1(stack):
    service_name = run_capadd(stack, "--cap-add", "ALL")

    r = rio_capadd_chk(stack, service_name)
    assert r == "[ALL]"

    k = kube_capadd_chk(stack, service_name, "ALL")
    assert k == ['ALL']


def test_cap_2(stack):
    service_name = run_capadd(stack, "--cap-add", "ALL", "--cap-add", "SYSLOG")

    r = rio_capadd_chk(stack, service_name)
    assert r == "[ALL SYSLOG]"

    k = kube_capadd_chk(stack, service_name, "ALL", "SYSLOG")
    assert k == ['ALL', 'SYSLOG']
