from random import randint
import util


def rio_publish_setup(stack, pvalue, tvalue):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = "%s/%s" % (stack, name)

    cmd = (f'rio run -n {fullName} --publish {pvalue}:{tvalue} nginx')

    util.run(cmd)
    print(cmd)
    util.run(f"rio wait {fullName}")

    return name


def rio_return_port(stack, sname):

    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    return inspect['ports'][0]['port']


def rio_return_target(stack, sname):

    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    return inspect['ports'][0]['targetPort']


def kube_return_port(stack, sname):

    fullName = (f"{stack}/{sname}")
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", sname)
    container = obj['spec']['template']['spec']['containers'][0]
    results = container['ports'][0]['hostPort']

    return results


def kube_return_target(stack, sname):

    fullName = (f"{stack}/{sname}")
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", sname)
    container = obj['spec']['template']['spec']['containers'][0]
    results = container['ports'][0]['containerPort']

    return results


def test_publish1(stack):
    portValue = 8080
    targetValue = 80

    serviceName = rio_publish_setup(stack, portValue, targetValue)

    assert rio_return_port(stack, serviceName) == portValue
    assert rio_return_target(stack, serviceName) == targetValue

    assert kube_return_port(stack, serviceName) == portValue
    assert kube_return_target(stack, serviceName) == targetValue
