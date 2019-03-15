from random import randint
import util


def run_healthcmd_setup(stack, cmd, interval='',
                        retries='', start='', timeout=''):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")

    options = (f'{interval}{retries}{start}{timeout}')

    rcmd = (f'rio run -n {fullName} --health-cmd {cmd} {options}nginx')

    util.run(rcmd)
    util.run(f"rio wait {fullName}")

    return name


def riocmd(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['healthcheck']['test']


def riointerval(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['healthcheck']['intervalSeconds']


def rioretries(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['healthcheck']['healthyThreshold']


def riosperiod(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['healthcheck']['initialDelaySeconds']


def riotimeout(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['healthcheck']['timeoutSeconds']


def kubecmd(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]['livenessProbe']
    results = cnt['exec']['command']

    return results


def kubeinterval(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]['livenessProbe']
    results = cnt['periodSeconds']

    return results


def kuberetries(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]['livenessProbe']
    results = cnt['successThreshold']

    return results


def kubesperiod(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]['livenessProbe']
    results = cnt['initialDelaySeconds']

    return results


def kubetimeout(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]['livenessProbe']
    results = cnt['timeoutSeconds']

    return results


def test_healthcmd(stack):
    cmd = "'echo hello'"
    serviceName = run_healthcmd_setup(stack, cmd)

    assert riocmd(stack, serviceName) == ['CMD-SHELL', 'echo hello']

    assert kubecmd(stack, serviceName) == ['sh', '-c', 'echo hello']


def test_healthcmd2(stack):
    cmd = "'echo hello'"
    rintvl = '--health-interval 6s '
    rrtries = '--health-retries 1 '
    rsp = '--health-start-period 8s '
    rto = '--health-timeout 9s '
    serviceName = run_healthcmd_setup(stack, cmd, rintvl, rrtries, rsp, rto)

    assert riocmd(stack, serviceName) == ['CMD-SHELL', 'echo hello']
    assert riointerval(stack, serviceName) == 6
    assert rioretries(stack, serviceName) == 1
    assert riosperiod(stack, serviceName) == 8
    assert riotimeout(stack, serviceName) == 9

    assert kubecmd(stack, serviceName) == ['sh', '-c', 'echo hello']
    assert kubeinterval(stack, serviceName) == 6
    assert kuberetries(stack, serviceName) == 1
    assert kubesperiod(stack, serviceName) == 8
    assert kubetimeout(stack, serviceName) == 9
