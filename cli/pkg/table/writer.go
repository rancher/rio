package table

import (
	"encoding/json"
	"io"
	"os"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/docker/go-units"
	"github.com/rancher/norman/types/convert"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

var (
	idsHeader = [][]string{
		{"ID", "ID"},
	}
)

type Writer struct {
	quite         bool
	HeaderFormat  string
	ValueFormat   string
	err           error
	headerPrinted bool
	Writer        *tabwriter.Writer
	funcMap       map[string]interface{}
}

type FormatFunc interface{}

func NewWriter(values [][]string, ctx *cli.Context) *Writer {
	if ctx.Bool("ids") {
		values = append(idsHeader, values...)
	}

	t := &Writer{
		Writer: tabwriter.NewWriter(os.Stdout, 10, 1, 3, ' ', 0),
		funcMap: map[string]interface{}{
			"stackScopedName": FormatStackScopedName,
			"ago":             FormatCreated,
			"json":            FormatJSON,
			"yaml":            FormatYAML,
			"first":           FormatFirst,
		},
	}
	t.HeaderFormat, t.ValueFormat = SimpleFormat(values)

	if ctx.Bool("quiet") {
		t.HeaderFormat = ""
		t.ValueFormat = "{{.ID}}\n"
	}

	customFormat := ctx.String("format")
	if customFormat == "json" {
		t.HeaderFormat = ""
		t.ValueFormat = "json"
	} else if customFormat == "yaml" {
		t.HeaderFormat = ""
		t.ValueFormat = "yaml"
	} else if customFormat != "" {
		t.ValueFormat = customFormat + "\n"
		t.HeaderFormat = ""
	}

	return t
}

func (t *Writer) AddFormatFunc(name string, f FormatFunc) {
	t.funcMap[name] = f
}

func (t *Writer) Err() error {
	return t.err
}

func (t *Writer) writeHeader() {
	if t.HeaderFormat != "" && !t.headerPrinted {
		t.headerPrinted = true
		t.err = t.printTemplate(t.Writer, t.HeaderFormat, struct{}{})
		if t.err != nil {
			return
		}
	}
}

func (t *Writer) Write(obj interface{}) {
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

func (t *Writer) Close() error {
	if t.err != nil {
		return t.err
	}
	t.writeHeader()
	if t.err != nil {
		return t.err
	}
	return t.Writer.Flush()
}

func (t *Writer) printTemplate(out io.Writer, templateContent string, obj interface{}) error {
	tmpl, err := template.New("").Funcs(t.funcMap).Parse(templateContent)
	if err != nil {
		return err
	}

	return tmpl.Execute(out, obj)
}

func FormatStackScopedName(data, data2 interface{}) (string, error) {
	stackName, ok := data.(string)
	if !ok {
		return "", nil
	}

	serviceName, ok := data2.(string)
	if !ok {
		return "", nil
	}

	if stackName == "default" {
		return serviceName, nil
	}

	return stackName + "/" + serviceName, nil
}

func FormatCreated(data interface{}) (string, error) {
	s, ok := data.(string)
	if !ok {
		return "", nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return "", err
	}

	return units.HumanDuration(time.Now().UTC().Sub(t)) + " ago", nil
}

func FormatJSON(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "    ")
	return string(bytes) + "\n", err
}

func FormatYAML(data interface{}) (string, error) {
	bytes, err := yaml.Marshal(data)
	return string(bytes) + "\n", err
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
