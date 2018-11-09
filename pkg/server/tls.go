package server

import (
	"context"
	"net/http"

	"github.com/rancher/rancher/pkg/dynamiclistener"
	"github.com/rancher/rancher/pkg/settings"
	"github.com/rancher/rancher/pkg/tls"
	"github.com/rancher/rio/types/apis/space.cattle.io/v1beta1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func startTLS(ctx context.Context, httpPort, httpsPort int, handler http.Handler) error {
	v3.From(ctx)
	s := &storage{
		listenConfigs:      v1beta1.From(ctx).ListenConfigs(""),
		listenConfigLister: v1beta1.From(ctx).ListenConfigs("").Controller().Lister(),
	}
	s2 := &storage2{
		listenConfigs: s.listenConfigs,
	}

	lc, err := tls.ReadTLSConfig(nil)
	if err != nil {
		return err
	}

	if err := tls.SetupListenConfig(s2, false, lc); err != nil {
		return err
	}

	server := dynamiclistener.NewServer(ctx, s, handler, httpPort, httpsPort)
	settings.CACerts.Set(lc.CACerts)
	_, err = server.Enable(lc)
	return err
}

type storage2 struct {
	listenConfigs v1beta1.ListenConfigInterface
}

func (s *storage2) Create(lc *v3.ListenConfig) (*v3.ListenConfig, error) {
	createLC := &v1beta1.ListenConfig{
		ListenConfig: *lc,
	}
	createLC.APIVersion = "space.cattle.io/v1beta1"

	result, err := s.listenConfigs.Create(createLC)
	if err != nil {
		return nil, err
	}
	return &result.ListenConfig, nil
}

func (s *storage2) Get(name string, opts metav1.GetOptions) (*v3.ListenConfig, error) {
	lc, err := s.listenConfigs.Get(name, opts)
	if err != nil {
		return nil, err
	}
	return &lc.ListenConfig, nil
}

func (s *storage2) Update(lc *v3.ListenConfig) (*v3.ListenConfig, error) {
	updateLC := &v1beta1.ListenConfig{
		ListenConfig: *lc,
	}
	updateLC.APIVersion = "space.cattle.io/v1beta1"

	result, err := s.listenConfigs.Update(updateLC)
	if err != nil {
		return nil, err
	}
	return &result.ListenConfig, nil
}

type storage struct {
	listenConfigs      v1beta1.ListenConfigInterface
	listenConfigLister v1beta1.ListenConfigLister
}

func (s *storage) Update(lc *v3.ListenConfig) (*v3.ListenConfig, error) {
	updateLC := &v1beta1.ListenConfig{
		ListenConfig: *lc,
	}
	updateLC.APIVersion = "space.cattle.io/v1beta1"

	updateLC, err := s.listenConfigs.Update(updateLC)
	if err != nil {
		return nil, err
	}
	return &updateLC.ListenConfig, nil
}

func (s *storage) Get(namespace, name string) (*v3.ListenConfig, error) {
	lc, err := s.listenConfigLister.Get(namespace, name)
	if err != nil {
		return nil, err
	}
	return &lc.ListenConfig, nil
}
