import subprocess
import json
import time
import random
import hashlib


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


def wait_for_app(full_name):
    for i in range(1, 120):
        cmd = f"rio inspect --type app {full_name}"
        try:
            run(cmd)
            break
        except subprocess.CalledProcessError:
            print("waiting")
        time.sleep(1)


def wait_for_service(full_name):
    for i in range(1, 120):
        cmd = f"rio inspect --type service {full_name}"
        try:
            run(cmd)
            break
        except subprocess.CalledProcessError:
            print("waiting")
        time.sleep(1)


def rioRun(nspc, *args):
    srv = "tsrv" + str(random.randint(1000000, 9999999))
    fullName = (f"{nspc}/{srv}")
    print(f"{fullName}")

    cmd = (f"rio run -n {fullName} " + ' '.join(args))
    print(cmd)

    run(cmd)

    return srv


def rioStage(image, srv, version):
    cmd = (f"rio stage --image={image} {srv}:{version}")
    run(cmd)

    return


def assert_endpoint(endpoint, result):
    output = ""
    for i in range(1, 120):
        try:
            output = run(f"curl -s {endpoint}")
        except subprocess.CalledProcessError:
            pass
        if output == result:
            break
        time.sleep(1)


def rioInspect(resource, field=None):
    if field:
        print(resource)
        print(field)
        return run(f"rio inspect --format json {resource} | jq -r .{field}")
    else:
        return runToJson(f"rio inspect {resource}")


def kubectl(namespace, ktype, resource):
    cmd = (f"kubectl get -n {namespace} -o=json {ktype}/{resource}")
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
