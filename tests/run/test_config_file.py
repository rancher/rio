import util
import random


def config_setup(stack, *text):

    config = util.rioConfigCreate(stack, *text)
    fullname = (f"{stack}/{config}")

    return config


def run_config(stack, config_name):
    name = "tsrv" + str(random.randint(1000, 5000))
    fullName = "%s/%s" % (stack, name)

    tempdir = ":/temp" + str(random.randint(100, 999))

    cmd = (f'rio run -n {fullName}')
    cmd += (f' --config {config_name}{tempdir}') 
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
            out.append(str(item['configMap']['name']).split("-")[0])

    out.sort()
    print(out)

    return out


def test_config_name(stack):
    config_name1 = config_setup(stack, "2foo=2bar", "2foo1=2bar2")
    servicename = run_config(stack, config_name1)

    gotrio = rio_chk(stack, servicename)
    assert gotrio == [config_name1]

    gotk8s = kube_chk(stack, servicename)
    assert gotk8s == [config_name1]
