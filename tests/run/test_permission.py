from random import randint
import util


def rio_permission_setup(stack, *rpermission):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = "%s/%s" % (stack, name)

    cmd = (f'rio run -n {fullName}')

    for p in rpermission:
        cmd += " --permission " + p

    cmd += " nginx"
    util.run(cmd)
    print(cmd)
    util.run(f"rio wait {fullName}")

    return name


def riotest(stack, sname):

    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    return inspect['permissions'][0]['verbs']


def kubesatest(stack, sname):

    fullName = (f"{stack}/{sname}")
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "sa", sname)
    replicas = obj['metadata']['name']

    return replicas


def kuberolebindingtest(stack, sname):

    fullName = (f"{stack}/{sname}")
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    cmd = (f'rio kubectl get -n {namespace} -o=json rolebinding')
    obj = util.runToJson(cmd)
    replicas = obj['items'][0]['roleRef']['name']

    return replicas


def kuberoletest(stack, sname):

    fullName = (f"{stack}/{sname}")
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    cmd = (f'rio kubectl get -n {namespace} -o=json role')
    obj = util.runToJson(cmd)
    replicas = obj['items'][0]['rules'][0]['verbs']

    return replicas


def test_name1(stack):
    serviceName = rio_permission_setup(stack, "'update service'")

    assert riotest(stack, serviceName) == ['update']

    assert kubesatest(stack, serviceName) == serviceName
    assert kuberoletest(stack, serviceName) == ['update']
    assert kuberolebindingtest(stack, serviceName) == serviceName
