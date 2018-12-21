import util


def test_stack_create(stack):
    stack_status = util.rioInspect(stack, "state")
    assert stack_status == "active"


def test_add_service(stack, service):
    fullName = stack + "/" + service
    service_status = util.rioInspect(fullName, "state")
    assert service_status == "active"
