package config

import (
	"encoding/json"
	"strings"

	v1 "k8s.io/api/core/v1"

	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
)

var (
	ConfigName = "rio-config"
)

type Config struct {
	Features    map[string]FeatureConfig `json:"features,omitempty"`
	LetsEncrypt LetsEncrypt              `json:"letsEncrypt,omitempty"`
	Gateway     Gateway                  `json:"gateway,omitempty"`
}

type Gateway struct {
	StaticAddresses  []adminv1.Address `json:"staticAddresses,omitempty"`
	ServiceName      string            `json:"serviceName,omitempty"`
	ServiceNamespace string            `json:"serviceNamespace,omitempty"`
}

type LetsEncrypt struct {
	Account   string `json:"account,omitempty"`
	Email     string `json:"email,omitempty"`
	ServerURL string `json:"serverURL,omitempty"`
}

type Address struct {
	IP       string `json:"ip,omitempty"`
	Hostname string `json:"hostname,omitempty"`
}

type FeatureConfig struct {
	Enabled *bool             `json:"enabled,omitempty"`
	Options map[string]string `json:"options,omitempty"`
}

func FromConfigMap(cm *v1.ConfigMap) (result Config, err error) {
	configStr := cm.Data["config"]
	if configStr == "" {
		return
	}

	err = json.NewDecoder(strings.NewReader(configStr)).Decode(&result)
	return
}

func SetConfig(cm *v1.ConfigMap, config Config) (*v1.ConfigMap, error) {
	bytes, err := json.Marshal(config)
	if err != nil {
		return cm, err
	}

	cm = cm.DeepCopy()
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	cm.Data["config"] = string(bytes)
	return cm, nil
}
