from os import unlink
from random import randint
import util


def run_expose(stack, *prt):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = "%s/%s" % (stack, name)

    command = (f'rio run -n {fullName}')

    for c in prt:
        command += " --expose " + c

    command += " nginx"
    util.run(command)
    util.run(f"rio wait {fullName}")

    print(command)

    return name


def rio_chk(stack, sname):
    print(sname)
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    out = []
    for item in inspect["expose"]:
        out.append(item["targetPort"])
        out.append(item["protocol"])

    return out


def kube_chk(stack, sname):
    print(sname)
    fullName = (f"{stack}/{sname}")

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", sname)
    container = obj['spec']['template']['spec']['containers'][0]

    out = []
    for item in container['ports']:
        out.append(item['containerPort'])
        out.append(item['protocol'])

    print(out)

    return out


def test_content(stack):
    service_name = run_expose(stack, '22/tcp')
    print(service_name)

    gotrio = rio_chk(stack, service_name)
    assert gotrio == [22, 'tcp']

    gotk8s = kube_chk(stack, service_name)
    assert gotk8s == [22, 'TCP']


def test_content2(stack):
    service_name = run_expose(stack, '22/tcp', '80/udp')
    print(service_name)

    gotrio = rio_chk(stack, service_name)
    assert gotrio == [22, 'tcp', 80, 'udp']

    gotk8s = kube_chk(stack, service_name)
    assert gotk8s == [22, 'TCP', 80, 'UDP']
