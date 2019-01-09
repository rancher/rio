from random import randint
import util


def run_metadata(stack, *mdata):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = "%s/%s" % (stack, name)

    command = (f'rio run -n {fullName}')

    for d in mdata:
        command += " --metadata " + d

    command += " nginx"
    util.run(command)
    util.run(f"rio wait {fullName}")

    return name


def rio_chk(stack, sname):
    print(sname)
    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    return inspect['metadata']


def kube_chk(stack, sname, mdata):
    print(sname)
    fullName = (f"{stack}/{sname}")

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", sname)
    container = obj['spec']['template']['metadata']['annotations'][f"{mdata}"]

    return container


def test_content(stack):
    service_name = run_metadata(stack, 'foo=bar')

    gotrio = rio_chk(stack, service_name)
    assert gotrio == {'foo': 'bar'}

    print(service_name)
    gotk8s = kube_chk(stack, service_name, "foo")
    assert gotk8s == 'bar'


def test_content2(stack):
    service_name = run_metadata(stack, 'foo=bar', 'foo2=bar2')
    print(service_name)

    gotrio = rio_chk(stack, service_name)
    assert gotrio == {'foo': 'bar', 'foo2': 'bar2'}

#    print(service_name)
#    gotk8s = kube_chk(stack, service_name)
#    assert gotk8s == ['foo', 'bar', 'foo2', 'bar2']
