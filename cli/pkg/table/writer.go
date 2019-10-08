package table

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	gvk2 "github.com/rancher/wrangler/pkg/gvk"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"

	yaml2 "sigs.k8s.io/yaml"

	"github.com/Masterminds/sprig"
	"github.com/davecgh/go-spew/spew"
	"github.com/docker/go-units"
	"github.com/rancher/mapper/convert"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	idsHeader = [][]string{
		{"NAME", "Name"},
	}

	nsHeader = [][]string{
		{"NAMESPACE", "Namespace"},
	}

	localFuncMap = map[string]interface{}{
		"id":          FormatID,
		"ago":         FormatCreated,
		"json":        FormatJSON,
		"jsoncompact": FormatJSONCompact,
		"yaml":        FormatYAML,
		"first":       FormatFirst,
		"dump":        FormatSpew,
		"toJson":      ToJSON,
		"boolToStar":  BoolToStar,
		"array":       ToArray,
		"arrayFirst":  ToArrayFirst,
		"graph":       Graph,
		"pointer":     Pointer,
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
	Writer        io.Writer
	funcMap       map[string]interface{}
}

type FormatFunc interface{}

type WriterConfig interface {
	IDs() bool
	AllNamespaceSet() bool
	Quiet() bool
	Format() string
	Writer() io.Writer
	WithWriter(writer io.Writer)
}

func NewWriter(values [][]string, config WriterConfig) Writer {
	if config.IDs() {
		values = append(idsHeader, values...)
	}

	if config.AllNamespaceSet() {
		values = append(nsHeader, values...)
	}

	funcMap := sprig.TxtFuncMap()
	for k, v := range localFuncMap {
		funcMap[k] = v
	}

	t := &writer{
		funcMap: funcMap,
	}

	if config.Format() == "raw" {
		t.Writer = config.Writer()
	} else {
		t.Writer = tabwriter.NewWriter(config.Writer(), 10, 1, 3, ' ', 0)
	}

	t.HeaderFormat, t.ValueFormat = SimpleFormat(values)

	if config.Quiet() {
		t.HeaderFormat = ""
		t.ValueFormat = "{{.Name}}\n"
	}

	switch customFormat := config.Format(); customFormat {
	case "json":
		t.HeaderFormat = ""
		t.ValueFormat = "json"
	case "jsoncompact":
		t.HeaderFormat = ""
		t.ValueFormat = "jsoncompact"
	case "yaml":
		t.HeaderFormat = ""
		t.ValueFormat = "yaml"
	case "raw":
	default:
		if customFormat != "" {
			t.ValueFormat = customFormat + "\n"
			t.HeaderFormat = ""
		}
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

	switch t.ValueFormat {
	case "json":
		content, err := FormatJSON(obj)
		t.err = err
		if t.err != nil {
			return
		}
		_, t.err = t.Writer.Write([]byte(content + "\n"))
	case "jsoncompact":
		content, err := FormatJSONCompact(obj)
		t.err = err
		if t.err != nil {
			return
		}
		_, t.err = t.Writer.Write([]byte(content))
	case "yaml":
		content, err := FormatJSON(obj)
		t.err = err
		if t.err != nil {
			return
		}
		converted, err := yaml2.JSONToYAML([]byte(content))
		t.err = err
		if t.err != nil {
			return
		}
		t.Writer.Write([]byte("---\n"))
		_, t.err = t.Writer.Write(append(converted, []byte("\n")...))
	default:
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
	if _, ok := t.Writer.(*tabwriter.Writer); ok {
		return t.Writer.(*tabwriter.Writer).Flush()
	}
	return nil
}

func (t *writer) printTemplate(out io.Writer, templateContent string, obj interface{}) error {
	tmpl, err := template.New("").Funcs(t.funcMap).Parse(templateContent)
	if err != nil {
		return err
	}

	return tmpl.Execute(out, obj)
}

func ToArray(s []string) (string, error) {
	return strings.Join(s, ", "), nil
}

func ToArrayFirst(s []string) (string, error) {
	if len(s) > 0 {
		return s[0], nil
	}
	return "", nil
}

func Graph(value int) (string, error) {
	bars := int(float64(value) / 100.0 * 30)
	builder := &strings.Builder{}
	for i := 0; i < bars; i++ {
		if i == bars-1 {
			builder.WriteString(fmt.Sprintf("> %v", value))
			break
		}
		builder.WriteString("=")
	}
	return builder.String(), nil
}

func Pointer(data interface{}) string {
	if reflect.ValueOf(data).IsNil() {
		return ""
	}
	return fmt.Sprint(data)
}

func FormatID(obj runtime.Object) (string, error) {
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return "", err
	}

	kind := obj.GetObjectKind().GroupVersionKind().Kind
	if kind == "" {
		gvk, err := gvk2.Get(obj)
		if err != nil {
			return "", err
		}
		kind = gvk.Kind
	}

	if metaObj.GetNamespace() != "" {
		return fmt.Sprintf("%s/%s/%s", strings.ToLower(kind), metaObj.GetNamespace(), metaObj.GetName()), nil
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(kind), metaObj.GetName()), nil

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
	if b, ok := obj.(*bool); ok && b != nil && *b {
		return "*", nil
	}
	return "", nil
}
