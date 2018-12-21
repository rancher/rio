# Run Validation test.  Use functions to test run and get outpu
import util
from random import randint


def riorun(stack):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = "%s/%s" % (stack, name)

    cmd = (f'rio run -p 80/http -n {fullName} nginx:latest')

    util.run(cmd)
    util.run(f"rio wait {fullName}")

    return name


def riocreatedomain(stack, s_name):
    cmd = (f'rio domain add test.foo.bar2 {stack}/{s_name}')
    util.run(cmd)


def riotest():
    domainId = "rio domain --format '{{.ID}}'"
    domainId = util.run(domainId)

    inspect = util.rioInspect(domainId)

    return inspect['domainName']


def kubetest(stack, service):

    fullName = "%s/%s" % (stack, service)
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    contents = util.kubectl(namespace, 'virtualservice', service)
    print(contents)

    domainName = contents['spec']['hosts'][2]
    print(domainName)

    return domainName


def test_domain_name(stack):
    sname = riorun(stack)
    riocreatedomain(stack, sname)
    riodname = riotest()
    assert riodname == "test.foo.bar2"

    k8sdname = kubetest(stack, sname)
    assert k8sdname == "test.foo.bar2"

    cmd = (f'rio domain rm test.foo.bar2 ')
    util.run(cmd)
