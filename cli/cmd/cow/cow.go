package cow

import (
	"fmt"
	"path/filepath"

	"github.com/gdamore/tcell"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/version"
	"github.com/rivo/tview"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type resourceKind struct {
	title string
	kind  string
}

type Cow struct{}

func (c *Cow) Run(ctx *clicontext.CLIContext) error {
	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}
	project, err := ctx.Project()
	if err != nil {
		return err
	}

	app := newAppview(ctx, cluster, project)
	if err := app.init(); err != nil {
		return err
	}
	return app.Run()
}

type drawPrimitive func() tview.Primitive

type appView struct {
	*tview.Application
	rioVersion  string
	k8sVersion  string
	menuView    menuView
	footerView  footerView
	statusView  statusView
	content     contentView
	pageRows    map[string]position
	pageMap     map[string]drawPrimitive
	showMenu    bool
	currentPage string
	clicontext  *clicontext.CLIContext
	cluster     *clientcfg.Cluster
	project     *clientcfg.Project
	syncs       map[string]chan struct{}
}

type position struct {
	row, column int
}

func newAppview(cliContext *clicontext.CLIContext, cluster *clientcfg.Cluster, project *clientcfg.Project) *appView {
	v := &appView{Application: tview.NewApplication()}
	{
		v.menuView = menuView{appView: v, Flex: tview.NewFlex()}
		v.content = contentView{appView: v, Pages: tview.NewPages()}
		v.footerView = footerView{appView: v, TextView: tview.NewTextView()}
		v.statusView = statusView{appView: v, TextView: tview.NewTextView()}
		v.clicontext = cliContext
		v.cluster = cluster
		v.project = project
		v.pageRows = make(map[string]position)
		v.pageMap = map[string]drawPrimitive{
			stacktypes:   v.Stack,
			servicetypes: v.Service,
			volumetypes:  v.Volume,
			configtypes:  v.Config,
			routetypes:   v.Route,
		}

		{
			v.menuView.SetBackgroundColor(defaultBackGroundColor)
			v.content.Pages.SetBackgroundColor(defaultBackGroundColor)
			v.footerView.SetBackgroundColor(tcell.ColorDarkCyan)
		}
	}
	return v
}

func (a *appView) init() error {
	a.rioVersion = version.Version
	k8sversion, err := a.getK8sVersion()
	if err != nil {
		return err
	}
	a.k8sVersion = k8sversion
	a.menuView.init()
	a.footerView.init()
	a.statusView.init()
	a.content.init()

	a.setResourcePages()

	main := tview.NewFlex()
	{
		main.SetDirection(tview.FlexRow)
		main.AddItem(a.content, 0, 15, true)

		footer := tview.NewFlex()
		footer.AddItem(a.footerView, 0, 1, false)
		footer.AddItem(a.statusView, 0, 1, false)

		main.AddItem(footer, 1, 1, false)
	}

	a.SetRoot(main, true)
	return nil
}

const (
	stacktypes   = "stacks"
	servicetypes = "services"
	volumetypes  = "volumes"
	configtypes  = "configs"
	routetypes   = "routes"
)

/*
page 1: stack
page 2: service
page 3: Volume
page 4: config
page 5: route
*/
func (a *appView) setResourcePages() {
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case '1':
				a.footerView.TextView.Highlight(stacktypes).ScrollToHighlight()
				a.content.SwitchPage(stacktypes, a.pageMap[stacktypes]())
			case '2':
				a.footerView.TextView.Highlight(servicetypes).ScrollToHighlight()
				a.content.SwitchPage(servicetypes, a.pageMap[servicetypes]())
			case '3':
				a.footerView.TextView.Highlight(volumetypes).ScrollToHighlight()
				a.content.SwitchPage(volumetypes, a.pageMap[volumetypes]())
			case '4':
				a.footerView.TextView.Highlight(configtypes).ScrollToHighlight()
				a.content.SwitchPage(configtypes, a.pageMap[configtypes]())
			case '5':
				a.footerView.TextView.Highlight(routetypes).ScrollToHighlight()
				a.content.SwitchPage(routetypes, a.pageMap[routetypes]())
			case 'm':
				if !a.showMenu {
					newpage := tview.NewPages().AddPage("menu", a.pageMap[a.currentPage](), true, true).
						AddPage("menu-decor", center(a.menuView, 60, 15), true, true)
					a.content.SwitchPage(a.currentPage, newpage)
					a.showMenu = true
				} else {
					a.SwitchPage(a.currentPage, a.pageMap[a.currentPage]())
					a.showMenu = false
				}
			}
		case tcell.KeyEscape:
			a.SwitchPage(a.currentPage, a.pageMap[a.currentPage]())
		}
		return event
	})

	// default to stack pages
	a.footerView.TextView.Highlight(stacktypes).ScrollToHighlight()
	a.content.SwitchPage(stacktypes, a.Stack())
}

