package wrapper

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

/*
A wrapper.Client wraps a ResourceClient, applying a
Processing function to each read and written resource
*/
type Client struct {
	clients.ResourceClient
	ProcessResource func(resource resources.Resource)
}

func (c *Client) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	res, err := c.ResourceClient.Read(namespace, name, opts)
	if err != nil {
		return nil, err
	}
	c.ProcessResource(res)
	return res, nil
}

func (c *Client) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	c.ProcessResource(resource)
	return c.ResourceClient.Write(resource, opts)
}

func (c *Client) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	list, err := c.ResourceClient.List(namespace, opts)
	if err != nil {
		return nil, err
	}
	list.Each(c.ProcessResource)
	return list, nil
}

func (c *Client) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	opts = opts.WithDefaults()
	resourceLists, errs, err := c.ResourceClient.Watch(namespace, opts)
	if err != nil {
		return nil, nil, err
	}

	out := make(chan resources.ResourceList)
	go func() {
		defer close(out)
		for list := range resourceLists {
			list.Each(c.ProcessResource)
			// Send the data to the output channel but return early
			// if the context has been cancelled.
			select {
			case out <- list:
			case <-opts.Ctx.Done():
				return
			}
		}
	}()
	return out, errs, nil

}
