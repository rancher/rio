# Run Validation test.  Use functions to test run and get output

import util
import time


def test_rio_scale(service):
    fullName = (f"{service}:v0")
    scalefield = "spec.scale"
    print(f"{fullName}")

    for i in range(1, 60):
        inspect = util.rioInspect(fullName, scalefield)
        if inspect != "":
            break
        time.sleep(1)
    assert inspect == '1'


def test_rio_image(service):
    fullName = (f"{service}:v0")
    scalefield = "spec.image"
    print(f"{fullName}")

    for i in range(1, 60):
        inspect = util.rioInspect(fullName, scalefield)
        if inspect != "":
            break
        time.sleep(1)
    assert inspect == 'nginx'


def test_rio_weight(service):
    fullName = (f"{service}:v0")
    scalefield = "spec.weight"
    print(f"{fullName}")

    for i in range(1, 60):
        inspect = util.rioInspect(fullName, scalefield)
        if inspect != "":
            break
        time.sleep(1)
    assert inspect == '100'
