import util
import os


def riovolume(stack, path):
    cmd = (f'rio up {stack} {path}')

    util.run(cmd)


def rio_check_bound(stack, vname):
    fullVolName = (f"{stack}/{vname}")
    state = 'bound'

    util.wait_for_state(fullVolName, state)
    inspect = util.rioInspect(fullVolName)

    return inspect['state']


def rio_bind_workload(stack, vname, wrklname):
    fullVolName = (f"{stack}/{vname}")
    fullWklName = (f"{stack}/{wrklname}")

    util.wait_for_state(fullVolName, "active")
    util.run(f"rio exec {fullVolName} touch /persistentvolumes/helloworld")
    util.run(f"rio run -n {fullWklName} -v data-{vname}-0:/data nginx")
    util.run(f"rio wait {fullWklName}")
    output = util.run(f"rio exec {fullWklName} ls /data")

    print(f'OUTPUT = {output}')

    return output


def test_vol_template(stack):
    if os.environ['RUN_NFS_TEST'] != "true":
        print('WARNING! nfs_not_enabled all test will fail')

    riovolume(stack, './nfs-stack/volume-template-stack.yaml')
    # to run locally you have to change url to
    # ./tests/nfs-stack/volume-template-stack.yaml

    cmd = (f"rio inspect {stack}/data --format json | jq '.template'")
    template_results = util.run(cmd)

    assert template_results == "true"

    assert rio_check_bound(stack, 'data-test1-0') == 'bound'
    assert rio_check_bound(stack, 'data-test2-0') == 'bound'

    assert rio_bind_workload(stack, 'test1', 'inspect-v1') == 'helloworld'
    assert rio_bind_workload(stack, 'test2', 'inspect-v2') == 'helloworld'
