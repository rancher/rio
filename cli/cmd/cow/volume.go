package cow

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/rivo/tview"
)

var (
	volume = resourceKind{
		kind:  "stack",
		title: "Stack",
	}

	volumeRefresher = func(b *bytes.Buffer) error {
		cmd := exec.Command("rio", "volume")
		errBuffer := &strings.Builder{}
		cmd.Stdout = b
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		return nil
	}
)

func (app *appView) Volume() tview.Primitive {
	feeder := &cmdDataFeeder{
		buffer:    new(bytes.Buffer),
		refresher: volumeRefresher,
	}

	t := newTableView()
	t.init(app, volume, feeder, defaultAction)

	if err := t.run(); err != nil {
		return t.updateStatus(err.Error(), true)
	}
	return t
}
