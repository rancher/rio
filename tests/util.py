import subprocess
import json
import time
import random
import hashlib
import tempfile
from os import unlink

# runAndExpect("rio asldfkhsldkfj", 3)
# output = runAndExpect("rio ps")


def run(cmd, status=0):
    # @TODO actually check status
    result = subprocess.check_output(cmd, shell=True)
    result = result.decode("utf-8").strip()

    return result


def runwait(cmd, fullName):
    result = run(cmd)
    wait = "rio wait " + fullName
    result = run(wait)

    return result


def runToJson(cmd, status=0):
    print(cmd) 
    result = run(cmd, status)
    print(result)
    return json.loads(result)


def rioRun(stack, *args):
    name = "tsrv" + str(random.randint(1000, 5000))
    fullName = "%s/%s" % (stack, name)

    cmd = f'rio --wait --wait-timeout=60 run -n {fullName} ' + ' '.join(args)

    print(cmd)
    run(cmd)

    return name


def rioConfigCreate(stack, *configs):
    config_name = "tconfig" + str(random.randint(1000, 5000))

    fp = tempfile.NamedTemporaryFile(delete=False)

    for c in configs:
        fp.write(bytes(c + " ", 'utf8'))

    fp.close()

    run(f"rio config create {stack}/{config_name} {fp.name}")
    unlink(fp.name)

    return config_name


def rioInspect(resource, field=None):
    if field:
        return run("rio inspect --format '{{.%s}}' %s" % (field, resource))
    else:
        return runToJson("rio inspect %s" % resource)


def kubectl(namespace, type, resource):
    cmd = "rio kubectl get -n rio-cloud -o=json %s/%s-%s" % \
          (type, resource, namespace.split("-")[1])
    return runToJson(cmd)


def wait_for_state(name, exp_state):
    results = rioInspect(name)

    for i in range(1, 20):
        current_state = results['state']

        if current_state == exp_state:
            break

        time.sleep(1)


def hash_if_need(name, stack_name, project_name):
    sha = hashlib.sha256()
    sha.update("{}:{}".format(project_name, stack_name).encode('utf-8'))
    return "{}-{}".format(name, sha.hexdigest()[:8])
