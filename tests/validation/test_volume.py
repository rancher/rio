# Run Validation test.  Use functions to test run and get output

from random import randint
import util


def riovolume(stack, size):
    vname = "tvol" + str(randint(1000, 5000))
    fullName = (f"{stack}/{vname}")

    cmd = (f'rio volume create {fullName} {size}')

    util.run(cmd)

    return fullName


def riotest(vname):
    inspect = util.rioInspect(vname)

    return inspect['sizeInGb']


def kubetest(vname):

    id = util.rioInspect(vname, "id")
    nspace = id.split(":")[0]

    obj = util.kubectl(nspace, "pvc", vname)
    volsize = obj['spec']['resources']['requests']['storage']

    return volsize


def test_vol_size(stack):
    volname = riovolume(stack, 1)

    assert riotest(volname) == 1

#    assert kubetest(volname) == '1Gi'
