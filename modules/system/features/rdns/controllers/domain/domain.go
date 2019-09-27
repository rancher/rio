package domain

import (
	"context"
	"fmt"
	"time"

	approuter "github.com/rancher/rdns-server/client"
	projectv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	v1 "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/ticker"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Controller struct {
	ctx                 context.Context
	namespace           string
	rdnsClient          *approuter.Client
	clusterDomainClient v1.ClusterDomainClient
	started             bool
}

func Register(ctx context.Context, rContext *types.Context) error {
	rdnsClient := approuter.NewClient(rContext.Core.Core().V1().Secret(),
		rContext.Core.Core().V1().Secret().Cache(),
		rContext.Namespace)
	rdnsClient.SetBaseURL(constants.RDNSURL)

	g := &Controller{
		ctx:        ctx,
		namespace:  rContext.Namespace,
		rdnsClient: rdnsClient,
	}

	handler := v1.UpdateClusterDomainOnChange(rContext.Global.Admin().V1().ClusterDomain().Updater(),
		g.onChange)
	rContext.Global.Admin().V1().ClusterDomain().OnChange(ctx, "rdns", handler)
	return nil
}

func (g *Controller) onChange(key string, obj *projectv1.ClusterDomain) (*projectv1.ClusterDomain, error) {
	if obj == nil || key != g.namespace+"/"+constants.ClusterDomainName {
		return nil, nil
	}

	return obj, projectv1.ClusterDomainConditionReady.Do(func() (runtime.Object, error) {
		domain, err := g.getDomain(obj)
		if err != nil {
			return obj, err
		}
		g.start()
		obj.Status.ClusterDomain = domain
		return obj, err
	})
}

func (g *Controller) getDomain(obj *projectv1.ClusterDomain) (string, error) {
	var hosts []string
	var cname bool

	for _, addr := range obj.Spec.Addresses {
		if addr.Hostname != "" {
			cname = true
			hosts = append(hosts, addr.Hostname)
			break
		}
		if addr.IP != "" {
			hosts = append(hosts, addr.IP)
		}
	}

	if err := g.ensureDomainExists(hosts, cname); err != nil {
		return "", err
	}

	subDomains := map[string][]string{}
	for _, subDomain := range obj.Spec.Subdomains {
		for _, addr := range subDomain.Addresses {
			if addr.IP == "" {
				continue
			}
			subDomains[subDomain.Name] = append(subDomains[subDomain.Name], addr.IP)
		}
	}

	return g.rdnsClient.UpdateDomain(hosts, subDomains, cname)
}

func (g *Controller) ensureDomainExists(hosts []string, cname bool) error {
	domain, err := g.rdnsClient.GetDomain(cname)
	if err != nil || domain != nil {
		return err
	}

	if _, err := g.rdnsClient.CreateDomain(hosts, cname); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(g.ctx, time.Second*5)
	defer cancel()
	wait.JitterUntil(func() {
		domain, err = g.rdnsClient.GetDomain(cname)
		if err != nil {
			logrus.Debug("failed to get domain")
		}
	}, time.Second, 1.3, true, ctx.Done())

	if domain == nil {
		return fmt.Errorf("failed to create domain")
	}

	return nil
}

func (g *Controller) start() {
	if g.started {
		return
	}

	g.started = true
	go func() {
		for range ticker.Context(g.ctx, 6*time.Hour) {
			if err := g.renew(); err != nil {
				logrus.Errorf("failed to renew domain: %v", err)
			}
		}
	}()
}

func (g *Controller) renew() error {
	if _, err := g.rdnsClient.RenewDomain(); err != nil {
		return err
	}
	return nil
}
