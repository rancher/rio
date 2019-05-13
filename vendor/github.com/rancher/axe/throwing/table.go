package throwing

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rancher/axe/throwing/datafeeder"
	"github.com/rancher/axe/throwing/types"
	"github.com/rivo/tview"
	"golang.org/x/net/context"
	"k8s.io/client-go/kubernetes"
)

const (
	errorDelayTime = 1
)

type TableView struct {
	*tview.Table

	drawer       types.Drawer
	navigateMap  map[rune]string
	client       *kubernetes.Clientset
	app          *AppView
	data         []interface{}
	dataSource   datafeeder.DataSource
	lock         sync.Mutex
	sync         chan struct{}
	actions      []types.Action
	resourceKind types.ResourceKind
	search       string
}

type EventHandler func(t *TableView) func(event *tcell.EventKey) *tcell.EventKey

func NewTableView(app *AppView, kind string, drawer types.Drawer) *TableView {
	view := drawer.ViewMap[kind]
	t := &TableView{
		Table:  tview.NewTable(),
		drawer: drawer,
	}
	t.init(app, view.Kind, view.Feeder, view.Actions, drawer.PageNav, nil)
	if err := t.refresh(); err != nil {
		return t.UpdateStatus(err.Error(), true).(*TableView)
	}
	return t
}

func (t *TableView) NewNestTableView(kind types.ResourceKind, feeder datafeeder.DataSource, actions []types.Action, pageNav map[rune]string, embeddedHandler EventHandler) *TableView {
	nt := &TableView{
		Table:  tview.NewTable(),
		drawer: t.drawer,
	}
	nt.init(t.app, kind, feeder, actions, pageNav, embeddedHandler)
	if err := nt.refresh(); err != nil {
		return t.UpdateStatus(err.Error(), true).(*TableView)
	}
	return nt
}

func (t *TableView) init(app *AppView, resource types.ResourceKind, dataFeeder datafeeder.DataSource, actions []types.Action, pageNav map[rune]string, embeddedHandler EventHandler) {
	{
		t.app = app
		t.resourceKind = resource
		t.dataSource = dataFeeder
		t.sync = app.syncs[resource.Kind]
		t.actions = actions
		t.client = app.clientset
		t.navigateMap = pageNav
	}
	{
		t.Table.SetBorder(true)
		t.Table.SetBackgroundColor(tcell.ColorBlack)
		t.Table.SetBorderAttributes(tcell.AttrBold)
		t.Table.SetSelectable(true, false)
		t.Table.SetTitle(t.resourceKind.Title)
	}
	if t.sync == nil {
		t.sync = make(chan struct{}, 0)
	}

	if p, ok := t.app.pageRows[t.resourceKind.Kind]; ok {
		t.Table.Select(p.row, p.column)
	}

	actionMap := map[rune]types.Action{}
	for _, a := range t.actions {
		actionMap[a.Shortcut] = a
	}

	// todo: this needs to be changed to rowID to track selection if refresh happens
	t.Table.SetSelectionChangedFunc(func(row, column int) {
		t.app.pageRows[t.resourceKind.Kind] = position{
			row:    row,
			column: column,
		}
	})

	if embeddedHandler != nil {
		t.SetInputCapture(embeddedHandler(t))
		return
	}

	if app.handler != nil {
		t.SetInputCapture(app.handler(t))
	}

	go func() {
		t.run(app.context)
	}()
}

func (t *TableView) run(ctx context.Context) {
	for {
		select {
		case <-t.sync:
			if t.resourceKind.Kind != t.app.currentPage {
				continue
			}
			if err := t.refresh(); err != nil {
				t.UpdateStatus(err.Error(), true)
			}
			t.SwitchPage(t.app.currentPage, t.app.tableViews[t.app.currentPage])
		case <-ctx.Done():
			return
		}
	}
}

func (t *TableView) GetSelectionName() string {
	row, _ := t.Table.GetSelection()
	cell := t.Table.GetCell(row, 0)

	return strings.SplitN(cell.Text, " ", 2)[0]
}

func (t *TableView) SwitchToRootPage() {
	t.app.SwitchToRootPage()
}

func (t *TableView) refresh() error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if err := t.dataSource.Refresh(); err != nil {
		return err
	}
	t.draw()
	return nil
}

func (t *TableView) draw() {
	t.Clear()

	header := t.dataSource.Header()
	data := t.dataSource.Data()

	nameRow := 0
	for col, name := range header {
		if name == "NAME" {
			nameRow = col
		}
		t.addHeaderCell(col, name)
	}

	r := 0
	for _, row := range data {
		if len(row) > 0 && row[0] == "" {
			continue
		}
		if t.search != "" && !strings.Contains(row[nameRow], t.search) {
			continue
		}
		for col, value := range row {
			t.addBodyCell(r, col, value)
		}
		r++
	}
	if t.search != "" {
		t.search = ""
	}
	t.GetApplication().Draw()
}

