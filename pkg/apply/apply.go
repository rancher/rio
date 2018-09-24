package apply

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

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

func Apply(objects []runtime.Object, empty []string, namespace, groupID string, injectors ...ConfigInjector) error {
	whitelist, content, err := constructApplyData(objects, groupID)
	if err != nil {
		return err
	}

	if len(content) != 0 {
		for _, inject := range injectors {
			if content, err = inject(content); err != nil {
				return err
			}
		}
	}

	return execApply(namespace, empty, whitelist, content, groupID)
}

func hash(id string) string {
	h := md5.Sum([]byte(id))
	return hex.EncodeToString(h[:])
}

func execApply(ns string, empty []string, whitelist map[string]bool, content []byte, groupID string) error {
	key, val, ok := shouldApply(ns, content, groupID)
	if !ok {
		return nil
	}

	hashGroupID := hash(groupID)
	if len(content) > 0 {
		args := []string{"-n", ns, "apply", "--force", "--grace-period", "120", "--prune", "-l", "apply.cattle.io/groupID=" + hashGroupID, "-o", "json", "-f", "-"}
		for group := range whitelist {
			args = append(args, "--prune-whitelist="+group)
		}
		err := run(groupID, content, args...)
		if err != nil {
			return err
		}
	}

	if len(empty) > 0 {
		args := []string{"-n", ns, "delete", "-l", "apply.cattle.io/groupID=" + hashGroupID, strings.Join(empty, ",")}
		err := run(groupID, nil, args...)
		if err != nil {
			return err
		}
	}

	applied.Store(key, val)
	return nil
}

func run(groupID string, content []byte, args ...string) error {
	output := &bytes.Buffer{}
	errOutput := &bytes.Buffer{}
	cmd := reexec.Command("kubectl")
	cmd.Args = append(cmd.Args, args...)
	cmd.Stdin = bytes.NewReader(content)
	cmd.Stdout = output
	cmd.Stderr = errOutput

	if err := cmd.Run(); err != nil {
		if logrus.GetLevel() >= logrus.DebugLevel {
			if len(content) > 0 {
				fmt.Printf("[\n%s\n]\nFailed to apply %s: %s", content, groupID, errOutput.String())
			}
		}
		return fmt.Errorf("failed to apply %s %v: %s", groupID, cmd.Args, errOutput.String())
	}

	if logrus.GetLevel() >= logrus.DebugLevel {
		fmt.Printf("Applied: %s", output.String())
	}

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

func constructApplyData(objects []runtime.Object, groupID string) (map[string]bool, []byte, error) {
	if len(objects) == 0 {
		return nil, nil, nil
	}

	groupID = hash(groupID)

	buffer := &bytes.Buffer{}
	whitelist := map[string]bool{}
	ns := ""
	for i, obj := range objects {
		if i > 0 {
			buffer.WriteString("\n---\n")
		}

		objType, err := meta.TypeAccessor(obj)
		if err != nil {
			return nil, nil, fmt.Errorf("resource type data can not be accessed")
		}

		metaObj, ok := obj.(v1.Object)
		if !ok {
			return nil, nil, fmt.Errorf("resource type is not a meta object")
		}
		labels := metaObj.GetLabels()
		newLabels := map[string]string{}
		for k, v := range labels {
			newLabels[k] = v
		}
		newLabels["apply.cattle.io/groupID"] = groupID
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
			return nil, nil, errors.Wrapf(err, "failed to encode %s/%s/%s/%s", objType.GetAPIVersion(), objType.GetKind(), metaObj.GetNamespace(), metaObj.GetName())
		}
		buffer.Write(bytes)
	}

	return whitelist, buffer.Bytes(), nil
}
