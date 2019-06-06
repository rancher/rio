import pytest
import os
import random
import time


@pytest.fixture(scope="module")
def nspc():
    nsp = "nsp" + str(random.randint(1000000, 9999999))

    yield nsp

    os.system(f"kubectl delete namespaces {nsp}")


@pytest.fixture
def service(nspc):
    srv = "tsrv" + str(random.randint(1000, 5000))
    fullName = (f"{nspc}/{srv}")

    os.system(f"rio run -n {fullName} nginx")
    time.sleep(5)
    print(f"{fullName}")

#    os.system(f"rio wait {fullName}")

    yield fullName

    os.system(f"rio rm {fullName}")
