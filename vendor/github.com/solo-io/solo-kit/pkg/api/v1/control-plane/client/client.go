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

package client

import (
	"context"

	google_rpc "github.com/gogo/googleapis/google/rpc"

	"github.com/gogo/protobuf/proto"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"

	"google.golang.org/grpc"
)

type TypeRecord interface {
	Type() string
	EmptyProto() cache.ResourceProto
	ProtoToResource(r cache.ResourceProto) cache.Resource
}

type typeRecord struct {
	rtype           string
	proto           func() cache.ResourceProto
	protoToResource func(r cache.ResourceProto) cache.Resource
}

func NewTypeRecord(
	rtype string,
	proto func() cache.ResourceProto,
	protoToResource func(r cache.ResourceProto) cache.Resource,
) TypeRecord {
	return &typeRecord{
		rtype:           rtype,
		proto:           proto,
		protoToResource: protoToResource,
	}
}

func (t *typeRecord) Type() string {
	return t.rtype
}
func (t *typeRecord) EmptyProto() cache.ResourceProto {
	return t.proto()
}
func (t *typeRecord) ProtoToResource(r cache.ResourceProto) cache.Resource {
	return t.protoToResource(r)
}

type Client interface {
	Start(ctx context.Context, cc *grpc.ClientConn) error
}

type client struct {
	nodeinfo *core.Node
	rtype    TypeRecord
	apply    func(cache.Resources) error
}

func NewClient(nodeinfo *core.Node, rtype TypeRecord, apply func(cache.Resources) error) Client {
	return &client{
		nodeinfo: nodeinfo,
		rtype:    rtype,
		apply:    apply,
	}

}

/**
 * Start a client. this function is blocking.
 */

func (c *client) Start(ctx context.Context, cc *grpc.ClientConn) error {
	client := discovery.NewAggregatedDiscoveryServiceClient(cc)
	resourceclient, err := client.StreamAggregatedResources(ctx)
	if err != nil {
		return err
	}
	// get a request going
	dr := &v2.DiscoveryRequest{
		VersionInfo:   "",
		Node:          c.nodeinfo,
		ResourceNames: []string{},
		TypeUrl:       c.rtype.Type(),
		ResponseNonce: "",
		ErrorDetail:   nil,
	}
	for {
		// make a copy of dr, to guarantee it doesnt get modified
		tosend := *dr
		err = resourceclient.Send(&tosend)
		if err != nil {
			return err
		}
		resp, err := resourceclient.Recv()
		if err != nil {
			return err
		}

		dr.ResponseNonce = resp.Nonce

		var resources cache.Resources
		resources.Version = resp.VersionInfo
		resources.Items = make(map[string]cache.Resource)
		for _, r := range resp.Resources {
			into := c.rtype.EmptyProto()
			err = proto.Unmarshal(r.Value, into)
			if err != nil {
				break
			}
			resource := c.rtype.ProtoToResource(into)
			resources.Items[resource.Self().Name] = resource
		}
		// If we have an error, don't update version info to signal NACK.
		if err != nil {
			dr.ErrorDetail = &google_rpc.Status{
				Code:    int32(google_rpc.INVALID_ARGUMENT),
				Message: err.Error(),
			}
		} else if err = c.apply(resources); err != nil {
			dr.ErrorDetail = &google_rpc.Status{
				Code:    int32(google_rpc.UNKNOWN),
				Message: err.Error(),
			}
		} else {
			dr.VersionInfo = resp.VersionInfo
			dr.ErrorDetail = nil
		}
	}

}
