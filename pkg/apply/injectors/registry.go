package injectors

import "github.com/rancher/rio/pkg/apply"

var (
	injectors = map[string]apply.ConfigInjector{}
)

func Register(name string, injector apply.ConfigInjector) {
	injectors[name] = injector
}

func Get(name string) apply.ConfigInjector {
	return injectors[name]
}
