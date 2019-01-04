from os import unlink
from random import randint
import util
import tempfile


def lbl_file_setup(stack, *lbl):

    fp = tempfile.NamedTemporaryFile(delete=False)

    for c in lbl:
        fp.write(bytes(c+"\n", 'utf8'))

    fp.close()

    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")
    cmd = (f'rio run -n {fullName}')

    cmd += " --label-file " + fp.name + " nginx"

    util.run(cmd)
    util.run(f"rio wait {fullName}")

    unlink(fp.name)

    return name


def rio_chk(stack, sname):
    print(sname)
    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    return inspect['labels']


def kube_chk(stack, sname, *lbl):
    print(sname)
    fullName = (f"{stack}/{sname}")

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", sname)

    results = ""
    for c in lbl:
        results += obj['spec']['template']['metadata']['labels'][f"{lbl}"]

    return results


def test_content(stack):
    service_name = lbl_file_setup(stack, 'foo=bar')

    gotrio = rio_chk(stack, service_name)
    assert gotrio == {'foo': 'bar'}

    print(service_name)
    gotk8s = kube_chk(stack, service_name, "foo")
    assert gotk8s == 'bar'


def test_content2(stack):
    service_name = lbl_file_setup(stack, 'foo=bar', 'foo2=bar2')
    print(service_name)

    gotrio = rio_chk(stack, service_name)
    assert gotrio == {'foo': 'bar', 'foo2': 'bar2'}

    gotk8s = kube_chk(stack, service_name, "foo2")
    assert gotk8s == ['foo', 'bar', 'foo2', 'bar2']
