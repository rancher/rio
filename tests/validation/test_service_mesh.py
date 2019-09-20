# Run Validation test.  Use functions to test run and get output

import util
import time


def create_service(nspc, image):
    port = "-p 80/http"
    srv = util.rioRun(nspc, port, image)
    fullName = (f"{nspc}/{srv}")

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


def wait_for_revision(fullName, number):
    for i in range(1, 120):
        length = util.run(f"rio inspect --type app --format json {fullName} "
                          f"| jq -r .spec.revisions | jq length")
        print(length)
        if length == number:
            break
        time.sleep(1)


def test_rio_svc_weight(nspc):
    image = "ibuildthecloud/demo:v1"
    image2 = "ibuildthecloud/demo:v3"

    fullName = create_service(nspc, image)
    util.wait_for_app(fullName)
    stage_service(image2, fullName, "v3")
    fullName1 = (f"{fullName}:v0")
    fullName2 = (f"{fullName}:v3")

    wait_for_revision(fullName, "2")

    results1 = get_app_info(fullName1, "spec.weight")
    results2 = get_app_info(fullName2, "spec.weight")

    print(f"{results1}")
    print(f"{results2}")

    assert results1 == '100'
    assert results2 == 'null'


def test_rio_svc_weight2(nspc):
    image = "ibuildthecloud/demo:v1"
    image2 = "ibuildthecloud/demo:v3"

    fullName = create_service(nspc, image)
    util.wait_for_app(fullName)
    stage_service(image2, fullName, "v3")
    fullName1 = f"{fullName}:v0"
    fullName2 = f"{fullName}:v3"

    wait_for_revision(fullName, "2")

    change_weight(fullName, "v3", "5%")

    results1 = get_app_info(fullName1, "spec.weight")
    results2 = get_app_info(fullName2, "spec.weight")

    assert results1 == '95'
    assert results2 == '5'
