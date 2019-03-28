package cow

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/chroma/styles"
	"github.com/gdamore/tcell"
	"github.com/rancher/rio/cli/cmd/ps"
	"github.com/rivo/tview"
)

const (
	defaultStyle   = "native"
	errorDelayTime = 1
)

type tableView struct {
	*tview.Table

	app          *appView
	data         []interface{}
	dataSource   DataSource
	projectName  string
	lock         sync.Mutex
	sync         <-chan struct{}
	actions      []action
	resourceKind resourceKind
}

type action struct {
	name        string
	description string
	shortcut    rune
}

type Row []string

type DataSource interface {
	Data() []Row
	Header() Row
	Refresh() error
}

var defaultAction = []action{
	{
		name:        "inspect",
		shortcut:    'i',
		description: "inspect a resource",
	},
	{
		name:        "edit",
		shortcut:    'e',
		description: "edit a resource",
	},
	{
		name:        "create",
		shortcut:    'c',
		description: "create a resource",
	},
	{
		name:        "delete",
		shortcut:    'd',
		description: "delete a resource",
	},
}

func newTableView() *tableView {
	return &tableView{
		Table: tview.NewTable(),
	}
}

func (t *tableView) init(app *appView, resource resourceKind, dataFeeder DataSource, actions []action) {
	{
		t.app = app
		t.projectName = app.project.Project.Name
		t.resourceKind = resource
		t.dataSource = dataFeeder
		t.sync = app.syncs[resource.kind]
		t.actions = actions
	}
	{
		t.Table.SetBorder(true)
		t.Table.SetBackgroundColor(defaultBackGroundColor)
		t.Table.SetBorderAttributes(tcell.AttrBold)
		t.Table.SetSelectable(true, false)
		t.setTitle("")
	}
	if p, ok := t.app.pageRows[t.resourceKind.kind]; ok {
		t.Table.Select(p.row, p.column)
	}

	actionMap := map[rune]action{}
	for _, a := range t.actions {
		actionMap[a.shortcut] = a
	}
	t.Table.SetSelectionChangedFunc(func(row, column int) {
		t.app.pageRows[t.resourceKind.kind] = position{
			row:    row,
			column: column,
		}
	})

	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			t.actionView()
		case tcell.KeyRune:
			switch event.Rune() {
			case 'i':
				t.inspect("yaml", defaultStyle, false)
			case 'l':
				t.logs("")
			case 'x':
				t.exec("")
			case 'd':
				t.rm()
			}
		}
		return event
	})
}

func (t *tableView) run() error {
	go func() {
		for {
			select {
			case <-t.app.clicontext.Ctx.Done():
				break
			case <-t.sync:
				if err := t.refresh(); err != nil {
					t.updateStatus(err.Error(), true)
				}
			}
		}
	}()

	return t.refresh()
}

func (t *tableView) getSelectedName() string {
	row, _ := t.Table.GetSelection()
	cell := t.Table.GetCell(row, 0)

	return strings.SplitN(cell.Text, " ", 2)[0]
}

func (t *tableView) inspect(format, style string, colorful bool) {
	name := t.getSelectedName()
	outBuffer := &strings.Builder{}
	errbuffer := &strings.Builder{}
	args := []string{"inspect", "-t", t.resourceKind.kind, name}
	cmd := exec.Command("rio", args...)
	cmd.Stdout = outBuffer
	cmd.Stderr = errbuffer
	if err := cmd.Run(); err != nil {
		t.updateStatus(errbuffer.String(), true)
		return
	}

	inspectBox := tview.NewTextView()
	if colorful {
		inspectBox.SetDynamicColors(true).SetBackgroundColor(tcell.Color(styles.Registry[style].Get(chroma.Background).Background))
		writer := tview.ANSIWriter(inspectBox)
		if err := quick.Highlight(writer, outBuffer.String(), format, "terminal256", style); err != nil {
			t.updateStatus(err.Error(), true)
			return
		}
	} else {
		inspectBox.SetDynamicColors(true).SetBackgroundColor(defaultBackGroundColor)
		inspectBox.SetText(outBuffer.String())
	}

	newpage := tview.NewPages().AddPage("inspect", inspectBox, true, true)
	t.app.SwitchPage(t.app.currentPage, newpage)
}

