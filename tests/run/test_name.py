
import util


def rio_name_setup(stack, sname):

    fullName = (f"{stack}/{sname}")
    util.run(f"rio run --name {fullName} nginx")
    util.run(f"rio wait {fullName}")

    return sname


def riotest(stack, sname):

    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    return inspect['name']


def kubetest(stack, sname):

    fullName = (f"{stack}/{sname}")
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "deployment", sname)
    replicas = obj['metadata']['name']

    return replicas


def test_name1(stack):
    serviceName = "testname"
    rio_name_setup(stack, serviceName)
    assert serviceName in riotest(stack, serviceName)

    assert serviceName in kubetest(stack, serviceName)


def test_name2(stack):
    serviceName = "a"
    rio_name_setup(stack, serviceName)

    assert serviceName in riotest(stack, serviceName)

    assert serviceName in kubetest(stack, serviceName)


def test_name3(stack):
    serviceName = "areallylongnamethatcanbelonger"
    rio_name_setup(stack, serviceName)

    assert serviceName in riotest(stack, serviceName)

    assert serviceName in kubetest(stack, serviceName)
