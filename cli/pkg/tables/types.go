package tables

import (
	"sort"

	"github.com/rancher/rio/cli/pkg/table"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

type TableWriter interface {
	Write(obs []runtime.Object) error
	WriteObjects(obs []Object) error
	TableWriter() table.Writer
}

type Object interface {
	Object() runtime.Object
}

type ObjectHolder struct {
	O runtime.Object
}

func (o ObjectHolder) Object() runtime.Object {
	return o.O
}

type Config interface {
	table.WriterConfig
	Domain() (*v1.ClusterDomain, error)
}

type tableWriter struct {
	writer  table.Writer
	context interface{}
}

type data struct {
	Namespace string
	Context   interface{}
	Data      Object
	Obj       runtime.Object
}

func (t *tableWriter) Write(objs []runtime.Object) (err error) {
	var os []Object
	for _, o := range objs {
		os = append(os, ObjectHolder{o})
	}
	return t.WriteObjects(os)
}

func (t *tableWriter) WriteObjects(objs []Object) (err error) {
	sort.Slice(objs, func(i, j int) bool {
		leftMeta, _ := meta.Accessor(objs[i].Object())
		rightMeta, _ := meta.Accessor(objs[j].Object())
		if leftMeta.GetNamespace() != rightMeta.GetNamespace() {
			return leftMeta.GetNamespace() < rightMeta.GetNamespace()
		}
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
		metaObj, err := meta.Accessor(obj.Object())
		if err != nil {
			return err
		}

		t.writer.Write(&data{
			Namespace: metaObj.GetNamespace(),
			Context:   t.context,
			Data:      obj,
			Obj:       obj.Object(),
		})
	}

	return t.writer.Err()
}

func (t *tableWriter) TableWriter() table.Writer {
	return t.writer
}
