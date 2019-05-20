package throwing

import (
	"context"
	"fmt"
	"sync"

	"github.com/gdamore/tcell"
	"github.com/rancher/axe/throwing/types"
	"github.com/rivo/tview"
	"k8s.io/client-go/kubernetes"
)

var logo = ` 
    _              
   / \   __  _____ 
  / _ \  \ \/ / _ \
 / ___ \  >  <  __/
/_/   \_\/_/\_\___|
`

type AppView struct {
	*tview.Flex
	*tview.Application
	types.Drawer
	handler          EventHandler
	context          context.Context
	cancel           context.CancelFunc
	version          string
	k8sVersion       string
	clientset        *kubernetes.Clientset
	menuView         menuView
	footerView       footerView
	searchView       cmdView
	content          contentView
	drawQueue        *PrimitiveQueue
	tableViews       map[string]*TableView
	pageRows         map[string]position
	showMenu         bool
	currentPage      string
	currentPrimitive *TableView
	switchPage       chan struct{}
	syncs            map[string]chan struct{}
	lock             sync.Mutex
}

type position struct {
	row, column int
}

/*
NewAppView takes 4 parameters:
	Clientset: Kubernetes client
	Drawer: Generic drawer to define how the table view looks like
	Handler: Event handler
	RefresherSignals: External Signal to trigger table refresh, mapped by resource kind
*/
func NewAppView(clientset *kubernetes.Clientset, dr types.Drawer, handler EventHandler, refreshSignals map[string]chan struct{}) *AppView {
	v := &AppView{Application: tview.NewApplication()}
	{
		v.Flex = tview.NewFlex()
		v.drawQueue = &PrimitiveQueue{AppView: v}
		v.menuView = menuView{AppView: v, TextView: tview.NewTextView()}
		v.content = contentView{AppView: v, Pages: tview.NewPages()}
		v.footerView = footerView{AppView: v, TextView: tview.NewTextView()}
		v.searchView = cmdView{AppView: v, InputField: tview.NewInputField()}
		v.pageRows = make(map[string]position)
		v.clientset = clientset
		v.Drawer = dr
		v.handler = handler
		v.syncs = refreshSignals

		{
			v.menuView.SetBackgroundColor(tcell.ColorBlack)
			v.content.Pages.SetBackgroundColor(tcell.ColorBlack)
			v.footerView.SetBackgroundColor(tcell.ColorDarkCyan)
		}
	}
	return v
}

func (app *AppView) Init() error {
	k8sversion, err := app.getK8sVersion()
	if err != nil {
		return err
	}
	app.context, app.cancel = context.WithCancel(context.Background())
	app.tableViews = map[string]*TableView{
		app.RootPage: NewTableView(app, app.RootPage, app.Drawer),
	}
	app.k8sVersion = k8sversion
	app.menuView.init()
	app.footerView.init()
	app.content.init()
	app.switchPage = make(chan struct{}, 1)

	// set default page to root page
	app.footerView.TextView.Highlight(app.RootPage).ScrollToHighlight()
	app.SwitchPage(app.RootPage, app.tableViews[app.RootPage])

	// Initialize after switching page so that it has context of current page to search for
	app.searchView.init()

	app.setInputHandler()

	//go app.watch()

	main := tview.NewFlex()
	{
		main.SetDirection(tview.FlexRow)
		main.AddItem(app.content, 0, 15, true)

		search := tview.NewFlex().SetDirection(tview.FlexRow)
		search.AddItem(app.searchView.InputField, 0, 1, true)

		footer := tview.NewFlex().SetDirection(tview.FlexColumn)
		footer.AddItem(app.footerView, 0, 1, false)
		footer.AddItem(app.menuView, 0, 1, false)

		main.AddItem(search, 1, 1, false)
		main.AddItem(footer, 1, 1, false)
	}

	app.Application.SetRoot(main, true)
	return nil
}

func (app *AppView) watch() {
	go app.currentPrimitive.run(app.context)
	for {
		select {
		case <-app.switchPage:
			app.cancel()
			// recreate context
			app.context, app.cancel = context.WithCancel(app.context)
			go app.currentPrimitive.run(app.context)
		}
	}
}

/*
setInputHandler setup the input event handler for main page

PageNav: Navigate different pages listed in footer
M(Menu): Menu view
Escape: go back to the previous view
*/
func (app *AppView) setInputHandler() {
	app.SetInputCapture(EscapeEventHandler(app))
}

func (app *AppView) menuDecor(page string, p tview.Primitive) {
	newpage := tview.NewPages()
	newpage.AddPage(page, p, true, true).AddPage("menu", center(app.menuView, 30, 20), true, true)
	app.SwitchPage(page, newpage)
}

