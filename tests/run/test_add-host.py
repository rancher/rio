import util


def run_addhost(nspc, *value):

    name = util.rioRun(nspc, ' '.join(value), 'nginx')

    return name


def rio_chk_host(nspc, sname, arraynum, arraynum2):
    fullName = (f"{nspc}/{sname}:v0")
    print(fullName)
    hostname = (f"spec.hostAliases[{arraynum2}].hostnames[{arraynum}]")
    inspect = util.rioInspect(fullName, hostname)

    return inspect


def rio_chk_ip(nspc, sname, arraynum):
    fullName = (f"{nspc}/{sname}:v0")
    print(fullName)
    hostname = (f"spec.hostAliases[{arraynum}].ip")
    inspect = util.rioInspect(fullName, hostname)

    return inspect


def test_addhost1(nspc):
    service_name = run_addhost(nspc, '--add-host', 'example.com:1.2.3.4')
    arraynum = 0

    rio_hostname = rio_chk_host(nspc, service_name, arraynum, arraynum)
    assert rio_hostname == 'example.com'

    rio_hostip = rio_chk_ip(nspc, service_name, arraynum)
    assert rio_hostip == '1.2.3.4'


def test_addhost2(nspc):
    service_name = run_addhost(nspc, '--add-host', 'example.com:1.2.3.4',
                               '--add-host', 'bexample.com:2.3.4.5')
    arraynum1 = 0
    arraynum2 = 1

    rio_hostname = rio_chk_host(nspc, service_name, arraynum1, arraynum1)
    assert rio_hostname == 'example.com'

    rio_hostip = rio_chk_ip(nspc, service_name, arraynum1)
    assert rio_hostip == '1.2.3.4'

    rio_hostname = rio_chk_host(nspc, service_name, arraynum1, arraynum2)
    assert rio_hostname == 'bexample.com'

    rio_hostip = rio_chk_ip(nspc, service_name, arraynum2)
    assert rio_hostip == '2.3.4.5'
