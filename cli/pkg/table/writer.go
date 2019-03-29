package table

import (
	"encoding/json"
	"io"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/rancher/rio/pkg/apis/common"

	"github.com/Masterminds/sprig"
	"github.com/davecgh/go-spew/spew"
	"github.com/docker/go-units"
	"github.com/rancher/mapper/convert"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	idsHeader = [][]string{
		{"ID", "ID"},
	}

	localFuncMap = map[string]interface{}{
		"ago":           FormatCreated,
		"json":          FormatJSON,
		"jsoncompact":   FormatJSONCompact,
		"yaml":          FormatYAML,
		"first":         FormatFirst,
		"dump":          FormatSpew,
		"toJson":        ToJSON,
		"boolToStar":    BoolToStar,
		"state":         State,
		"transitioning": Transitioning,
	}
)

type Writer interface {
	Write(obj interface{})
	Close() error
	Err() error
	AddFormatFunc(name string, f FormatFunc)
}

type writer struct {
	closed        bool
	quite         bool
	HeaderFormat  string
	ValueFormat   string
	err           error
	headerPrinted bool
	Writer        *tabwriter.Writer
	funcMap       map[string]interface{}
}

type FormatFunc interface{}

type WriterConfig interface {
	IDs() bool
	Quiet() bool
	Format() string
	Writer() io.Writer
}

func NewWriter(values [][]string, config WriterConfig) Writer {
	if config.IDs() {
		values = append(idsHeader, values...)
	}

	funcMap := sprig.TxtFuncMap()
	for k, v := range localFuncMap {
		funcMap[k] = v
	}

	t := &writer{
		Writer:  tabwriter.NewWriter(config.Writer(), 10, 1, 3, ' ', 0),
		funcMap: funcMap,
	}

	t.HeaderFormat, t.ValueFormat = SimpleFormat(values)

	if config.Quiet() {
		t.HeaderFormat = ""
		t.ValueFormat = "{{.ID}}\n"
	}

	customFormat := config.Format()
	if customFormat == "json" {
		t.HeaderFormat = ""
		t.ValueFormat = "json"
	} else if customFormat == "jsoncompact" {
		t.HeaderFormat = ""
		t.ValueFormat = "jsoncompact"
	} else if customFormat == "yaml" {
		t.HeaderFormat = ""
		t.ValueFormat = "yaml"
	} else if customFormat != "" {
		t.ValueFormat = customFormat + "\n"
		t.HeaderFormat = ""
	}

	return t
}

func (t *writer) AddFormatFunc(name string, f FormatFunc) {
	t.funcMap[name] = f
}

func (t *writer) Err() error {
	return t.Close()
}

func (t *writer) writeHeader() {
	if t.HeaderFormat != "" && !t.headerPrinted {
		t.headerPrinted = true
		t.err = t.printTemplate(t.Writer, t.HeaderFormat, struct{}{})
		if t.err != nil {
			return
		}
	}
}

func (t *writer) Write(obj interface{}) {
	if t.err != nil {
		return
	}

	t.writeHeader()
	if t.err != nil {
		return
	}

	if t.ValueFormat == "json" {
		content, err := FormatJSON(obj)
		t.err = err
		if t.err != nil {
			return
		}
		_, t.err = t.Writer.Write([]byte(content + "\n"))
	} else if t.ValueFormat == "jsoncompact" {
		content, err := FormatJSONCompact(obj)
		t.err = err
		if t.err != nil {
			return
		}
		_, t.err = t.Writer.Write([]byte(content))
	} else if t.ValueFormat == "yaml" {
		content, err := FormatYAML(obj)
		t.err = err
		if t.err != nil {
			return
		}
		t.Writer.Write([]byte("---\n"))
		_, t.err = t.Writer.Write([]byte(content + "\n"))
	} else {
		t.err = t.printTemplate(t.Writer, t.ValueFormat, obj)
	}
}

func (t *writer) Close() error {
	if t.closed {
		return t.err
	}
	if t.err != nil {
		return t.err
	}

	defer func() {
		t.closed = true
	}()
	t.writeHeader()
	if t.err != nil {
		return t.err
	}
	return t.Writer.Flush()
}

func (t *writer) printTemplate(out io.Writer, templateContent string, obj interface{}) error {
	tmpl, err := template.New("").Funcs(t.funcMap).Parse(templateContent)
	if err != nil {
		return err
	}

	return tmpl.Execute(out, obj)
}

func FormatStackScopedName(defaultStackName string) func(interface{}, interface{}) (string, error) {
	return func(data, data2 interface{}) (string, error) {
		stackName, ok := data.(string)
		if !ok {
			return "", nil
		}

		serviceName, ok := data2.(string)
		if !ok {
			return "", nil
		}

		if stackName == defaultStackName {
			return serviceName, nil
		}

		return stackName + "/" + serviceName, nil
	}
}

func FormatCreated(data interface{}) (string, error) {
	t, ok := data.(metav1.Time)
	if !ok {
		return "", nil
	}

	return units.HumanDuration(time.Now().UTC().Sub(t.Time)) + " ago", nil
}

func FormatJSON(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "    ")
	return string(bytes) + "\n", err
}

func FormatJSONCompact(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	return string(bytes) + "\n", err
}

func FormatYAML(data interface{}) (string, error) {
	bytes, err := yaml.Marshal(data)
	return string(bytes) + "\n", err
}

func FormatSpew(data interface{}) (string, error) {
	return spew.Sdump(data), nil
}

func FormatFirst(data, data2 interface{}) (string, error) {
	str := convert.ToString(data)
	if str != "" {
		return str, nil
	}

	str = convert.ToString(data2)
	if str != "" {
		return str, nil
	}

	return "", nil
}

func ToJSON(data interface{}) (map[string]interface{}, error) {
	return convert.EncodeToMap(data)
}

func BoolToStar(obj interface{}) (string, error) {
	if b, ok := obj.(bool); ok && b {
		return "*", nil
	}
	return "", nil
}

func State(obj interface{}) (string, error) {
	if b, ok := obj.(common.StateGetter); ok {
		return b.State().State, nil
	}
	return "", nil
}

func Transitioning(obj interface{}) (string, error) {
	if b, ok := obj.(common.StateGetter); ok {
		return b.State().Message, nil
	}
	return "", nil
}
