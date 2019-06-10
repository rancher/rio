from random import randint
import util


def rio_permission_setup(stack, *rpermission):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = "%s/%s" % (stack, name)

    cmd = (f'rio run -n {fullName}')

    for p in rpermission:
        cmd += " --global-permission " + p

    cmd += " nginx"
    util.runwait(cmd, fullName)
    print(cmd)

    return name


def riotest(stack, sname):

    fullName = (f"{stack}/{sname}")
    inspect = util.rioInspect(fullName)

    return inspect['globalPermissions'][0]['verbs']


def kubesatest(stack, sname):

    fullName = (f"{stack}/{sname}")
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]

    obj = util.kubectl(namespace, "sa", sname)
    replicas = obj['metadata']['name']

    return replicas


def kubeclusterrolebindingtest(stack, sname):

    fullName = (f"{stack}/{sname}")
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0].split("-")[1]

    cmd = (f'rio kubectl get -o=json clusterrolebinding')
    obj = util.runToJson(cmd)

    for item in obj['items']:
        if item['metadata']['name'] == (f'{sname}-{namespace}'):
            return item['subjects'][0]['name']

    return None


def kubeclusterroletest(stack, sname):

    fullName = (f"{stack}/{sname}")
    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0].split("-")[1]

    cmd = (f'rio kubectl get -o=json clusterrole')
    obj = util.runToJson(cmd)

    for item in obj['items']:
        if item['metadata']['name'] == (f'{sname}-{namespace}'):
            return item['rules'][0]['verbs']

    return None


def test_name1(stack):
    serviceName = rio_permission_setup(stack, "'update cluster'")

    assert riotest(stack, serviceName) == ['update']

    assert serviceName in kubesatest(stack, serviceName)
    assert kubeclusterroletest(stack, serviceName) == ['update']
    assert serviceName in kubeclusterrolebindingtest(stack, serviceName)
