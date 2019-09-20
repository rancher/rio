import pytest
import os
import random
import util
import subprocess
import time


@pytest.fixture(scope="module")
def nspc():
    nsp = "nsp" + str(random.randint(1000000, 9999999))

    yield nsp

    os.system(f"kubectl delete namespaces {nsp} &")


@pytest.fixture
def service(nspc):
    srv = "tsrv" + str(random.randint(1000000, 9999999))
    fullName = (f"{nspc}/{srv}")

    os.system(f"rio run -n {fullName} nginx")
    for i in range(1, 120):
        try:
            time.sleep(1)
            util.wait_for_app(fullName)
            break
        except subprocess.CalledProcessError:
            print("waiting")

    print(f"{fullName}")

    yield fullName

    os.system(f"rio rm {fullName}")
