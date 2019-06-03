import util


def run_voldriver(stack, *value):

    name = util.rioRun(stack, ' '.join(value), 'nginx')

    return name


def rio_chk_voldriver(stack, sname):
    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    return inspect["defaultVolumeDriver"]


def kube_chk_voldriver(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]
    results = cnt['defaultVolumeDriver']

    return results


def test_voldriver1(stack):
    service_name = run_voldriver(stack, '--volume-driver', 'local')

    rio_voldriver = rio_chk_voldriver(stack, service_name)
    assert rio_voldriver == "local"

    kube_voldriver = kube_chk_voldriver(stack, service_name)
    assert kube_voldriver == "local"
