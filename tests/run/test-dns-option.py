from random import randint
import util


def run_dns(stack, *dns_options):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")
    cmd = (f'rio run -n {fullName}')

    for c in dns_options:
        cmd += " --dns-option " + c

    cmd += " nginx"

    print(cmd)
    util.run(cmd)
    util.run(f"rio wait {fullName}")
    print(name)

    return name


def rio_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['dnsOptions']


def kube_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", sname)

    out = []
    for item in obj['spec']['template']['spec']['dnsConfig']['options']:
        out.append(item['name'])

    return out


def test_content(stack):
    service_name = run_dns(stack, "debug")

    gotrio = rio_chk(stack, service_name)
    assert gotrio == ['debug']

    gotk8s = kube_chk(stack, service_name)
    assert gotk8s == ['debug']


def test_content2(stack):
    service_name = run_dns(stack, "debug", "attempts:2")

    got = rio_chk(stack, service_name)
    assert got == ['debug', 'attempts:2']

    gotk8s = kube_chk(stack, service_name)
    assert gotk8s == ['debug', 'attempts:2']
