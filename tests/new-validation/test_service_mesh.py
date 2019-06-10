# Run Validation test.  Use functions to test run and get output

import util
import time


def create_service(nspc, image):
    port = "-p 80/http"
    fullName = util.rioRun(nspc, port, image)

    return fullName


def stage_service(image, fullName, version):

    util.rioStage(image, fullName, version)

    return


def get_app_info(fullName, field):

    inspect = util.rioInspect(fullName, field)

    return inspect


def change_weight(fullName, version, percent):

    cmd = (f"rio weight {fullName}:{version}={percent}")
    util.run(cmd)

    return


def test_rio_svc_weight(nspc):
    image = "ibuildthecloud/demo:v1"
    image2 = "ibuildthecloud/demo:v3"

    fullName = create_service(nspc, image)
    stage_service(image2, fullName, "v3")

    results1 = get_app_info(fullName, "status.revisionWeight.v0.weight")
    results2 = get_app_info(fullName, "status.revisionWeight.v3.weight")

    assert results1 == '100'
    assert results2 == 'null'


def test_rio_svc_weight2(nspc):
    image = "ibuildthecloud/demo:v1"
    image2 = "ibuildthecloud/demo:v3"

    fullName = create_service(nspc, image)
    stage_service(image2, fullName, "v3")
    time.sleep(5)

    change_weight(fullName, "v3", "5%")

    time.sleep(5)

    results1 = get_app_info(fullName, "status.revisionWeight.v0.weight")
    results2 = get_app_info(fullName, "status.revisionWeight.v3.weight")

    assert results1 == '95'
    assert results2 == '5'
