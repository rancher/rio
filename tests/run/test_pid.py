from random import randint
import util


def run_pid_setup(stack, value):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")

    cmd = (f'rio run -n {fullName} --pid={value} nginx')

    util.run(cmd)
    util.run(f"rio wait {fullName}")

    return name


def riotest(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['pid']


def kubetest(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    results = obj['spec']['template']['spec']['pid']

    return results


def test_pid(stack):
    pidValue = "host"
    serviceName = run_pid_setup(stack, pidValue)

    assert riotest(stack, serviceName) == pidValue

#    assert kubetest(stack, serviceName) is True
#    bug #137 in rio issues, pid is not being set


def test_pid2(stack):
    pidValue = "default"
    serviceName = run_pid_setup(stack, pidValue)

    assert riotest(stack, serviceName) == pidValue
