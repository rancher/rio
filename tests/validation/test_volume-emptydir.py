# Run Validation test.  Use functions to test run and get outpu

import util
from random import randint


def riorun(stack, vname):
    sname = "tsrv" + str(randint(1000, 5000))
    fullName = "%s/%s" % (stack, sname)

    cmd = (f'rio run -n {fullName} -v {vname} nginx')
    util.run(cmd)

    util.run(f"rio wait {fullName}")

    return sname


def rio_vol_target(stack, sname):
    fullName = (f"{stack}/{sname}")

    inspect = util.rioInspect(fullName)

    return inspect['volumes'][0]['target']


def kube_vol_target(stack, sname):
    fullName = (f"{stack}/{sname}")

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", sname)
    obj = obj['spec']['template']['spec']['containers'][0]

    return obj['volumeMounts'][0]['mountPath']


def kube_vol_name(stack, sname):
    fullName = (f"{stack}/{sname}")

    id = util.rioInspect(fullName, "id")
    namespace = id.split(":")[0]
    obj = util.kubectl(namespace, "deployment", sname)
    obj = obj['spec']['template']['spec']['containers'][0]

    return obj['volumeMounts'][0]['name']


def test_anon(stack):
    srv_name = riorun(stack, '/data')

    target = rio_vol_target(stack, srv_name)
    assert target == '/data'

    mpath = kube_vol_target(stack, srv_name)
    assert mpath == '/data'

    vol_name = kube_vol_name(stack, srv_name)
    assert vol_name == 'anon-data'


def test_named(stack):
    srv_name = riorun(stack, 'test:/data')

    target = rio_vol_target(stack, srv_name)
    assert target == '/data'

    mpath = kube_vol_target(stack, srv_name)
    assert mpath == '/data'

    vol_name = kube_vol_name(stack, srv_name)
    assert vol_name == 'test'
