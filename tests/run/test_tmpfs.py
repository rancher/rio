import util


def run_tmpfs(stack, *value):

    name = util.rioRun(stack, ' '.join(value), 'nginx')

    return name


def rio_chk_tmpfs(stack, sname):
    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    return inspect["tmpfs"][0]["path"]


def kube_chk_tmpfs(stack, service):
    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", service)
    cnt = obj['spec']['template']['spec']['containers'][0]
    results = cnt['volumeMounts'][0]['mountPath']

    return results


def test_tty1(stack):
    service_name = run_tmpfs(stack, '--tmpfs /tmp')

    rio_tty = rio_chk_tmpfs(stack, service_name)
    assert rio_tty == "/tmp"

#    Bug #159
#    kube_tty = kube_chk_tmpfs(stack, service_name)
#    assert kube_tty is True
