import pytest
import os
import random


random.seed(os.urandom(8))


@pytest.fixture(scope="module")
def stack():
    name = "tstk" + str(random.randint(1000, 5000))
    os.system("rio stack create %s" % name)
    os.system("rio wait %s" % name)

    yield name

    os.system("rio stack rm %s" % name)


@pytest.fixture
def service(stack):
    name = "tsrv" + str(random.randint(1000, 5000))
    fullName = "%s/%s" % (stack, name)

    os.system("rio run -n %s nginx" % fullName)
    os.system("rio wait %s" % fullName)

    yield name

    os.system("rio rm %s" % fullName)
