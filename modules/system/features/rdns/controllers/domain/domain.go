package domain

import (
	"context"
	"fmt"
	"time"

	approuter "github.com/rancher/rdns-server/client"
	"github.com/rancher/rio/cli/pkg/constants"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/ticker"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
)

type Controller struct {
	ctx                 context.Context
	namespace           string
	rdnsClient          *approuter.Client
	clusterDomainClient v1.ClusterDomainClient
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

	handler := v1.UpdateClusterDomainOnChange(rContext.Global.Project().V1().ClusterDomain().Updater(),
		g.onChange)
	rContext.Global.Project().V1().ClusterDomain().OnChange(ctx, "rdns", handler)
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
		obj.Status.ClusterDomain = domain
		return obj, err
	})
}

func (g *Controller) getDomain(obj *projectv1.ClusterDomain) (string, error) {
	var ips []string
	for _, addr := range obj.Spec.Addresses {
		if addr.IP != "" {
			ips = append(ips, addr.IP)
		}
	}

	if len(ips) == 0 {
		return "", nil
	}

	if err := g.ensureDomainExists(ips); err != nil {
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

	return g.rdnsClient.UpdateDomain(ips, subDomains)
}

func (g *Controller) ensureDomainExists(ips []string) error {
	domain, err := g.rdnsClient.GetDomain()
	if err != nil || domain != nil {
		return err
	}

	if _, err := g.rdnsClient.CreateDomain(ips); err != nil {
		return err
	}

	if domain, err = g.rdnsClient.GetDomain(); err != nil {
		return err
	}

	if domain == nil {
		return fmt.Errorf("failed to create domain")
	}

	return nil
}

func (g *Controller) start() {
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
