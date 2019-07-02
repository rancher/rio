# Run Validation test.  Use functions to test run and get output

import util


def create_external_service(nspc, extsvc, dnsrec):

    cmd = (f"rio externalservice create {nspc}/{extsvc} {dnsrec}")
    util.run(cmd)

    return


def test_rio_externalservice1(nspc):

    extsvcName = "extsvc1"
    dnsrec = "1.1.1.1"

    create_external_service(nspc, extsvcName, dnsrec)
    fullName = (f"{nspc}/{extsvcName}")

    result = util.rioInspect(fullName, "spec.ipAddresses[0]")

    assert result == dnsrec


def test_rio_externalservice2(nspc):

    extsvcName = "extsvc2"
    dnsrec = "my.domain.com"

    create_external_service(nspc, extsvcName, dnsrec)
    fullName = (f"{nspc}/{extsvcName}")

    result = util.rioInspect(fullName, "spec.fqdn")

    assert result == dnsrec
