# Setup
import pytest
import util


@pytest.fixture
def serviceJson(stack, service):
    fullName = "%s/%s" % (stack, service)
    obj = util.runToJson("rio export -o json -t service %s" % fullName)
    return obj


@pytest.fixture
def stackJson(stack, service):
    # service is passed in so that there will be one in the stack export
    obj = util.runToJson("rio export -o json %s" % stack)
    return obj


def test_service_export_image(serviceJson):
    assert serviceJson['image'] == "nginx"


def test_service_export_name(service, serviceJson):
    assert serviceJson['name'] == ("%s" % service)


def test_service_export_scale(serviceJson):
    assert serviceJson['scale'] == 1


def test_stack_export_service_image(stackJson, service):
    assert stackJson['services'][service]['image'] == "nginx"


def test_stack_export_service_scale(stackJson, service):
    assert stackJson['services'][service]['scale'] == 1
