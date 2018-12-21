import util
import time


def set_scale(stack, service, scale):
    fullName = "%s/%s" % (stack, service)
    util.run("rio scale %s=%s" % (fullName, scale))
    util.run("rio wait %s" % fullName)


def rio_return_scale(stack, service, scale):
    fullName = "%s/%s" % (stack, service)
    set_scale(stack, service, scale)
    scale_amount = util.rioInspect(fullName, "scale")

    return scale_amount


def kube_return_scale(stack, service, scale):
    fullName = "%s/%s" % (stack, service)
    print(scale)
    set_scale(stack, service, scale)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", service)
    replicas = obj['status']['replicas']

    return replicas


def test_kube_scale(stack, service):
    replicas = kube_return_scale(stack, service, 3)
    time.sleep(5)
    assert replicas == 3


def test_rio_scale3(stack, service):
    scale_amount = rio_return_scale(stack, service, 3)
    assert scale_amount == '3'


def test_rio_scale5(stack, service):
    scale_amount = rio_return_scale(stack, service, 5)
    assert scale_amount == '5'


def test_rio_scale10(stack, service):
    scale_amount = rio_return_scale(stack, service, 10)
    assert scale_amount == '10'
