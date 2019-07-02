import util


def set_scale(service, scale):
    cmd = (f"rio scale {service}={scale}")

    util.run(cmd)


def rio_return_scale(service):
    scalefield = "spec.scale"

    inspect = util.rioInspect(service, scalefield)
    return inspect


def test_scale3(service):
    set_scale(service, 3)

    scale_amount = rio_return_scale(service)
    assert scale_amount == '3'


def test_rio_scale5(service):
    set_scale(service, 5)

    scale_amount = rio_return_scale(service)
    assert scale_amount == '5'


def test_rio_scale10(service):
    set_scale(service, 10)

    scale_amount = rio_return_scale(service)
    assert scale_amount == '10'
