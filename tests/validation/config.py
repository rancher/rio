# Setup
import pytest
from os import unlink
from random import randint
import util
import tempfile


@pytest.fixture(scope="module")
def config_setup():
    config = "tconfig" + str(randint(1000, 5000))

    fp = tempfile.NamedTemporaryFile(delete=False)
    fp.write(b"foo=bar")
    fp.close()

    util.run("rio config create %s %s" % (config, fp.name))

    yield config

    util.run("rio config rm %s" % config)
    unlink(fp.name)


# Validation tests


def test_rio_content(config_setup):
    print(config_setup)
    rio_content = "rio inspect --format '{{.content}}' %s" % config_setup
    rio_content = util.run(rio_content)

    assert rio_content == "foo=bar"


def test_kube_content(config_setup):
    print(config_setup)
    nsp = "rio inspect --format '{{.id}}' %s | cut -f1 -d:" % config_setup
    nsp = util.run(nsp)
    k_cmd = "rio kubectl get config -n %s -o=json %s" % (nsp, config_setup)
    kube_content = k_cmd + "| jq -r .spec.content"
    kube_content = util.run(kube_content)

    assert kube_content == "foo=bar"
