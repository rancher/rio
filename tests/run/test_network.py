from random import randint
import util


def run_network_setup(stack, value):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")

    cmd = (f'rio run -n {fullName} --network {value} nginx')

    util.run(cmd)
    util.run(f"rio wait {fullName}")

    return name


def riotest(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['net']


def kubetest(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    results = obj['spec']['template']['spec']['hostNetwork']

    return results


def test_network(stack):
    networkValue = "host"
    serviceName = run_network_setup(stack, networkValue)

    assert riotest(stack, serviceName) == networkValue

    assert kubetest(stack, serviceName) is True


def test_network2(stack):
    networkValue = "default"
    serviceName = run_network_setup(stack, networkValue)

    assert riotest(stack, serviceName) == networkValue
