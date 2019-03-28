package cow

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/rivo/tview"
)

var (
	route = resourceKind{
		title: "Configs",
		kind:  "config",
	}

	routeRefresher = func(b *bytes.Buffer) error {
		cmd := exec.Command("rio", "route")
		errBuffer := &strings.Builder{}
		cmd.Stdout = b
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		return nil
	}
)

func (app *appView) Route() tview.Primitive {
	feeder := &cmdDataFeeder{
		buffer:    new(bytes.Buffer),
		refresher: routeRefresher,
	}

	t := newTableView()
	t.init(app, route, feeder, defaultAction)

	if err := t.run(); err != nil {
		return t.updateStatus(err.Error(), true)
	}
	return t
}
