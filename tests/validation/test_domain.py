# Run Validation test.  Use functions to test run and get output

import util


def create_service(nspc, image):
    srv = util.rioRun(nspc, image)
    fullName = (f"{nspc}/{srv}")

    return fullName


def create_domain(dname, fname):

    cmd = (f"rio domain register {dname} {fname}")
    print(cmd)
    util.run(cmd)

    return


def test_rio_domaintest(nspc):

    domainName = "foo.bar"
    image = "nginx"

    fullName = create_service(nspc, image)
    create_domain(domainName, fullName)
    dName = (f'{nspc}/foo-bar')

    result1 = util.rioInspect(dName, "spec.domainName")

    assert result1 == domainName
