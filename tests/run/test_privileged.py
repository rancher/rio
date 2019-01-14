from random import randint
import util


def run_privileged_setup(stack, option=''):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")

    cmd = (f'rio run -n {fullName} {option}nginx')

    util.run(cmd)
    util.run(f"rio wait {fullName}")

    return name


def riotest(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['privileged']


def kubetest(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    container = obj['spec']['template']['spec']['containers'][0]
    results = container['securityContext']['privileged']

    return results


def test_privileged(stack):
    value = "--privileged "
    serviceName = run_privileged_setup(stack, value)

    assert riotest(stack, serviceName) is True

    assert kubetest(stack, serviceName) is True


def test_privileged2(stack):
    serviceName = run_privileged_setup(stack)

    assert riotest(stack, serviceName) is False

    assert kubetest(stack, serviceName) is False