func (t *tableView) logs(container string) {
	name := t.getSelectedName()
	args := []string{"logs", "-f"}
	if container != "" {
		args = append(args, "-c", container)
	}
	args = append(args, name)
	cmd := exec.Command("rio", args...)

	logbox := tview.NewTextView()
	{
		logbox.SetTitle(fmt.Sprintf("Logs - %s - (%s)", name, t.projectName))
		logbox.SetBorder(true)
		logbox.SetTitleColor(tcell.ColorPurple)
		logbox.SetDynamicColors(true)
		logbox.SetBackgroundColor(defaultBackGroundColor)
		logbox.SetChangedFunc(func() {
			logbox.ScrollToEnd()
			t.app.Draw()
		})
		logbox.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEscape {
				cmd.Process.Kill()
			}
		})
	}

	cmd.Stdout = tview.ANSIWriter(logbox)
	go func() {
		if err := cmd.Run(); err != nil {
			return
		}
	}()

	newpage := tview.NewPages().AddPage("logs", logbox, true, true)
	t.app.SwitchPage(t.app.currentPage, newpage)
}

func (t *tableView) rm() {
	name := t.getSelectedName()
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Do you want to delete %s %s?", t.resourceKind, name)).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				args := []string{"rm", "-t", t.resourceKind.kind, name}
				cmd := exec.Command("rio", args...)
				if err := cmd.Run(); err != nil {
					t.updateStatus(err.Error(), true)
					return
				}
				t.switchToRootPage()
			}
		})
	t.embeddedDialog("delete", t.app.pageMap[t.app.currentPage](), modal)
}

func (t *tableView) switchToRootPage() {
	t.app.SwitchPage(t.app.currentPage, t.app.pageMap[t.app.currentPage]())
	t.app.Application.Draw()
}

func (t *tableView) exec(container string) {
	name := t.getSelectedName()
	shellArgs := []string{"/bin/sh", "-c", "TERM=xterm-256color; export TERM; [ -x /bin/bash ] && ([ -x /usr/bin/script ] && /usr/bin/script -q -c /bin/bash /dev/null || exec /bin/bash) || exec /bin/sh"}
	args := []string{"exec", "-it"}
	if container != "" {
		args = append(args, "-c", container)
	}
	args = append(args, name)
	cmd := exec.Command("rio", append(args, shellArgs...)...)
	errorBuffer := &strings.Builder{}
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, errorBuffer, os.Stdin

	t.app.Suspend(func() {
		clearScreen()
		if err := cmd.Run(); err != nil {
			t.updateStatus(errorBuffer.String(), true)
		}
		return
	})
}

func (t *tableView) actionView() {
	list := newSelectionList()
	list.SetSelectedFunc(func(i int, s string, s2 string, r rune) {
		switch s {
		case "inspect":
			formatList := newSelectionList()
			formatList.SetSelectedFunc(func(i int, s string, s2 string, r rune) {
				if s2 == "json format" || s2 == "yaml format" {
					t.inspect(s, defaultStyle, false)
				} else {
					colorList := newSelectionList()
					for _, s := range colorStyles {
						colorList.AddItem(s, "", ' ', nil)
					}
					colorList.SetSelectedFunc(func(i int, style string, s2 string, r rune) {
						t.inspect(s, style, true)
					})
					t.embeddedDialog("inspect-color", t.app.pageMap[t.app.currentPage](), colorList)
				}
			})
			formatList.AddItem("yaml", "yaml format", 'y', nil)
			formatList.AddItem("yaml", "yaml with different color styles", 'u', nil)
			formatList.AddItem("json", "json format", 'j', nil)
			formatList.AddItem("json", "json with different color styles", 'k', nil)
			t.embeddedDialog("inspect-format", t.app.pageMap[t.app.currentPage](), formatList)
		case "log":
			list := t.newContainerSelectionList()
			list.SetSelectedFunc(func(i int, s string, s2 string, r rune) {
				t.logs(s)
			})
			t.embeddedDialog("logs-containers", t.app.pageMap[t.app.currentPage](), list)
		case "exec":
			list := t.newContainerSelectionList()
			list.SetSelectedFunc(func(i int, s string, s2 string, r rune) {
				t.exec(s)
			})
			t.embeddedDialog("exec-containers", t.app.pageMap[t.app.currentPage](), list)
		case "delete":
			t.rm()
		}
	})

	for _, a := range t.actions {
		list.AddItem(a.name, a.description, a.shortcut, nil)
	}
	t.embeddedDialog("option", t.app.pageMap[t.app.currentPage](), list)
}

