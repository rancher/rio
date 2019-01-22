import pytest
import os
import random


@pytest.fixture(scope="module")
def stack():
    name = "tstk" + str(random.randint(1000000, 9999999))
    os.system("rio --wait --wait-timeout=60 stack create %s" % name)

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
