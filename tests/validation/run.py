""" Run Validation test.  Use functions to test run and get output. """
import pytest
import util

def riotest(stack, service):
    rio_com = "rio inspect --format '{{.state}}' %s/%s" % (stack, service)
    results = util.run(rio_com)

    return results

def kubetest(stack, service):
    n_com = "rio inspect --format '{{.id}}' %s/%s | cut -f1 -d:" % (stack, service)
    nsp = util.run(n_com)
    kube_com = "rio kubectl get -n %s -o=json deploy/%s | jq -r .status.replicas" % (nsp, service)
    replicas = util.run(kube_com)

    return replicas

def test_rio_status(stack, service):
    assert riotest(stack, service) == "active"

def test_kube_replicas(stack, service):
    assert kubetest(stack, service) == "1"
