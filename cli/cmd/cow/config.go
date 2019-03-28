package cow

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/rivo/tview"
)

var (
	config = resourceKind{
		title: "Configs",
		kind:  "config",
	}

	configRefresher = func(b *bytes.Buffer) error {
		cmd := exec.Command("rio", "configs")
		errBuffer := &strings.Builder{}
		cmd.Stdout = b
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		return nil
	}
)

func (app *appView) Config() tview.Primitive {
	feeder := &cmdDataFeeder{
		buffer:    new(bytes.Buffer),
		refresher: configRefresher,
	}

	t := newTableView()
	t.init(app, config, feeder, defaultAction)

	if err := t.run(); err != nil {
		return t.updateStatus(err.Error(), true)
	}
	return t
}