func (a *appView) menuDecor(page string, p tview.Primitive) {
	a.content.Pages.AddPage(page, p, true, true).AddPage("menu", center(a.menuView, 30, 20), true, true)
}

func (a *appView) getK8sVersion() (string, error) {
	kubeconfig := filepath.Join(a.clicontext.KubeconfigCache(), a.cluster.Checksum)
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return "", err
	}
	clientset := kubernetes.NewForConfigOrDie(restConfig)
	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return version.GitVersion, nil
}

func (a *appView) SwitchPage(page string, p tview.Primitive) {
	a.currentPage = page
	a.content.RemovePage(page)
	a.content.AddPage(page, p, true, true)
	a.content.SwitchToPage(page)
	a.SetFocus(p)
}

type menuView struct {
	*tview.Flex
	*appView
}

func (m *menuView) init() {
	{
		m.Flex.SetDirection(tview.FlexRow)
		m.Flex.SetBackgroundColor(tcell.ColorGray)
		m.Flex.AddItem(m.logoView(), 6, 1, false)
		m.Flex.AddItem(m.versionView(), 4, 1, false)
		m.Flex.AddItem(m.tipsView(), 12, 1, false)
	}
}

var logo = ` 
 ____  ___ ___  
|  _ \|_ _/ _ \ 
| |_) || | | | |
|  _ < | | |_| |
|_| \_\___\___/ 
                
`

func (m *menuView) logoView() *tview.TextView {
	t := tview.NewTextView()
	t.SetBackgroundColor(tcell.ColorGray)
	t.SetText(logo).SetTextColor(defaultBackGroundColor).SetTextAlign(tview.AlignCenter).SetBorderAttributes(tcell.AttrBold)
	return t
}

func (m *menuView) versionView() *tview.Table {
	t := tview.NewTable()
	t.SetBackgroundColor(tcell.ColorGray)
	t.SetBorder(true)
	t.SetTitle("Version")
	rioVersionHeader := tview.NewTableCell("Rio Version:").SetAlign(tview.AlignCenter).SetExpansion(2)
	rioVersionValue := tview.NewTableCell(m.rioVersion).SetTextColor(tcell.ColorPurple).SetAlign(tview.AlignCenter).SetExpansion(2)

	k8sVersionHeader := tview.NewTableCell("K8s Version:").SetAlign(tview.AlignCenter).SetExpansion(2)
	k8sVersionValue := tview.NewTableCell(m.k8sVersion).SetTextColor(tcell.ColorPurple).SetAlign(tview.AlignCenter).SetExpansion(2)

	t.SetCell(0, 0, rioVersionHeader)
	t.SetCell(0, 1, rioVersionValue)
	t.SetCell(1, 0, k8sVersionHeader)
	t.SetCell(1, 1, k8sVersionValue)
	return t
}

var shortcuts = [][]string{
	{"Key i", "Inspect"},
	{"Key e", "Edit"},
	{"Key l", "Logs"},
	{"Key x", "Exec"},
	{"Key n", "Create"},
	{"Key d", "Delete"},
}

func (m *menuView) tipsView() *tview.Table {
	t := tview.NewTable()
	t.SetBorderPadding(1, 0, 0, 0)
	t.SetBackgroundColor(tcell.ColorGray)
	t.SetBorder(true)
	t.SetTitle("Shortcuts")
	var row int
	for _, values := range shortcuts {
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

type statusView struct {
	*tview.TextView
	*appView
}

func (s *statusView) init() {
	s.TextView.SetBackgroundColor(tcell.ColorGray)
}

type resourceView struct {
	tview.Primitive
	*appView
	kind  string
	index int
}

var footers = []resourceView{
	{
		kind:  stacktypes,
		index: 1,
	},
	{
		kind:  servicetypes,
		index: 2,
	},
	{
		kind:  volumetypes,
		index: 3,
	},
	{
		kind:  configtypes,
		index: 4,
	},
	{
		kind:  routetypes,
		index: 5,
	},
}

type footerView struct {
	*tview.TextView
	*appView
}

func (f *footerView) init() {
	f.TextView.
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).SetBackgroundColor(tcell.ColorGray)
	for index, t := range footers {
		fmt.Fprintf(f.TextView, `%d ["%s"][black]%s[white][""] `, index+1, t.kind, t.kind)
	}
}

type contentView struct {
	*tview.Pages
	*appView
}

func (c *contentView) init() {}

var center = func(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, false).
		AddItem(nil, 0, 1, false)
}
