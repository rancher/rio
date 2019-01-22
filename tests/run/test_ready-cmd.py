from random import randint
import util


def run_readycmd_setup(stack, cmd, interval='',
                       retries='', start='', timeout=''):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")

    options = (f'{interval}{retries}{start}{timeout}')

    cmd = (f'rio run -n {fullName} --ready-cmd {cmd} {options}nginx')

    util.run(cmd)
    util.run(f"rio wait {fullName}")

    return name


def riocmd(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['readycheck']['test']


def riointerval(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['readycheck']['intervalSeconds']


def rioretries(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['readycheck']['healthyThreshold']


def riosperiod(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['readycheck']['initialDelaySeconds']


def riotimeout(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['readycheck']['timeoutSeconds']


def kubecmd(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]['readinessProbe']
    results = cnt['exec']['command']

    return results


def kubeinterval(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]['readinessProbe']
    results = cnt['periodSeconds']

    return results


def kuberetries(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]['readinessProbe']
    results = cnt['successThreshold']

    return results


def kubesperiod(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]['readinessProbe']
    results = cnt['initialDelaySeconds']

    return results


def kubetimeout(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]['readinessProbe']
    results = cnt['timeoutSeconds']

    return results


def test_readycmd(stack):
    cmd = "'echo hello'"
    serviceName = run_readycmd_setup(stack, cmd)

    assert riocmd(stack, serviceName) == ['CMD-SHELL', 'echo hello']

    assert kubecmd(stack, serviceName) == ['sh', '-c', 'echo hello']


def test_readycmd2(stack):
    cmd = "'echo hello'"
    rintvl = '--ready-interval 6s '
    rrtries = '--ready-retries 7 '
    rsp = '--ready-start-period 8s '
    rto = '--ready-timeout 9s '
    serviceName = run_readycmd_setup(stack, cmd, rintvl, rrtries, rsp, rto)

    assert riocmd(stack, serviceName) == ['CMD-SHELL', 'echo hello']
    assert riointerval(stack, serviceName) == 6
    assert rioretries(stack, serviceName) == 7
    assert riosperiod(stack, serviceName) == 8
    assert riotimeout(stack, serviceName) == 9

    assert kubecmd(stack, serviceName) == ['sh', '-c', 'echo hello']
    assert kubeinterval(stack, serviceName) == 6
    assert kuberetries(stack, serviceName) == 7
    assert kubesperiod(stack, serviceName) == 8
    assert kubetimeout(stack, serviceName) == 9
