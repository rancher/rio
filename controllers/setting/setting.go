package setting

import (
	"context"
	"os"
	"strings"

	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	sp := &settingsProvider{
		settings:       rContext.Global.Setting,
		settingsLister: rContext.Global.Setting.Cache(),
	}
	return settings.SetProvider(sp)
}

type settingsProvider struct {
	settings       projectv1.SettingClient
	settingsLister projectv1.SettingClientCache
	fallback       map[string]string
}

func (s *settingsProvider) Get(name string) string {
	obj, err := s.settingsLister.Get("", name)
	if err != nil {
		return s.fallback[name]
	}
	if obj.Value == "" {
		return obj.Default
	}
	return obj.Value
}

func (s *settingsProvider) Set(name, value string) error {
	obj, err := s.settings.Get("", name, v1.GetOptions{})
	if err != nil {
		return err
	}

	obj.Value = value
	_, err = s.settings.Update(obj)
	return err
}

func (s *settingsProvider) SetIfUnset(name, value string) error {
	obj, err := s.settings.Get("", name, v1.GetOptions{})
	if err != nil {
		return err
	}

	if obj.Value != "" {
		return nil
	}

	obj.Value = value
	_, err = s.settings.Update(obj)
	return err
}

func (s *settingsProvider) SetAll(settings map[string]settings.Setting) error {
	fallback := map[string]string{}

	for name, setting := range settings {
		key := "RIO_" + strings.ToUpper(strings.Replace(name, "-", "_", -1))
		value := os.Getenv(key)

		obj, err := s.settings.Get("", setting.Name, v1.GetOptions{})
		if errors.IsNotFound(err) {
			newSetting := &projectv1.Setting{}
			newSetting.Name = setting.Name
			newSetting.Default = setting.Default
			if value != "" {
				newSetting.Value = value
			}
			fallback[newSetting.Name] = newSetting.Value
			_, err := s.settings.Create(newSetting)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			update := false
			if obj.Default != setting.Default {
				obj.Default = setting.Default
				update = true
			}
			if value != "" && obj.Value != value {
				obj.Value = value
				update = true
			}
			fallback[obj.Name] = obj.Value
			if update {
				_, err := s.settings.Update(obj)
				if err != nil {
					return err
				}
			}
		}
	}

	s.fallback = fallback

	return nil
}
