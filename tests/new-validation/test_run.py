# Run Validation test.  Use functions to test run and get output

import util


def test_rio_scale(service):
    fullName = (f"{service}:v0")
    scalefield = "spec.scale"
    print(f"{fullName}")

    inspect = util.rioInspect(fullName, scalefield)
    assert inspect == '1'


def test_rio_image(service):
    fullName = (f"{service}:v0")
    scalefield = "spec.image"
    print(f"{fullName}")

    inspect = util.rioInspect(fullName, scalefield)
    assert inspect == 'nginx'


def test_rio_weight(service):
    fullName = (f"{service}:v0")
    scalefield = "spec.weight"
    print(f"{fullName}")

    inspect = util.rioInspect(fullName, scalefield)
    assert inspect == '100'