func (t *TableView) addHeaderCell(col int, name string) {
	c := tview.NewTableCell(fmt.Sprintf("[white]%s", name)).SetSelectable(false)
	{
		c.SetExpansion(1)
		c.SetTextColor(tcell.ColorAntiqueWhite)
		c.SetAttributes(tcell.AttrBold)
	}
	t.Table.SetCell(0, col, c)
}

func (t *TableView) addBodyCell(row, col int, value string) {
	c := tview.NewTableCell(fmt.Sprintf("%s", value))
	{
		c.SetExpansion(1)
		c.SetTextColor(tcell.ColorAntiqueWhite)
	}
	t.Table.SetCell(row+1, col, c)
}

func (t *TableView) InsertDialog(name string, page tview.Primitive, dialog tview.Primitive) {
	newpage := tview.NewPages()
	newpage.AddPage(name, page, true, true).
		AddPage("dialog", center(dialog, 50, 20), true, true)
	t.app.SwitchPage(t.app.currentPage, newpage)
	t.app.Application.SetFocus(dialog)
}

func (t *TableView) UpdateStatus(status string, isError bool) tview.Primitive {
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
	newpage := tview.NewPages()
	if _, ok := t.app.tableViews[t.app.currentPage]; ok {
		newpage.AddPage("handler", t.app.currentPrimitive, true, true)
	}
	newpage.AddPage("dialog", center(statusBar, 100, 5), true, true)
	t.app.SwitchPage(t.app.currentPage, newpage)

	go func() {
		time.Sleep(time.Second * errorDelayTime)
		t.SwitchToRootPage()
		t.GetApplication().Draw()
	}()
	return t
}

func (t *TableView) GetClientSet() *kubernetes.Clientset {
	return t.client
}

func (t *TableView) GetResourceKind() string {
	return t.resourceKind.Kind
}

func (t *TableView) GetCurrentPage() string {
	return t.app.currentPage
}

func (t *TableView) GetAction() []types.Action {
	return t.actions
}

func (t *TableView) GetApplication() *tview.Application {
	return t.app.Application
}

func (t *TableView) UpdateFeeder(kind string, feeder datafeeder.DataSource) {
	tableview := t.app.tableViews[kind]
	tableview.dataSource = feeder
}

func (t *TableView) GetCurrentPrimitive() tview.Primitive {
	if t.app.drawQueue.Empty() {
		return t.app.tableViews[t.drawer.RootPage]
	}
	return t.app.drawQueue.Last()
}

func (t *TableView) SwitchPage(page string, draw tview.Primitive) {
	t.app.SwitchPage(page, draw)
}

func (t *TableView) SetCurrentPage(page string) {
	t.app.currentPage = page
}

func (t *TableView) GetTable() *tview.Table {
	return t.Table
}

func (t *TableView) GetTableView(kind string) *TableView {
	if _, ok := t.app.tableViews[kind]; !ok {
		t.app.tableViews[kind] = NewTableView(t.app, kind, t.drawer)
	}
	return t.app.tableViews[kind]
}

func (t *TableView) BackPage() {
	t.app.LastPage()
}

func (t *TableView) Refresh() {
	go func() {
		t.sync <- struct{}{}
	}()
}

func (t *TableView) RefreshManual() {
	if err := t.refresh(); err != nil {
		t.UpdateStatus(err.Error(), true)
	}
}

func (t *TableView) UpdateWithSearch(search string) {
	t.search = search
}

func (t *TableView) ShowMenu() {
	app := t.app
	if !app.showMenu {
		newpage := tview.NewPages().AddPage("menu", app.CurrentPage(), true, true).
			AddPage("menu-decor", center(app.menuView, 60, 35), true, true)
		app.SwitchPage(app.currentPage, newpage)
		app.SetFocus(app.menuView)
		app.showMenu = true
	}
}

func (t *TableView) ShowSearch() {
	t.app.SetFocus(t.app.searchView.InputField)
}

func (t *TableView) Navigate(r rune) {
	app := t.app
	if kind, ok := t.navigateMap[r]; ok {
		app.footerView.TextView.Highlight(kind).ScrollToHighlight()
		if _, ok := app.tableViews[kind]; !ok {
			app.tableViews[kind] = NewTableView(app, kind, t.drawer)
		}
		app.SwitchPage(kind, app.tableViews[kind])
	}
}

func (t *TableView) RootPage() {
	t.SwitchPage(t.drawer.RootPage, t.app.tableViews[t.drawer.RootPage])
	t.app.tableViews[t.drawer.RootPage].Refresh()
}

func (t *TableView) LastPage() {
	t.app.LastPage()
}

func (t *TableView) GetNestedTable(kind string) *TableView {
	return t.app.tableViews[kind]
}

func (t *TableView) SetTableView(kind string, nt *TableView) {
	t.app.tableViews[kind] = nt
}
