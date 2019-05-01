package tables

import (
	"sort"

	"github.com/rancher/rio/cli/pkg/table"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

type TableWriter interface {
	Write(obs []runtime.Object) error
	TableWriter() table.Writer
}

type tableWriter struct {
	writer  table.Writer
	context interface{}
}

type data struct {
	ID      string
	Context interface{}
	Obj     runtime.Object
}

func (t *tableWriter) Write(objs []runtime.Object) (err error) {
	sort.Slice(objs, func(i, j int) bool {
		leftMeta, _ := meta.Accessor(objs[i])
		rightMeta, _ := meta.Accessor(objs[j])
		return leftMeta.GetName() < rightMeta.GetName()
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

		t.writer.Write(&data{
			ID:      metaObj.GetName(),
			Context: t.context,
			Obj:     obj,
		})
	}

	return t.writer.Err()
}

func (t *tableWriter) TableWriter() table.Writer {
	return t.writer
}
