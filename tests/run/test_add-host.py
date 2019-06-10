import util


def run_addhost(stack, *value):

    name = util.rioRun(stack, ' '.join(value), 'nginx')

    return name


def rio_chk_addhost(stack, sname):
    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    return inspect["extraHosts"]


def kube_chk_addhost(stack, sname):
    fullName = (f"{stack}/{sname}")
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", sname)
    cnt = obj['spec']['template']['spec']

    out = []
    for item in cnt['hostAliases']:
        out.append(item['hostnames'])
        out.append(item['ip'])

#    hostname = [0]
#    ip = cnt['ip']
#    results = (f"{ip}:{hostname}")

    return out


def test_addhost1(stack):
    service_name = run_addhost(stack, '--add-host', '1.2.3.4:example.com')

    rio_addhost = rio_chk_addhost(stack, service_name)
    assert rio_addhost == ['1.2.3.4:example.com']

    kube_addhost = kube_chk_addhost(stack, service_name)
    assert kube_addhost == [['example.com'], '1.2.3.4']


def test_addhost2(stack):
    service_name = run_addhost(stack, '--add-host', '1.2.3.4:example.com', 
                               '--add-host', '2.3.4.5:bexample.com')

    rio_addhost = rio_chk_addhost(stack, service_name)
    assert rio_addhost == ['1.2.3.4:example.com', '2.3.4.5:bexample.com']

    kube_addhost = kube_chk_addhost(stack, service_name)
    result = [['example.com'], '1.2.3.4', ['bexample.com'], '2.3.4.5']
    assert kube_addhost == result
