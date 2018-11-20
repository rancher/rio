import util
from os import system
from random import randint
import tempfile


def config_version(stack, version):
    config = "tconfig" + str(randint(1000, 5000))

    fp = tempfile.NamedTemporaryFile(delete=False)
    version_as_bytes = str.encode(version)
    fp.write(b'%s' % version_as_bytes)
    fp.close()
    print(stack)
    config_name = stack + "/" + config
    print(config_name)

    util.run("rio config create %s %s" % (config_name, fp.name))

    return config


def create_service(stack, config):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = "%s/%s" % (stack, name)
    path = "/usr/share/nginx/html/index.html"
    print(fullName)
    print(config)
    run_command = "rio run -n %s -p 80/http --config %s:%s nginx" % (
        fullName, config, path
        )
    print(run_command)
    system(run_command)
    system("rio wait %s" % fullName)

    return name


def stage_service(stack, name, version, second_config):
    fullName = "%s/%s" % (stack, name)
    path = "/usr/share/nginx/html/index.html"
    command = "rio stage --image=nginx --config %s:%s %s:%s" % (
        second_config, path, fullName, version
    )
    print(command)
    system(command)
    system("rio wait %s" % fullName)
    stackJson = util.runToJson("rio export -o json %s" % stack)
    got = stackJson['services'][name]['revisions']['v2']['scale']

    return got


def weight_service(stack, name, version, weight):
    fullName = "%s/%s" % (stack, name)
    command = "rio weight %s:%s=%s" % (fullName, version, weight)
    system(command)
    stackJson = util.runToJson("rio export -o json %s" % stack)
    got = stackJson['services'][name]['revisions']['v2']['weight']

    return got


def promote_service(stack, name, version):
    fullName = "%s/%s" % (stack, name)
    command = "rio promote %s:%s" % (fullName, version)
    system(command)
    stackJson = util.runToJson("rio export -o json %s" % stack)
    got = stackJson['services'][name]['version']

    return got


def test_stage_service(stack):
    config_name = config_version(stack, "1")
    name = create_service(stack, config_name)
    second_config = config_version(stack, "2")
    got = stage_service(stack, name, "v2", second_config)
    assert got == 1


def test_weight_service(stack):
    config_name = config_version(stack, "1")
    name = create_service(stack, config_name)
    second_config = config_version(stack, "2")
    stage_service(stack, name, "v2", second_config)
    got = weight_service(stack, name, "v2", "50")
    assert got == 50


def test_promote_service(stack):
    config_name = config_version(stack, "1")
    name = create_service(stack, config_name)
    second_config = config_version(stack, "2")
    stage_service(stack, name, "v2", second_config)
    weight_service(stack, name, "v2", "50")
    got = promote_service(stack, name, "v2")
    assert got == "v2"
