from os import unlink
from random import randint
import util
import tempfile


def env_file_setup(stack, *envs):

    fp = tempfile.NamedTemporaryFile(delete=False)

    for c in envs:
        fp.write(bytes(c+"\n", 'utf8'))

    fp.close()

    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")
    cmd = (f'rio run -n {fullName}')

    cmd += " --env-file " + fp.name + " nginx"

    util.runwait(cmd, fullName)

    unlink(fp.name)

    return name


def rio_chk(stack, sname):
    print(sname)
    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    return inspect['environment']


def kube_chk(stack, sname):
    print(sname)
    fullName = (f"{stack}/{sname}")

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", sname)
    container = obj['spec']['template']['spec']['containers'][0]

    out = []
    for item in container['env']:
        out.append(item['name'])
        out.append(item['value'])

    print(out)

    return out


def test_content(stack):
    service_name = env_file_setup(stack, 'foo=bar')
    print(service_name)

    gotrio = rio_chk(stack, service_name)
    assert gotrio == ['foo=bar']

    print(service_name)
    gotk8s = kube_chk(stack, service_name)
    assert gotk8s == ['foo', 'bar']


def test_content2(stack):
    service_name = env_file_setup(stack, 'foo=bar', 'foo2=bar2')
    print(service_name)

    gotrio = rio_chk(stack, service_name)
    assert gotrio == ['foo=bar', 'foo2=bar2']

    print(service_name)
    gotk8s = kube_chk(stack, service_name)
    assert gotk8s == ['foo', 'bar', 'foo2', 'bar2']
