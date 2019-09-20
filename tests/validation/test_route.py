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


def route_service(nspc, rName, sname, fName, vs):

    cmd = (f"rio route add {rName}.{nspc}/to-{sname}-{vs} to {fName}:{vs}")
    print(f"{cmd}")
    util.run(cmd)

    return


def test_rio_svc_route(nspc):
    image = "ibuildthecloud/demo:v1"
    image2 = "ibuildthecloud/demo:v3"
    rName = "route1"

    fullName = create_service(nspc, image)
    util.wait_for_app(fullName)
    stage_service(image2, fullName, "v3")

    sName = get_app_info(fullName, "metadata.name")

    route_service(nspc, rName, sName, fullName, "v0")
    route_service(nspc, rName, sName, fullName, "v3")

    sName = get_app_info(fullName, "metadata.name")
    print(f"{sName}")
    nsproute = (f"{nspc}/{rName}")
    result = ""
    for i in range(1, 120):
        result = get_app_info(nsproute, "status.endpoints[0]")
        if result != "null":
            break
        time.sleep(1)
    print(f"{result}")

    util.assert_endpoint(f"{result}/to-{sName}-v0", 'Hello World')
    util.assert_endpoint(f"{result}/to-{sName}-v3", 'Hello World v3')
