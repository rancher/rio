package config

import (
	"encoding/json"
	"strings"

	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	corev1 "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ConfigName = "rio-config"

	ConfigController = ControllerConfig{}
)

type ControllerConfig struct {
	RunAPIValidatorWebhook bool
	WebhookPort            string
	WebhookHost            string
	IPAddresses            string
	Features               string
}

type Config struct {
	Features    map[string]FeatureConfig `json:"features,omitempty"`
	LetsEncrypt LetsEncrypt              `json:"letsEncrypt,omitempty"`
	RdnsURL     string                   `json:"rdnsUrl,omitempty"`
	Gateway     Gateway                  `json:"gateway,omitempty"`
}

type Gateway struct {
	StaticAddresses  []adminv1.Address `json:"staticAddresses,omitempty"`
	ServiceName      string            `json:"serviceName,omitempty"`
	ServiceNamespace string            `json:"serviceNamespace,omitempty"`

	IngressName      string `json:"ingressName,omitempty"`
	IngressNamespace string `json:"ingressNamespace,omitempty"`
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
	Enabled     *bool             `json:"enabled,omitempty"`
	Description string            `json:"description,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
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

func GetConfig(namespace string, client corev1.ConfigMapClient) (Config, error) {
	config, err := client.Get(namespace, ConfigName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			config = constructors.NewConfigMap(namespace, ConfigName, v1.ConfigMap{})
		} else {
			return Config{}, err
		}
	}
	return FromConfigMap(config)
}
