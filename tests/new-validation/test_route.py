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


def route_service(nspc, rName, sname, fName, vs):

    cmd = (f"rio route append {rName}.{nspc}/to-{sname}-{vs} to {fName}:{vs}")
    print(f"{cmd}")
    util.run(cmd)

    return


def test_rio_svc_route(nspc):
    image = "ibuildthecloud/demo:v1"
    image2 = "ibuildthecloud/demo:v3"
    rName = "route1"

    fullName = create_service(nspc, image)
    stage_service(image2, fullName, "v3")

    sName = get_app_info(fullName, "metadata.name")

    route_service(nspc, rName, sName, fullName, "v0")
    route_service(nspc, rName, sName, fullName, "v3")

    sName = get_app_info(fullName, "metadata.name")
    print(f"{sName}")
    nsproute = (f"{nspc}/{rName}")
    result = get_app_info(nsproute, "status.endpoint[0]")
    print(f"{result}")

    time.sleep(2)

    cmd = (f"curl -s {result}/to-{sName}-v0")
    print(f"{cmd}")
    cmd2 = (f"curl -s {result}/to-{sName}-v3")
    print(f"{cmd2}")

    results1 = util.run(cmd)
    results2 = util.run(cmd2)
    print(f"{results1}")

    assert results1 == 'Hello World'
    assert results2 == 'Hello World v3'
