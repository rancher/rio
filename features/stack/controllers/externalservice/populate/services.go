package populate

import (
	"net"
	"net/url"
	"strings"

	"github.com/rancher/norman/pkg/objectset"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	v1client "github.com/rancher/types/apis/core/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ServiceForExternalService(es *riov1.ExternalService, os *objectset.ObjectSet) error {
	svc := v1client.NewService(es.Namespace, es.Name, v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: es.Labels,
		},
	})

	target := es.Spec.Target
	if !strings.HasPrefix(target, "https://") && !strings.HasPrefix(target, "http://") {
		target = "http://" + target
	}
	u, err := url.Parse(target)
	if err != nil {
		return err
	}

	if ip := net.ParseIP(u.Host); ip == nil {
		svc.Spec = v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: u.Host,
		}
	} else {
		svc.Spec = v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: "None",
			ExternalIPs: []string{
				u.Host,
			},
		}
	}

	os.Add(svc)
	return nil
}
