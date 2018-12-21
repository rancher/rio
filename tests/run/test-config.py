from os import unlink
from random import randint
import util
import tempfile


def config_setup(stack, *configs):
    config_name = "tconfig" + str(randint(1000, 5000))

    fp = tempfile.NamedTemporaryFile(delete=False)

    for c in configs:
        fp.write(bytes(c+"\n", 'utf8'))

    fp.close()

    util.run(f"rio config create {stack}/{config_name} {fp.name}")
    unlink(fp.name)

    return config_name


def run_config(stack, config_names):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = "%s/%s" % (stack, name)
    cmd = (f'rio run -n {fullName}')

    for c in config_names:
        tempdir = ":/temp" + str(randint(100, 999))
        cmd += " --config " + c + tempdir

    cmd += " nginx"

    print(cmd)
    util.run(cmd)
    util.run(f"rio wait {fullName}")
    print(name)

    return name


def rio_chk(stack, sname):
    fullName = "%s/%s" % (stack, sname)

    inspect = util.rioInspect(fullName)

    out = []
    for item in inspect["configs"]:
        out.append(item["source"])

    out.sort()

    return out


def kube_chk(stack, service_name):
    fullName = "%s/%s" % (stack, service_name)

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", service_name)

    out = []
    for item in obj['spec']['template']['spec']['volumes']:
        if 'configMap' in item:
            out.append(item['configMap']['name'])

    out.sort()
    print(out)

    return out


def test_content(stack):
    config_name1 = config_setup(stack, "1foo=1bar", "1foo2=1bar2")
    config_setup(stack, "2foo=2bar", "2foo1=2bar2")

    expect = [config_name1]
    expect.sort()

    servicename = run_config(stack, expect)
    print(stack, servicename)

    gotrio = rio_chk(stack, servicename)
    assert expect == gotrio

    gotk8s = kube_chk(stack, servicename)
    assert expect == gotk8s
