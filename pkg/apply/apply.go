package apply

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"crypto/sha1"

	"github.com/docker/docker/pkg/reexec"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const refreshAfter = 2 * time.Minute

var applied sync.Map

type appliedValue struct {
	time time.Time
	hash [20]byte
}

type ConfigInjector func(config []byte) ([]byte, error)

func Content(content []byte) error {
	errOutput := &bytes.Buffer{}
	cmd := reexec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = bytes.NewReader(content)
	cmd.Stdout = nil
	cmd.Stderr = errOutput

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to apply: %s", errOutput.String())
	}

	return nil
}

func Apply(objects []runtime.Object, groupID string, generation int64, injectors ...ConfigInjector) error {
	if len(objects) == 0 {
		return nil
	}

	ns, whitelist, content, err := constructApplyData(objects, groupID, generation)
	if err != nil {
		return err
	}

	for _, inject := range injectors {
		if content, err = inject(content); err != nil {
			return err
		}
	}

	return execApply(ns, whitelist, content, groupID)
}

func ApplyAnyNamespace(objects []runtime.Object, groupID string, generation int64) error {
	if len(objects) == 0 {
		return nil
	}

	_, whitelist, content, err := constructApplyData(objects, groupID, generation)
	if err != nil {
		return err
	}

	return execApply("", whitelist, content, groupID)
}

func execApply(ns string, whitelist map[string]bool, content []byte, groupID string) error {
	key, val, ok := shouldApply(ns, content, groupID)
	if !ok {
		return nil
	}

	output := &bytes.Buffer{}
	errOutput := &bytes.Buffer{}
	cmd := reexec.Command("kubectl", "-n", ns, "apply", "--force", "--grace-period", "120", "--prune", "-l", "apply.cattle.io/groupID="+groupID, "-o", "json", "-f", "-")
	for group := range whitelist {
		cmd.Args = append(cmd.Args, "--prune-whitelist="+group)
	}
	cmd.Stdin = bytes.NewReader(content)
	cmd.Stdout = output
	cmd.Stderr = errOutput

	if err := cmd.Run(); err != nil {
		logrus.Errorf("Failed to apply %s: %s", errOutput.String(), string(content))
		return fmt.Errorf("failed to apply: %s", errOutput.String())
	}

	if logrus.GetLevel() >= logrus.DebugLevel {
		fmt.Printf("Applied: %s", output.String())
	}

	applied.Store(key, val)
	return nil
}

func shouldApply(ns string, content []byte, groupID string) (interface{}, interface{}, bool) {
	key := ns + "/" + groupID
	val := &appliedValue{
		time: time.Now(),
		hash: sha1.Sum(content),
	}

	obj, ok := applied.Load(key)
	if !ok {
		return key, val, true
	}

	old := obj.(*appliedValue)
	if val.time.Sub(old.time) > refreshAfter {
		return key, val, true
	}

	if val.hash != old.hash {
		return key, val, true
	}

	return key, val, false
}

func constructApplyData(objects []runtime.Object, groupID string, generation int64) (string, map[string]bool, []byte, error) {
	buffer := &bytes.Buffer{}
	whitelist := map[string]bool{}
	ns := ""
	for i, obj := range objects {
		if i > 0 {
			buffer.WriteString("\n---\n")
		}

		objType, err := meta.TypeAccessor(obj)
		if err != nil {
			return "", nil, nil, fmt.Errorf("resource type data can not be accessed")
		}

		metaObj, ok := obj.(v1.Object)
		if !ok {
			return "", nil, nil, fmt.Errorf("resource type is not a meta object")
		}
		labels := metaObj.GetLabels()
		newLabels := map[string]string{}
		for k, v := range labels {
			newLabels[k] = v
		}
		newLabels["apply.cattle.io/groupID"] = groupID
		newLabels["apply.cattle.io/generationID"] = strconv.FormatInt(generation, 10)
		metaObj.SetLabels(newLabels)

		if len(ns) == 0 {
			ns = metaObj.GetNamespace()
		}

		gvk := fmt.Sprintf("%s/%s", objType.GetAPIVersion(), objType.GetKind())
		if len(strings.Split(gvk, "/")) < 3 {
			gvk = "/" + gvk
		}
		whitelist[gvk] = true

		bytes, err := yaml.Marshal(obj)
		if err != nil {
			return "", nil, nil, errors.Wrapf(err, "failed to encode %s/%s/%s/%s", objType.GetAPIVersion(), objType.GetKind(), metaObj.GetNamespace(), metaObj.GetName())
		}
		buffer.Write(bytes)
	}

	return ns, whitelist, buffer.Bytes(), nil
}
