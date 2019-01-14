from random import randint
import util


def run_readonly_setup(stack, option=''):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")

    cmd = (f'rio run -n {fullName} {option}ubuntu')

    util.run(cmd)
    util.run(f"rio wait {fullName}")

    return name


def riotest(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['readOnly']


def kubetest(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    container = obj['spec']['template']['spec']['containers'][0]
    results = container['securityContext']['readOnlyRootFilesystem']

    return results


def test_privileged(stack):
    value = "--read-only -i "
    serviceName = run_readonly_setup(stack, value)

    assert riotest(stack, serviceName) is True

    assert kubetest(stack, serviceName) is True


def test_privileged2(stack):
    value = "-i "
    serviceName = run_readonly_setup(stack, value)

    assert riotest(stack, serviceName) is False

    assert kubetest(stack, serviceName) is False
