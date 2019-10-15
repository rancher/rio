package wrapper

import (
	"context"
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

// predefine the options for starting a watch
type ClientWatchOpts struct {
	// the watch client to start the watch
	BaseClient clients.ResourceWatcher
	// the namespace to watch
	// will be ignored if the resource
	// is cluster-scoped
	Namespace string

	// if provided, the watch will return lists
	// containing only the single named resource
	// or an empty list
	ResourceName string

	// the label selector to apply to the watch
	Selector map[string]string
}

func AggregatedWatchFromClients(clientOpts ...ClientWatchOpts) clients.ResourceWatch {
	var watches []clients.ResourceWatch
	for _, opt := range clientOpts {
		watch := ResourceWatch(opt.BaseClient, opt.Namespace, opt.Selector)
		if opt.ResourceName != "" {
			watch = SingleResourceWatch(watch, opt.ResourceName)
		}
		watches = append(watches, watch)
	}
	return AggregatedWatch(watches...)
}

func ResourceWatch(rw clients.ResourceWatcher, namespace string, selector map[string]string) clients.ResourceWatch {
	return func(ctx context.Context) (<-chan resources.ResourceList, <-chan error, error) {
		return rw.Watch(namespace, clients.WatchOpts{
			Ctx:      ctx,
			Selector: selector,
		})
	}
}

func SingleResourceWatch(rw clients.ResourceWatch, resourceName string) clients.ResourceWatch {
	return func(ctx context.Context) (<-chan resources.ResourceList, <-chan error, error) {
		unfiltered, errs, err := rw(ctx)
		if err != nil {
			return nil, nil, err
		}
		out := make(chan resources.ResourceList)
		go func() {
			defer close(out)
			for {
				select {
				case <-ctx.Done():
					return
				case list, ok := <-unfiltered:
					if !ok {
						return
					}
					var singleResourceList resources.ResourceList
					for _, resource := range list {
						if resource.GetMetadata().Name == resourceName {
							singleResourceList = append(singleResourceList, resource)
							break
						}
					}
					select {
					case <-ctx.Done():
						return
					case out <- singleResourceList:
					}
				}
			}
		}()
		return out, errs, nil
	}
}

func AggregatedWatch(watches ...clients.ResourceWatch) clients.ResourceWatch {
	return func(ctx context.Context) (<-chan resources.ResourceList, <-chan error, error) {
		listsByWatcher := newResourcesByWatchIndex()
		out := make(chan resources.ResourceList)
		aggregatedErrs := make(chan error)
		sourceWatches := sync.WaitGroup{}

		for i, w := range watches {
			sourceWatches.Add(1)
			key := i
			lists, errs, err := w(ctx)
			if err != nil {
				return nil, nil, err
			}
			go func() {
				defer sourceWatches.Done()
				defer listsByWatcher.delete(key)
				for {
					select {
					case <-ctx.Done():
						return
					case err := <-errs:
						select {
						case <-ctx.Done():
							return
						case aggregatedErrs <- err:
						}
						// if the source starts returning errors, remove its list from the snapshot
						select {
						case <-ctx.Done():
							return
						case out <- listsByWatcher.delete(key).merge():
						}
					case list, ok := <-lists:
						if !ok {
							return
						}
						// add/update the list to the snapshot
						select {
						case <-ctx.Done():
							return
						case out <- listsByWatcher.set(key, list).merge():
						}
					}
				}
			}()
		}

		go func() {
			// context is closed, clean up watch resources
			<-ctx.Done()
			// wait for source watches to be closed before closing the sink
			sourceWatches.Wait()
			close(out)
			close(aggregatedErrs)
		}()
		return out, aggregatedErrs, nil
	}
}

// aggregate resources by the channel they were read from
type resourcesByWatchIndex struct {
	access    sync.RWMutex
	resources map[int]resources.ResourceList
}

func newResourcesByWatchIndex() *resourcesByWatchIndex {
	return &resourcesByWatchIndex{resources: make(map[int]resources.ResourceList)}
}

func (rbw *resourcesByWatchIndex) set(key int, val resources.ResourceList) *resourcesByWatchIndex {
	rbw.access.Lock()
	rbw.resources[key] = val
	rbw.access.Unlock()
	return rbw
}

func (rbw *resourcesByWatchIndex) delete(key int) *resourcesByWatchIndex {
	rbw.access.Lock()
	delete(rbw.resources, key)
	rbw.access.Unlock()
	return rbw
}

func (rbw *resourcesByWatchIndex) merge() resources.ResourceList {
	rbw.access.RLock()
	var merged resources.ResourceList
	for _, list := range rbw.resources {
		merged = append(merged, list...)
	}
	rbw.access.RUnlock()
	return merged.Sort()
}
