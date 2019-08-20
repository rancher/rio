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

    time.sleep(10)
    inspect = util.rioInspect(fullName, field)

    return inspect


def get_version_endpoint(fullName, version):

    fullNameVersion = (f"{fullName}:{version}")

    time.sleep(10)
    endpoint = "status.endpoints[0]"
    print(f"{fullNameVersion}")

    inspect = util.rioInspect(fullNameVersion, endpoint)

    return inspect


def test_rio_app_endpoint(nspc):
    image = "ibuildthecloud/demo:v1"
    image2 = "ibuildthecloud/demo:v3"

    srv = create_service(nspc, image)
    fullName = (f"{nspc}/{srv}")
    print(fullName)
    time.sleep(5)
    stage_service(image2, fullName, "v3")

    appEndpoint = get_app_info(fullName, "status.endpoints[0]")
    print(f"{appEndpoint}")

    results = util.run(f"curl -s {appEndpoint}")
    print(f"{results}")

    assert results == 'Hello World'


def test_rio_svc_endpoint1(nspc):
    image = "ibuildthecloud/demo:v1"
    image2 = "ibuildthecloud/demo:v3"

    srv = create_service(nspc, image)
    fullName = (f"{nspc}/{srv}")

    stage_service(image2, fullName, "v3")

    svcEndpoint = get_version_endpoint(fullName, "v0")
    svcEndpoint2 = get_version_endpoint(fullName, "v3")

    results1 = util.run(f"curl {svcEndpoint}")
    results2 = util.run(f'curl {svcEndpoint2}')
    print(f"{results1}")

    assert results1 == 'Hello World'
    assert results2 == 'Hello World v3'
