package tables

import (
	"fmt"
	"sort"

	"github.com/rancher/rio/cli/pkg/table"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

type TableWriter interface {
	Write(obs []runtime.Object) error
	TableWriter() table.Writer
}

type Config interface {
	table.WriterConfig
	GetDefaultNamespace() string
	GetSetNamespace() string
	Domain() (string, error)
}

type tableWriter struct {
	writer  table.Writer
	context interface{}
}

type data struct {
	Name    string
	Context interface{}
	Obj     runtime.Object
}

func (t *tableWriter) Write(objs []runtime.Object) (err error) {
	sort.Slice(objs, func(i, j int) bool {
		leftMeta, _ := meta.Accessor(objs[i])
		rightMeta, _ := meta.Accessor(objs[j])
		leftCreated := leftMeta.GetCreationTimestamp()
		return leftCreated.After(rightMeta.GetCreationTimestamp().Time)
	})

	defer func() {
		cErr := t.writer.Close()
		if cErr != nil {
			err = cErr
		}
	}()

	for _, obj := range objs {
		metaObj, err := meta.Accessor(obj)
		if err != nil {
			return err
		}
		id := metaObj.GetName()
		if metaObj.GetNamespace() != "" {
			id = fmt.Sprintf("%s/%s", metaObj.GetNamespace(), id)
		}

		t.writer.Write(&data{
			Name:    id,
			Context: t.context,
			Obj:     obj,
		})
	}

	return t.writer.Err()
}

func (t *tableWriter) TableWriter() table.Writer {
	return t.writer
}
