import util


def run_workdir(stack, *value):

    name = util.rioRun(stack, ' '.join(value), 'nginx')

    return name


def rio_chk_workdir(stack, sname):
    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    return inspect["workingDir"]


def kube_chk_workdir(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]
    results = cnt['workingDir']

    return results


def test_workdir1(stack):
    service_name = run_workdir(stack, '--workdir', '/foo')

    rio_workdir = rio_chk_workdir(stack, service_name)
    assert rio_workdir == "/foo"

    kube_workdir = kube_chk_workdir(stack, service_name)
    assert kube_workdir == "/foo"
