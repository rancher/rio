from random import randint
import util


def run_dns(stack, *dns):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")
    cmd = (f'rio run -n {fullName}')

    for c in dns:
        cmd += " --dns " + c

    cmd += " nginx"

    print(cmd)
    util.runwait(cmd, fullName)
    print(name)

    return name


def rio_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['dns']


def kube_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", sname)

    return obj['spec']['template']['spec']['dnsConfig']['nameservers']


def test_content(stack):
    service_name = run_dns(stack, '1.1.1.1')

    gotrio = rio_chk(stack, service_name)
    assert gotrio == ['1.1.1.1']

    gotk8s = kube_chk(stack, service_name)
    assert gotk8s == ['1.1.1.1']


def test_content2(stack):
    service_name = run_dns(stack, "1.1.1.1", "2.2.2.2")

    got = rio_chk(stack, service_name)
    assert got == ['1.1.1.1', '2.2.2.2']

    gotk8s = kube_chk(stack, service_name)
    assert gotk8s == ['1.1.1.1', '2.2.2.2']
