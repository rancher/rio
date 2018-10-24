import pytest
from os import system
from random import randint


@pytest.fixture(scope="module")
def stack():
    name = "tstk" + str(randint(1000, 5000))
    system("rio stack create %s" % name)
    system("rio wait %s" % name)

    yield name

    system("rio stack rm %s" % name)


@pytest.fixture
def service(stack):
    name = "tsrv" + str(randint(1000, 5000))
    fullName = "%s/%s" % (stack, name)

    system("rio run -n %s nginx" % fullName)
    system("rio wait %s" % fullName)

    yield name

    system("rio rm %s" % fullName)
