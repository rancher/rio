import util


def run_secret(stack, *value):

    name = util.rioRun(stack, ' '.join(value), 'nginx')

    return name


def rio_chk_source(stack, sname):
    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    out = []
    for item in inspect["secrets"]:
        out.append(item["source"])

    out.sort()

    return out


def rio_chk_target(stack, sname):
    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    out = []
    for item in inspect["secrets"]:
        out.append(item["target"])

    out.sort()

    return out


def kube_chk_mountName(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]
    results = cnt['volumeMounts'][0]['mountPath']

    out = []
    for item in cnt['volumeMounts']:
        out.append(item['name'])

    out.sort()

    return out


def kube_chk_mountPath(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]
    results = cnt['volumeMounts'][0]['mountPath']

    out = []
    for item in cnt['volumeMounts']:
        out.append(item['mountPath'])

    out.sort()

    return out


def test_ipc1(stack):
    service_name = run_secret(stack, '--secret', 'foo:bar')

    rio_source = rio_chk_source(stack, service_name)
    rio_target = rio_chk_target(stack, service_name)

    assert rio_source == ['foo']
    assert rio_target == ['bar']

    kube_mountName = kube_chk_mountName(stack, service_name)
    kube_mountPath = kube_chk_mountPath(stack, service_name)

    assert kube_mountName == ['secret-foo']
    assert kube_mountPath == ['bar']


def test_ipc2(stack):
    value1 = 'foo:bar'
    value2 = 'foo2:bar2'
    service_name = run_secret(stack, '--secret', value1, '--secret', value2)

    rio_source = rio_chk_source(stack, service_name)
    rio_target = rio_chk_target(stack, service_name)

    assert rio_source == ['foo', 'foo2']
    assert rio_target == ['bar', 'bar2']

    kube_mountName = kube_chk_mountName(stack, service_name)
    kube_got = kube_chk_mountPath(stack, service_name)

    assert kube_mountName == ['secret-foo', 'secret-foo2']
    assert kube_got == ['bar', 'bar2']
