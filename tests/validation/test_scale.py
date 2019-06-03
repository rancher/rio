import util
import time


def set_scale(stack, service, scale):
    fullName = "%s/%s" % (stack, service)
    cmd = (f"rio scale {fullName}={scale}")
    util.runwait(cmd, fullName)


def rio_return_scale(stack, service):
    fullName = "%s/%s" % (stack, service)
    scale_amount = util.rioInspect(fullName, "scale")

    return scale_amount


def kube_return_scale(stack, service):
    fullName = "%s/%s" % (stack, service)

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", service)
    replicas = obj['status']['replicas']

    return replicas


def test_scale3(stack, service):
    set_scale(stack, service, 3)

    scale_amount = rio_return_scale(stack, service)
    assert scale_amount == '3'

    time.sleep(5)

    replicas = kube_return_scale(stack, service)
    assert replicas == 3


def test_rio_scale5(stack, service):
    set_scale(stack, service, 5)

    scale_amount = rio_return_scale(stack, service)
    assert scale_amount == '5'


def test_rio_scale10(stack, service):
    set_scale(stack, service, 10)

    scale_amount = rio_return_scale(stack, service)
    assert scale_amount == '10'
