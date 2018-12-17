from os import unlink
from random import randint
import util


def run_image_pull(stack, img):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = "%s/%s" % (stack, name)

    command = (f'rio run -n {fullName}')
    command += " --image-pull-policy " + img

    command += " nginx"
    util.run(command)
    util.run(f"rio wait {fullName}")

    print(command)

    return name


def rio_chk(stack, sname):
    print(sname)
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['imagePullPolicy']


def kube_chk(stack, sname):
    print(sname)
    fullName = (f"{stack}/{sname}")

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", sname)

    return obj['spec']['template']['spec']['containers'][0]['imagePullPolicy']


def test_content(stack):
    service_name = run_image_pull(stack, 'always')
    print(service_name)

    gotrio = rio_chk(stack, service_name)
    assert gotrio == 'always'

    gotk8s = kube_chk(stack, service_name)
    assert gotk8s == 'Always'


def test_content2(stack):
    service_name = run_image_pull(stack, 'never')
    print(service_name)

    gotrio = rio_chk(stack, service_name)
    assert gotrio == 'never'

    gotk8s = kube_chk(stack, service_name)
    assert gotk8s == 'Never'


def test_content3(stack):
    service_name = run_image_pull(stack, 'not-present')
    print(service_name)

    gotrio = rio_chk(stack, service_name)
    assert gotrio == 'not-present'

    gotk8s = kube_chk(stack, service_name)
    assert gotk8s == 'IfNotPresent'
