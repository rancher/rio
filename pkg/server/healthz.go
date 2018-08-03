package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"k8s.io/apiserver/pkg/server/healthz"
)

type healthMux mux.Router

func (h *healthMux) Handle(pattern string, handler http.Handler) {
	(*mux.Router)(h).Handle(pattern, handler)
}

func installHealth(router *mux.Router) {
	healthz.InstallHandler((*healthMux)(router), healthz.PingHealthz)
}
