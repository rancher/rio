package apply

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/rancher/norman/types"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

type Data struct {
	errs      []error
	GroupID   string
	Objects   map[string][]runtime.Object
	Empty     map[string][]string
	Injectors []ConfigInjector
}

type Namespaced struct {
	Objects []runtime.Object
	Empty   []string
}

func (a *Data) Apply() error {
	if err := a.Err(); err != nil {
		return err
	}

	nspaces := map[string]bool{}
	for k := range a.Objects {
		nspaces[k] = true
	}
	for k := range a.Empty {
		nspaces[k] = true
	}

	if err := a.applyNS(""); err != nil {
		return a.err(err)
	}

	for ns := range nspaces {
		if ns == "" {
			continue
		}
		if err := a.applyNS(ns); err != nil {
			a.err(err)
		}
	}

	return a.Err()
}

func (a *Data) applyNS(ns string) error {
	return Apply(a.Objects[ns], a.Empty[ns], ns, a.GroupID, a.Injectors...)
}

func (a *Data) Err() error {
	return types.NewErrors(a.errs...)
}

func (a *Data) err(err error) error {
	if err == nil {
		return nil
	}
	a.errs = append(a.errs, err)
	return err
}

func gk(group, kind string) string {
	if group == "" {
		return kind
	}
	return kind + "." + group
}

func (a *Data) Add(ns string, group, kind string, runtimeObjectMap interface{}) {
	v := reflect.ValueOf(runtimeObjectMap)
	t := v.Type()
	if t.Kind() != reflect.Map {
		a.err(fmt.Errorf("obj must be map got %v", t))
		return
	}

	if v.Len() == 0 {
		if a.Empty == nil {
			a.Empty = map[string][]string{}
		}
		a.Empty[ns] = append(a.Empty[ns], gk(group, kind))

		return
	}

	var keys []string
	for _, key := range v.MapKeys() {
		s, ok := key.Interface().(string)
		if !ok {
			a.err(fmt.Errorf("map value must be a string"))
			return
		}
		keys = append(keys, s)
	}

	sort.Strings(keys)
	for _, key := range keys {
		ro, ok := v.MapIndex(reflect.ValueOf(key)).Interface().(runtime.Object)
		if !ok {
			a.err(fmt.Errorf("map value must be a runtime.Object"))
			return
		}

		mobj, err := meta.Accessor(ro)
		if err != nil {
			a.err(err)
			return
		}

		if mobj.GetNamespace() != ns {
			a.err(fmt.Errorf("namespace of object %s/%s does not match %s", mobj.GetNamespace(), mobj.GetName(), ns))
			return
		}

		if a.Objects == nil {
			a.Objects = map[string][]runtime.Object{}
		}
		a.Objects[ns] = append(a.Objects[ns], ro)
	}
}