func (t *tableView) newContainerSelectionList() *tview.List {
	list := newSelectionList()
	containers, err := t.listContainer()
	if err != nil {
		t.updateStatus(err.Error(), true)
		return nil
	}
	for _, c := range containers {
		list.AddItem(c, "", ' ', nil)
	}
	return list
}

func (t *tableView) listContainer() ([]string, error) {
	svcName := t.getSelectedName()
	pods, err := ps.ListPods(t.app.clicontext, true, svcName)
	if err != nil {
		return nil, err
	}
	var containers []string
	for _, p := range pods {
		for _, c := range p.Containers {
			containers = append(containers, c.Name)
		}

	}
	return containers, nil
}

func newSelectionList() *tview.List {
	list := tview.NewList()
	list.SetBackgroundColor(tcell.ColorGray)
	list.SetMainTextColor(tcell.ColorBlack)
	list.SetSecondaryTextColor(tcell.ColorPurple)
	list.SetShortcutColor(defaultBackGroundColor)
	return list
}

func (t *tableView) refresh() error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if err := t.dataSource.Refresh(); err != nil {
		return err
	}

	t.drawTable()
	return nil
}

func (t *tableView) drawTable() {
	t.Clear()

	header := t.dataSource.Header()
	data := t.dataSource.Data()

	for col, name := range header {
		t.addHeaderCell(col, name)
	}

	for r, row := range data {
		if len(row) > 0 && row[0] == "" {
			continue
		}
		for col, value := range row {
			t.addBodyCell(r, col, value)
		}
	}
}

func (t *tableView) addHeaderCell(col int, name string) {
	c := tview.NewTableCell(fmt.Sprintf("[white]%s", name)).SetSelectable(false)
	{
		c.SetExpansion(3)
		c.SetTextColor(tcell.ColorAntiqueWhite)
		c.SetAttributes(tcell.AttrBold)
	}
	t.SetCell(0, col, c)
}

func (t *tableView) addBodyCell(row, col int, value string) {
	c := tview.NewTableCell(fmt.Sprintf("%s", value))
	{
		c.SetExpansion(10)
		c.SetTextColor(tcell.ColorAntiqueWhite)
	}
	t.SetCell(row+1, col, c)
}

func (t *tableView) setTitle(status string) {
	title := fmt.Sprintf("%s - [white](%s) %s", t.resourceKind.title, t.projectName, status)
	t.Table.SetTitle(title)
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func (t *tableView) embeddedDialog(name string, page tview.Primitive, dialog tview.Primitive) {
	newpage := tview.NewPages().AddPage(name, t.app.pageMap[t.app.currentPage](), true, true).
		AddPage("dialog", center(dialog, 40, 15), true, true)
	t.app.SwitchPage(t.app.currentPage, newpage)
	t.app.SetFocus(dialog)
}

func (t *tableView) updateStatus(status string, isError bool) *tableView {
	statusBar := tview.NewTextView()
	statusBar.SetBorder(true)
	statusBar.SetBorderAttributes(tcell.AttrBold)
	statusBar.SetBorderPadding(1, 1, 1, 1)
	if isError {
		statusBar.SetTitle("Error")
		statusBar.SetTitleColor(tcell.ColorRed)
		statusBar.SetTextColor(tcell.ColorRed)
		statusBar.SetBorderColor(tcell.ColorRed)
	} else {
		statusBar.SetTitle("Progress")
		statusBar.SetTitleColor(tcell.ColorYellow)
		statusBar.SetTextColor(tcell.ColorYellow)
		statusBar.SetBorderColor(tcell.ColorYellow)
	}
	statusBar.SetText(status)
	statusBar.SetTextAlign(tview.AlignCenter)
	newpage := tview.NewPages().
		AddPage("status", t.app.pageMap[t.app.currentPage](), true, true).
		AddPage("dialog", center(statusBar, 100, 5), true, true)
	t.app.SwitchPage(t.app.currentPage, newpage)

	if isError {
		go func() {
			time.Sleep(time.Second * errorDelayTime)
			t.switchToRootPage()
		}()
	}
	return t
}
