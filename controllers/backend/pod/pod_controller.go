package pod

import (
	"context"
	"sync"
	"time"

	"strings"

	"fmt"

	"github.com/rancher/rancher/pkg/controllers/user/approuter"
	"github.com/rancher/rancher/pkg/ticker"
	"github.com/rancher/rio/pkg/namespace"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	v12 "github.com/rancher/types/apis/core/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

var (
	refreshInterval = 5 * time.Minute
)

func Register(ctx context.Context, rContext *types.Context) {
	rdnsClient := approuter.NewClient(rContext.Core.Secrets(""),
		rContext.Core.Secrets("").Controller().Lister(),
		settings.RioSystemNamespace)
	rdnsClient.SetBaseURL(settings.RDNSURL.Get())

	ns := namespace.StackNamespace(settings.RioSystemNamespace, "istio")
	pc := &PodController{
		namespace:  ns,
		rdnsClient: rdnsClient,
		podLister:  rContext.Core.Pods(ns).Controller().Lister(),
		nodeLister: rContext.Core.Nodes("").Controller().Lister(),
		ips:        map[string]string{},
		dirty:      true,
	}

	rContext.Core.Pods(ns).Controller().AddHandler("pod-controller", pc.sync)

	go func() {
		for range ticker.Context(ctx, 6*time.Hour) {
			pc.renew()
		}
	}()
}

type PodController struct {
	sync.Mutex

	namespace   string
	rdnsClient  *approuter.Client
	podLister   v12.PodLister
	nodeLister  v12.NodeLister
	ips         map[string]string
	dirty       bool
	refreshTime time.Time
}

func (p *PodController) renew() error {
	if err := p.sync("_none_", nil); err != nil {
		return err
	}
	_, err := p.rdnsClient.RenewDomain()
	return err
}

func (p *PodController) sync(key string, pod *v1.Pod) error {
	if !isGateway(pod) {
		return nil
	}

	p.Lock()
	defer p.Unlock()

	var err error
	if p.shouldFullSync() {
		err = p.fullSync()
	} else {
		err = p.mod(p.ips, key, pod)
	}
	if err != nil {
		return err
	}

	if p.dirty {
		if err := p.update(); err != nil {
			return err
		}
	}

	return nil
}

func (p *PodController) update() error {
	var (
		fqdn string
		ips  []string
		err  error
	)
	for _, ip := range p.ips {
		ips = append(ips, ip)
	}

	if len(ips) == 0 {
		p.dirty = false
		return nil
	}

	if len(ips) == 1 && ips[0] == "127.0.0.1" {
		fqdn = nip(ips[0])
	} else {
		_, fqdn, err = p.rdnsClient.ApplyDomain(ips)
		if err != nil {
			if settings.ClusterDomain.Get() == "" {
				settings.ClusterDomain.Set(nip(ips[0]))
			}
			return err
		}
	}

	settings.ClusterDomain.Set(fqdn)
	if err == nil {
		p.dirty = false
	}
	return err
}

func nip(ip string) string {
	return fmt.Sprintf("%s.nip.io", ip)
}

func (p *PodController) fullSync() error {
	newIps := map[string]string{}
	req, err := labels.NewRequirement("gateway", selection.Equals, []string{"external"})
	if err != nil {
		return err
	}

	// lister is already scoped to a single namespace
	pods, err := p.podLister.List("", labels.NewSelector().Add(*req))
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		return nil
	}

	for _, pod := range pods {
		p.mod(newIps, pod.Namespace+"/"+pod.Name, pod)
	}

	changed := false
	if len(newIps) == len(p.ips) {
		for k := range newIps {
			if newIps[k] != p.ips[k] {
				changed = true
				break
			}
		}
	} else {
		changed = true
	}

	p.refreshTime = time.Now()

	if changed {
		p.dirty = true
		p.ips = newIps
		return p.update()
	}

	return nil
}

func (p *PodController) shouldFullSync() bool {
	return time.Now().Sub(p.refreshTime) > refreshInterval
}

func isGateway(pod *v1.Pod) bool {
	if pod == nil {
		return true
	}
	if pod.Spec.NodeName == "" {
		return false
	}
	return pod.Labels["gateway"] == "external"
}

func (p *PodController) mod(ips map[string]string, key string, pod *v1.Pod) error {
	if pod == nil {
		if _, ok := ips[key]; ok {
			p.dirty = true
			delete(ips, key)
		}
		return nil
	}

	node, err := p.nodeLister.Get("", pod.Spec.NodeName)
	if err != nil || node == nil {
		return err
	}

	addr := getAddr(node, v1.NodeExternalIP, v1.NodeInternalIP)
	if addr == "" {
		return nil
	}

	if ips[key] != addr {
		p.dirty = true
		ips[key] = addr
	}

	return nil
}

func IsD4x(node *v1.Node) bool {
	return strings.Contains(strings.ToLower(node.Status.NodeInfo.OSImage), "docker for")
}

func getAddr(node *v1.Node, types ...v1.NodeAddressType) string {
	if IsD4x(node) {
		return "127.0.0.1"
	}
	for _, addrType := range types {
		for _, addr := range node.Status.Addresses {
			if addr.Type == addrType && addr.Address != "" {
				return addr.Address
			}
		}
	}

	return ""
}
