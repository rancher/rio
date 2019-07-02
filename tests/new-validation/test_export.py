# Setup
import util


def pullimage(nspc, service):
    cmd = (f"rio export -t service {service} | grep 'image:' | cut -d' ' -f 4")
    image = util.run(cmd)
    return image


def pullscale(nspc, service):
    cmd = (f"rio export -t service {service} | grep 'scale:' | cut -d' ' -f 4")
    image = util.run(cmd)
    return image


def test_service_export_image(nspc, service):
    results = pullimage(nspc, service)
    assert results == "nginx"


def test_service_export_scale(nspc, service):
    results = pullscale(nspc, service)
    assert results == '1'

