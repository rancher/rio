package server

import "net/http"

type prepareRt struct {
	rt          http.RoundTripper
	prepareFunc func(*http.Request)
}

func newCallback(rt http.RoundTripper, prepare func(*http.Request)) *prepareRt {
	return &prepareRt{
		rt:          rt,
		prepareFunc: prepare,
	}
}

func (p *prepareRt) RoundTrip(req *http.Request) (*http.Response, error) {
	p.prepareFunc(req)
	return p.rt.RoundTrip(req)
}
