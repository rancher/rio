from random import randint
import util


def run_readyurl_setup(stack, url, interval='',
                       retries='', start='', timeout=''):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")

    options = (f'{interval}{retries}{start}{timeout}')

    rcmd = (f'rio run -n {fullName} --ready-url {url} {options}nginx')

    util.run(rcmd)
    util.run(f"rio wait {fullName}")

    return name


def riourl(stack, sname):
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


def kubeurl(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]['readinessProbe']
    results = cnt['httpGet']['httpHeaders'][0]['value']

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


def test_readyurl(stack):
    url = "http://localhost:80"
    serviceName = run_readyurl_setup(stack, url)

    assert riourl(stack, serviceName) == ['http://localhost:80']

    assert kubeurl(stack, serviceName) == 'localhost:80'


def test_readyurl2(stack):
    url = "http://localhost:80"
    rintvl = '--ready-interval 6s '
    rrtries = '--ready-retries 7 '
    rsp = '--ready-start-period 8s '
    rto = '--ready-timeout 9s '
    serviceName = run_readyurl_setup(stack, url, rintvl, rrtries, rsp, rto)

    assert riourl(stack, serviceName) == ['http://localhost:80']
    assert riointerval(stack, serviceName) == 6
    assert rioretries(stack, serviceName) == 7
    assert riosperiod(stack, serviceName) == 8
    assert riotimeout(stack, serviceName) == 9

    assert kubeurl(stack, serviceName) == 'localhost:80'
    assert kubeinterval(stack, serviceName) == 6
    assert kuberetries(stack, serviceName) == 7
    assert kubesperiod(stack, serviceName) == 8
    assert kubetimeout(stack, serviceName) == 9
