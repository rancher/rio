from random import randint
import util


def run_entrypoint(stack, *entrypoint):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")
    cmd = (f'rio run -n {fullName}')

    for c in entrypoint:
        cmd += " -t -i --entrypoint " + c

    cmd += " nginx"

    print(cmd)
    util.run(cmd)
    util.run(f"rio wait {fullName}")
    print(name)

    return name


def rio_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['entrypoint']


def kube_chk(stack, sname):
    fullName = (f"{stack}/{sname}")

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", sname)

    return obj['spec']['template']['spec']['containers'][0]['command']


def test_content(stack):
    service_name = run_entrypoint(stack, 'sh')

    gotrio = rio_chk(stack, service_name)
    assert gotrio == ['sh']

    gotk8s = kube_chk(stack, service_name)
    assert gotk8s == ['sh']


# def test_content2(stack):
#    service_name = run_entrypoint(stack, "echo", "hello")

#    got = rio_chk(stack, service_name)
#    assert got == ['echo', 'hello']

#    gotk8s = kube_chk(stack, service_name)
#    assert gotk8s == ['echo', '$HOME']
