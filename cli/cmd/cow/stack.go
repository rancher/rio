package cow

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/rivo/tview"
)

var (
	stack = resourceKind{
		kind:  "stack",
		title: "Stack",
	}

	stackRefresher = func(b *bytes.Buffer) error {
		cmd := exec.Command("rio", "stack")
		errBuffer := &strings.Builder{}
		cmd.Stdout = b
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		return nil
	}
)

func (app *appView) Stack() tview.Primitive {
	feeder := &cmdDataFeeder{
		buffer:    new(bytes.Buffer),
		refresher: stackRefresher,
	}

	t := newTableView()
	t.init(app, stack, feeder, defaultAction)

	if err := t.run(); err != nil {
		return t.updateStatus(err.Error(), true)
	}
	return t
}
