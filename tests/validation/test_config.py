# Setup
import util


def config_setup(stack, *text):

    config = util.rioConfigCreate(stack, *text)
    fullname = (f"{stack}/{config}")
    return config


def rio_config_content(fullname):
    rio_content = "rio inspect --format '{{.content}}' %s" % fullname
    rio_content = util.run(rio_content)

    return rio_content


def kube_config_content(fullname, config_name):

    nsp = "rio inspect --format '{{.id}}' %s | cut -f1 -d:" % fullname
    nsp = util.run(nsp)
    k_cmd = "rio kubectl get config -n %s -o=json %s" % (nsp, config_name)
    kube_content = k_cmd + "| jq -r .spec.content"
    kube_content = util.run(kube_content)

    return kube_content

# Validation tests


def test_create_config1(stack):
    text = "foo=bar"
    config_name = config_setup(stack, text)
    fullname = (f"{stack}/{config_name}")

    rio_content = rio_config_content(fullname)
    assert rio_content == "foo=bar"

    kube_content = kube_config_content(fullname, config_name)
    assert kube_content == "foo=bar"


def test_create_config2(stack):
    text1 = "foo=bar"
    text2 = "foo2=bar2"
    config_name = config_setup(stack, text1, text2)
    fullname = (f"{stack}/{config_name}")

    rio_content = rio_config_content(fullname)
    assert rio_content == "foo=bar foo2=bar2"

    kube_content = kube_config_content(fullname, config_name)
    assert kube_content == "foo=bar foo2=bar2"
