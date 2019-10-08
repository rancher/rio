// Copyright 2018 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package server

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/gogo/protobuf/jsonpb"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/log"
)

// HTTPGateway is a custom implementation of [gRPC gateway](https://github.com/grpc-ecosystem/grpc-gateway)
// specialized to Envoy xDS API.
type HTTPGateway struct {
	// Log is an optional log for errors in response write
	Log log.Logger

	// Server is the underlying gRPC server
	Server Server

	UrlToType map[string]string
}

func NewHTTPGateway(log log.Logger, srv Server, urlToType ...map[string]string) *HTTPGateway {
	types := make(map[string]string)
	for _, t := range urlToType {
		for k, v := range t {
			types[k] = v
		}
	}
	return &HTTPGateway{Log: log, Server: srv, UrlToType: types}
}

func (h *HTTPGateway) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	p := path.Clean(req.URL.Path)

	typeURL, ok := h.UrlToType[p]
	if !ok {
		http.Error(resp, "no endpoint", http.StatusNotFound)
		return
	}

	if req.Body == nil {
		http.Error(resp, "nil body", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(resp, "cannot read body", http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(resp, "empty body", http.StatusBadRequest)
		return
	}

	// parse as JSON
	out := &v2.DiscoveryRequest{}
	err = jsonpb.UnmarshalString(string(body), out)
	if err != nil {
		http.Error(resp, "cannot parse JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}
	out.TypeUrl = typeURL

	// fetch results
	res, err := h.Server.Fetch(req.Context(), out)
	if err != nil {
		// Note that this is treated as internal error. We may want to use another code for
		// the latest version fetch request.
		http.Error(resp, "fetch error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	buf := &bytes.Buffer{}
	if err := (&jsonpb.Marshaler{OrigName: true}).Marshal(buf, res); err != nil {
		http.Error(resp, "marshal error: "+err.Error(), http.StatusInternalServerError)
	}

	if _, err = resp.Write(buf.Bytes()); err != nil && h.Log != nil {
		h.Log.Errorf("gateway error: %v", err)
	}
}
