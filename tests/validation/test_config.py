# Setup
import util


def config_setup(nspc, *text):

    config = util.rioConfigCreate(nspc, *text)
    return config


def rio_config_content(fullname):
    rio_cont = (f"rio inspect --format json {fullname} | jq -r .data.content")
    rio_cont = util.run(rio_cont)

    return rio_cont

# Validation tests


def test_create_config1(nspc):
    text = "foo=bar"
    config_name = config_setup(nspc, text)
    fullname = (f"{nspc}/{config_name}")

    rio_content = rio_config_content(fullname)
    assert rio_content == "foo=bar"


def test_create_config2(nspc):
    text1 = "foo=bar"
    text2 = "foo2=bar2"
    config_name = config_setup(nspc, text1, text2)
    fullname = (f"{nspc}/{config_name}")

    rio_content = rio_config_content(fullname)
    assert rio_content == "foo=bar foo2=bar2"
