# Run Validation test.  Use functions to test run and get output

from random import randint
import util


def riovolume(size):
    vname = "tvol" + str(randint(1000, 5000))
    cmd = (f'rio volume create {vname} {size}')

    util.run(cmd)

    return vname


def riotest(vname):
    inspect = util.rioInspect(vname)

    return inspect['sizeInGb']


def kubetest(vname):

    id = util.rioInspect(vname, "id")
    nspace = id.split(":")[0]

    obj = util.kubectl(nspace, "pvc", vname)
    volsize = obj['spec']['resources']['requests']['storage']

    return volsize


def test_vol_size():
    volname = riovolume(10)

    assert riotest(volname) == 10

#    assert kubetest(volname) == '10Gi'

    cmd = (f'rio rm {volname}')
    util.run(cmd)
