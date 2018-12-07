package server

import (
	"context"
	"net/http"

	"github.com/rancher/rancher/pkg/dynamiclistener"
	"github.com/rancher/rancher/pkg/settings"
	"github.com/rancher/rancher/pkg/tls"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func startTLS(ctx context.Context, httpPort, httpsPort int, handler http.Handler) error {
	v3.From(ctx)
	s := &storage{
		listenConfigs:      projectv1.From(ctx).ListenConfigs(""),
		listenConfigLister: projectv1.From(ctx).ListenConfigs("").Controller().Lister(),
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
	listenConfigs projectv1.ListenConfigInterface
}

func (s *storage2) Create(lc *v3.ListenConfig) (*v3.ListenConfig, error) {
	createLC := &projectv1.ListenConfig{
		ListenConfig: *lc,
	}
	createLC.APIVersion = "project.rio.cattle.io/v1"

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
	updateLC := &projectv1.ListenConfig{
		ListenConfig: *lc,
	}
	updateLC.APIVersion = "project.rio.cattle.io/v1"

	result, err := s.listenConfigs.Update(updateLC)
	if err != nil {
		return nil, err
	}
	return &result.ListenConfig, nil
}

type storage struct {
	listenConfigs      projectv1.ListenConfigInterface
	listenConfigLister projectv1.ListenConfigLister
}

func (s *storage) Update(lc *v3.ListenConfig) (*v3.ListenConfig, error) {
	updateLC := &projectv1.ListenConfig{
		ListenConfig: *lc,
	}
	updateLC.APIVersion = "project.rio.cattle.io/v1"

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
