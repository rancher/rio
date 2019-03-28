package cow

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"

	"github.com/rivo/tview"
)

var (
	service = resourceKind{
		title: "Services",
		kind:  "service",
	}

	serviceRefresher = func(b *bytes.Buffer) error {
		cmd := exec.Command("rio", "ps")
		errBuffer := &strings.Builder{}
		cmd.Stdout = b
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		return nil
	}

	serviceActions = append(defaultAction, []action{
		{
			name:        "exec",
			shortcut:    'x',
			description: "exec into a container or service",
		},
		{
			name:        "log",
			shortcut:    'l',
			description: "view logs of a service",
		},
	}...)
)

func (app *appView) Service() tview.Primitive {
	feeder := &cmdDataFeeder{
		buffer:    new(bytes.Buffer),
		refresher: serviceRefresher,
	}
	t := newTableView()
	t.init(app, service, feeder, serviceActions)

	if err := t.run(); err != nil {
		return t.updateStatus(err.Error(), true)
	}
	return t
}

func (app *appView) serviceActions() []action {
	var actions []action
	for _, ac := range defaultAction {
		actions = append(actions, ac)
	}
	actions = append(actions, action{
		name:        "log",
		shortcut:    'l',
		description: "view logs of a service",
	})
	actions = append(actions, action{
		name:        "exec",
		shortcut:    'x',
		description: "exec into a container or service",
	})
	return actions
}
