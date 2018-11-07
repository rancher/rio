# Setup
from random import randint
import util


def create_service(stack):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = "%s/%s" % (stack, name)
    rio_create = "rio create -n %s nginx" % fullName
    rio_create = util.run(rio_create)
    util.run("rio --wait-state inactive wait %s" % fullName)

    return fullName


def test_create_status(stack):
    fullName = create_service(stack)
    create_status = util.rioInspect(fullName, "state")

    assert create_status == "inactive"


def test_create_scale(stack):
    fullName = create_service(stack)
    create_scale = util.rioInspect(fullName, "scale")

    assert create_scale == '0'
