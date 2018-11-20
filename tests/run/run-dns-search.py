from random import randint
import util


def run_dns(stack, *dns_search):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")
    cmd = (f'rio run -n {fullName}')

    for c in dns_search:
        cmd += " --dns-search " + c

    cmd += " nginx"

    print(cmd)
    util.run(cmd)
    util.run(f"rio wait {fullName}")
    print(name)

    return name


def rio_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['dnsSearch']


def kube_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", sname)

    return obj['spec']['template']['spec']['dnsConfig']['searches']


def test_content(stack):
    service_name = run_dns(stack, "example.com")

    gotrio = rio_chk(stack, service_name)
    assert gotrio == ['example.com']

    gotk8s = kube_chk(stack, service_name)
    assert gotk8s == ['example.com']


def test_content2(stack):
    service_name = run_dns(stack, "example.com", "example2.com")

    got = rio_chk(stack, service_name)
    assert got == ['example.com', 'example2.com']

    gotk8s = kube_chk(stack, service_name)
    assert gotk8s == ['example.com', 'example2.com']
