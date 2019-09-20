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


def get_version_endpoint(fullName, version):
    inspect = ""
    for i in range(1, 120):
        fullNameVersion = f"{fullName}:{version}"

        endpoint = "status.endpoints[0]"
        print(f"{fullNameVersion}")

        inspect = util.rioInspect(fullNameVersion, endpoint)
        if inspect != "null":
            break
        time.sleep(1)

    return inspect


def test_rio_app_endpoint(nspc):
    image = "ibuildthecloud/demo:v1"
    image2 = "ibuildthecloud/demo:v3"

    srv = create_service(nspc, image)
    fullName = (f"{nspc}/{srv}")
    print(fullName)

    util.wait_for_app(fullName)
    stage_service(image2, fullName, "v3")

    for i in range(1, 120):
        appEndpoint = get_app_info(fullName, "status.endpoints[0]")
        if appEndpoint != "null":
            break
        time.sleep(1)
    print(f"{appEndpoint}")

    util.assert_endpoint(appEndpoint, 'Hello World')


def test_rio_svc_endpoint1(nspc):
    image = "ibuildthecloud/demo:v1"
    image2 = "ibuildthecloud/demo:v3"

    srv = create_service(nspc, image)
    fullName = (f"{nspc}/{srv}")

    util.wait_for_app(fullName)
    stage_service(image2, fullName, "v3")

    svcEndpoint = get_version_endpoint(fullName, "v0")
    svcEndpoint2 = get_version_endpoint(fullName, "v3")
    print(svcEndpoint)
    print(svcEndpoint2)

    util.assert_endpoint(svcEndpoint, 'Hello World')
    util.assert_endpoint(svcEndpoint2, 'Hello World v3')
