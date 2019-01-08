# Run Validation test.  Use functions to test run and get output

from random import randint
import util
import os


def riovolume(stack, size):
    vname = "tvol" + str(randint(1000, 5000))
    fullName = (f"{stack}/{vname}")
    print(fullName)

    cmd = (f'rio volume create -d nfs {fullName} {size}')

    util.run(cmd)

    return vname


def rio_bound(stack, vname):
    fullVolName = (f"{stack}/{vname}")
    state = 'bound'

    util.wait_for_state(fullVolName, state)
    inspect = util.rioInspect(fullVolName)

    return inspect['state']


def rio_bind_workload(stack, vname):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = (f"{stack}/{name}")

    cmd = (f'rio run -n {fullName} -v {vname}:/data nginx')
    util.run(cmd)
    util.run(f"rio wait {fullName}")
    util.run(f"rio exec {fullName} touch /data/helloworld")
    output = util.run(f"rio exec {fullName} ls /data")

    return output


def riotest(stack, vname):
    fullVolName = (f"{stack}/{vname}")

    inspect = util.rioInspect(fullVolName)

    return inspect['sizeInGb']


def kubetest(stack, vname):
    fullVolName = (f"{stack}/{vname}")

    id = util.rioInspect(fullVolName, "id")
    namespace = id.split(":")[0]
    print(namespace)

    obj = util.kubectl(namespace, "pvc", vname)
    volsize = obj['spec']['resources']['requests']['storage']

    return volsize


def test_nfs_vol_bound(stack):
    volname = riovolume(stack, 1)
    if os.environ['RUN_NFS_TEST'] != "true":
        print('WARNING! nfs_not_enabled all test will fail')

    assert rio_bound(stack, volname) == 'bound'
    assert riotest(stack, volname) == 1
    assert kubetest(stack, volname) == '1Gi'


def test_nfs_bind_workload(stack):
    volname = riovolume(stack, 1)
    if os.environ['RUN_NFS_TEST'] != "true":
        print('WARNING! nfs_not_enabled all test will fail')

    rio_bound(stack, volname)
    results = rio_bind_workload(stack, volname)

    assert results == 'helloworld'
