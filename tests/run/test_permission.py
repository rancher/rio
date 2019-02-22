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
    realname = sname + "-" + id.split(":")[0].split("-")[1]

    cmd = (f'rio kubectl get -n rio-cloud -o=json rolebinding {realname}')
    obj = util.runToJson(cmd)
    replicas = obj['roleRef']['name']

    return replicas


def kuberoletest(stack, sname):

    fullName = (f"{stack}/{sname}")
    id = util.rioInspect(fullName, "id")
    realname = sname + "-" + id.split(":")[0].split("-")[1]

    cmd = (f'rio kubectl get -n rio-cloud -o=json role {realname}')
    obj = util.runToJson(cmd)
    replicas = obj['rules'][0]['verbs']

    return replicas


def test_name1(stack):
    serviceName = rio_permission_setup(stack, "'update service'")

    assert riotest(stack, serviceName) == ['update']

    assert serviceName in kubesatest(stack, serviceName)
    assert kuberoletest(stack, serviceName) == ['update']
    assert serviceName in kuberolebindingtest(stack, serviceName)
