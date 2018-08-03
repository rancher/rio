package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/kubernetes/cmd/server"
)

func doAuth(serverConfig *server.ServerConfig, next http.Handler, rw http.ResponseWriter, req *http.Request) {
	if serverConfig.Authenticator == nil {
		next.ServeHTTP(rw, req)
		return
	}

	user, ok, err := serverConfig.Authenticator.AuthenticateRequest(req)
	if err != nil {
		logrus.Errorf("failed to authenticate request: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !ok {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	ctx := request.WithUser(req.Context(), user)
	req = req.WithContext(ctx)
	next.ServeHTTP(rw, req)
}

func authMiddleware(serverConfig *server.ServerConfig) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			doAuth(serverConfig, next, rw, req)
		})
	}
}
