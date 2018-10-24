import subprocess
import json

# runAndExpect("rio asldfkhsldkfj", 3)
# output = runAndExpect("rio ps")


def run(cmd, status=0):
    # @TODO actually check status
    result = subprocess.check_output(cmd, shell=True)
    result = result.decode("utf-8").strip()

    return result


def runToJson(cmd, status=0):
    print("COMMAND: %s" % cmd)
    result = run(cmd, status)
    print("JSON RESULT: %s" % result)

    return json.loads(result)


def rioInspect(resource, field):
    return run("rio inspect --format '{{.%s}}' %s" % (field, resource))


def kubectl(namespace, type, resource):
    cmd = "rio kubectl get -n %s -o=json %s/%s" % (namespace, type, resource)
    return runToJson(cmd)
