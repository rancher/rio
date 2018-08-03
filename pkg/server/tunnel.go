package server

import (
	"net/http"

	"github.com/rancher/rancher/pkg/remotedialer"
	"k8s.io/apiserver/pkg/endpoints/request"
)

func authorizer(req *http.Request) (clientKey string, authed bool, err error) {
	user, ok := request.UserFrom(req.Context())
	if !ok {
		return "", false, nil
	}

	if user.GetName() != "node" {
		return "", false, nil
	}

	nodeName := req.Header.Get("X-Rio-NodeName")
	if nodeName == "" {
		return "", false, nil
	}

	return nodeName, true, nil
}

func newTunnel() http.Handler {
	server := remotedialer.New(authorizer, remotedialer.DefaultErrorWriter, remotedialer.AlwaysReady)
	setupK3s(server)
	return server
}