func (app *AppView) getK8sVersion() (string, error) {
	ver, err := app.clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return ver.GitVersion, nil
}

func (app *AppView) SwitchPage(page string, p tview.Primitive) {
	if app.currentPage != page {
		cp := app.currentPage
		app.currentPage = page
		if _, ok := p.(*TableView); ok {
			app.currentPrimitive = p.(*TableView)
		}
		if cp != "" {
			go func() {
				app.switchPage <- struct{}{}
			}()
		}
	}
	app.content.AddAndSwitchToPage(page, p, true)

	app.drawQueue.Enqueue(PageTrack{
		PageName:  page,
		Primitive: p,
	})
	app.SetFocus(p)
}

func (app *AppView) SwitchToRootPage() {
	app.showMenu = false
	app.SwitchPage(app.currentPage, app.tableViews[app.currentPage])
}

func (app *AppView) CurrentPage() tview.Primitive {
	return app.currentPrimitive
}

func (app *AppView) LastPage() {
	app.drawQueue.Dequeue()
	page := app.drawQueue.Last()
	app.SwitchPage(page.PageName, page.Primitive)
}

type menuView struct {
	*tview.TextView
	*AppView
}

func (m *menuView) init() {
	m.TextView.
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight).
		SetWrap(false).SetBackgroundColor(tcell.ColorGray)
	for _, action := range m.Menu {
		fmt.Fprintf(m.TextView, "[blue]%v[black] %s ", string(action.Shortcut), action.Name)
	}
}

func (m *menuView) logoView() *tview.TextView {
	t := tview.NewTextView()
	t.SetBackgroundColor(tcell.ColorGray)
	t.SetText(logo).SetTextColor(tcell.ColorBlack).SetTextAlign(tview.AlignCenter).SetBorderAttributes(tcell.AttrBold)
	return t
}

func (m *menuView) versionView() *tview.Table {
	t := tview.NewTable()
	t.SetBackgroundColor(tcell.ColorGray)
	t.SetBorder(true)
	t.SetTitle("Version")
	rioVersionHeader := tview.NewTableCell("Axe Version:").SetAlign(tview.AlignCenter).SetExpansion(2)
	rioVersionValue := tview.NewTableCell(m.version).SetTextColor(tcell.ColorPurple).SetAlign(tview.AlignCenter).SetExpansion(2)

	k8sVersionHeader := tview.NewTableCell("K8s Version:").SetAlign(tview.AlignCenter).SetExpansion(2)
	k8sVersionValue := tview.NewTableCell(m.k8sVersion).SetTextColor(tcell.ColorPurple).SetAlign(tview.AlignCenter).SetExpansion(2)

	t.SetCell(0, 0, rioVersionHeader)
	t.SetCell(0, 1, rioVersionValue)
	t.SetCell(1, 0, k8sVersionHeader)
	t.SetCell(1, 1, k8sVersionValue)
	return t
}

func (m *menuView) tipsView() *tview.Table {
	t := tview.NewTable()
	t.SetBorderPadding(1, 0, 0, 0)
	t.SetBackgroundColor(tcell.ColorGray)
	t.SetBorder(true)
	t.SetTitle("Shortcuts")
	var row int
	for _, values := range m.Shortcuts {
		kc, vc := newKeyValueCell(values[0], values[1])
		t.SetCell(row, 0, kc)
		t.SetCell(row, 1, vc)
		row++
	}
	return t
}

func newKeyValueCell(key, value string) (*tview.TableCell, *tview.TableCell) {
	keycell := tview.NewTableCell(key).SetAlign(tview.AlignCenter).SetExpansion(2)
	valuecell := tview.NewTableCell(value).SetTextColor(tcell.ColorPurple).SetAlign(tview.AlignCenter).SetExpansion(2)
	return keycell, valuecell
}

type cmdView struct {
	*tview.InputField
	*AppView
}

func (s *cmdView) init() {
	s.InputField.SetFieldBackgroundColor(tcell.ColorBlack)
	s.InputField.SetFieldTextColor(tcell.ColorBlue)
	s.InputField.SetDoneFunc(searchDoneEventHandler(s.AppView))
}

type footerView struct {
	*tview.TextView
	*AppView
}

func (f *footerView) init() {
	f.TextView.
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).SetBackgroundColor(tcell.ColorGray)
	for index, t := range f.Footers {
		fmt.Fprintf(f.TextView, `%d ["%s"][black]%s[white][""] `, index+1, t.Kind, t.Title)
	}
}

type contentView struct {
	*tview.Pages
	*AppView
}

func (c *contentView) init() {
}

var center = func(p tview.Primitive, width, height int) tview.Primitive {
	newflex := tview.NewFlex()
	newflex.
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, false).
		AddItem(nil, 0, 1, false)
	return newflex
}
