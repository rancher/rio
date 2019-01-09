# Setup
from os import unlink
from random import randint
import util
import tempfile


def config_setup(stack):
    config = "tconfig" + str(randint(1000, 5000))

    fp = tempfile.NamedTemporaryFile(delete=False)
    fp.write(b"foo=bar")
    fp.close()

    fullname = (f"{stack}/{config}")

    util.run(f"rio config create {fullname} {fp.name}")
    unlink(fp.name)

    return config


# Validation tests


def test_rio_content(stack):
    config_name = config_setup(stack)
    fullname = (f"{stack}/{config_name}")
    print(config_name)
    rio_content = "rio inspect --format '{{.content}}' %s" % fullname
    rio_content = util.run(rio_content)

    assert rio_content == "foo=bar"


def test_kube_content(stack):
    config_name = config_setup(stack)
    fullname = (f"{stack}/{config_name}")

    nsp = "rio inspect --format '{{.id}}' %s | cut -f1 -d:" % fullname
    nsp = util.run(nsp)
    k_cmd = "rio kubectl get config -n %s -o=json %s" % (nsp, config_name)
    kube_content = k_cmd + "| jq -r .spec.content"
    kube_content = util.run(kube_content)

    assert kube_content == "foo=bar"
