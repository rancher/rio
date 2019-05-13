package tui

import (
	"github.com/gdamore/tcell"
	"github.com/rancher/axe/throwing"
	"github.com/rancher/axe/throwing/datafeeder"
	"github.com/rancher/axe/throwing/types"
)

const (
	serviceKind         = "service"
	routeKind           = "router"
	appKind             = "app"
	podKind             = "pod"
	configKind          = "config"
	publicdomainKind    = "publicdomain"
	externalServiceKind = "externalservice"
)

var (
	defaultBackGroundColor = tcell.ColorBlack

	colorStyles []string

	RootPage = appKind

	Shortcuts = [][]string{
		// CRUD
		{"Key c", "Create"},
		{"Key i", "Inspect"},
		{"Key e", "Edit"},
		{"Key d", "Delete"},

		// exec and log
		{"Key l", "Logs"},
		{"Key x", "Exec"},

		// view pods and revisions
		{"Key p", "View Pods"},
		{"Key v", "View revision"},

		{"Key /", "Search"},
		{"Key Ctrl+h", "Hit Endpoint"},
		{"Key Ctrl+r", "Refresh"},
		{"Key Ctrl+s", "Show system resource"},
	}

	Footers = []types.ResourceView{
		{
			Title: "Apps",
			Kind:  appKind,
			Index: 1,
		},
		{
			Title: "Routes",
			Kind:  routeKind,
			Index: 2,
		},
		{
			Title: "ExternalService",
			Kind:  externalServiceKind,
			Index: 3,
		},
		{
			Title: "PublicDomain",
			Kind:  publicdomainKind,
			Index: 4,
		},
		{
			Title: "Config",
			Kind:  configKind,
			Index: 5,
		},
	}

	PageNav = map[rune]string{
		'1': appKind,
		'2': routeKind,
		'3': externalServiceKind,
		'4': publicdomainKind,
		'5': configKind,
	}

	tableEventHandler = func(t *throwing.TableView) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEscape:
				kind := t.GetResourceKind()
				if kind == serviceKind || kind == podKind {
					kind = appKind
				}
				t.GetTableView(kind).RefreshManual()
				t.SwitchPage(kind, t.GetTableView(kind))
			case tcell.KeyEnter:
				actionView(t, false)
			case tcell.KeyCtrlR:
				t.Refresh()
			case tcell.KeyCtrlS:
				showSystem = !showSystem
				t.Refresh()
			case tcell.KeyCtrlH:
				hit(t)
			case tcell.KeyCtrlP:
				promote(t)
			case tcell.KeyRune:
				switch event.Rune() {
				case 'i':
					inspect("yaml", t)
				case 'l':
					logs("", t)
				case 'x':
					execute("", t)
				case 'd':
					rm(t)
				case '/':
					t.ShowSearch()
				case 'm', 'h', '?':
					t.ShowMenu()
				default:
					t.Navigate(event.Rune())
				case 'p':
					viewPods(t)
				case 'e':
					edit(t)
				case 'v':
					revisions(t)
				}
			}
			return event
		}
	}

	App = types.ResourceKind{
		Title: "Apps",
		Kind:  appKind,
	}

	Route = types.ResourceKind{
		Title: "Routers",
		Kind:  routeKind,
	}

	Config = types.ResourceKind{
		Title: "Configs",
		Kind:  configKind,
	}

	PublicDomain = types.ResourceKind{
		Title: "PublicDomains",
		Kind:  publicdomainKind,
	}

	Service = types.ResourceKind{
		Title: "Services",
		Kind:  serviceKind,
	}

	ExternalService = types.ResourceKind{
		Title: "ExternalServices",
		Kind:  externalServiceKind,
	}

	Pod = types.ResourceKind{
		Title: "Pods",
		Kind:  podKind,
	}

	DefaultAction = []types.Action{
		{
			Name:        "inspect",
			Shortcut:    'i',
			Description: "inspect a resource",
		},
		{
			Name:        "edit",
			Shortcut:    'e',
			Description: "edit a resource",
		},
		{
			Name:        "create",
			Shortcut:    'c',
			Description: "create a resource",
		},
		{
			Name:        "delete",
			Shortcut:    'd',
			Description: "delete a resource",
		},
	}

	podAction = []types.Action{
		{
			Name:        "pods",
			Shortcut:    'p',
			Description: "view pods of a service or app",
		},
	}

	revisionAction = []types.Action{
		{
			Name:        "revisions",
			Shortcut:    'v',
			Description: "view revisions of a app",
		},
	}

	execAndlog = []types.Action{
		{
			Name:        "exec",
			Shortcut:    'x',
			Description: "exec into a container or service",
		},
		{
			Name:        "log",
			Shortcut:    'l',
			Description: "view logs of a service",
		},
		{
			Name:        "hit",
			Shortcut:    'h',
			Description: "hit endpoint of a service(need jq and curl)",
		},
	}

	pods = append(DefaultAction, execAndlog...)

	services = append(DefaultAction, append(execAndlog, podAction...)...)

	apps = append(DefaultAction, append(execAndlog, append(podAction, revisionAction...)...)...)

	ViewMap = map[string]types.View{
		appKind: {
			Actions: apps,
			Kind:    App,
			Feeder:  datafeeder.NewDataFeeder(AppRefresher),
		},
		routeKind: {
			Actions: DefaultAction,
			Kind:    Route,
			Feeder:  datafeeder.NewDataFeeder(RouteRefresher),
		},
		externalServiceKind: {
			Actions: DefaultAction,
			Kind:    ExternalService,
			Feeder:  datafeeder.NewDataFeeder(ExternalRefresher),
		},
		configKind: {
			Actions: DefaultAction,
			Kind:    Config,
			Feeder:  datafeeder.NewDataFeeder(ConfigRefresher),
		},
		publicdomainKind: {
			Actions: DefaultAction,
			Kind:    PublicDomain,
			Feeder:  datafeeder.NewDataFeeder(PublicDomainRefresher),
		},
		serviceKind: {
			Actions: services,
			Kind:    Service,
			Feeder:  datafeeder.NewDataFeeder(ServiceRefresher),
		},
		podKind: {
			Actions: pods,
			Kind:    Pod,
			Feeder:  datafeeder.NewDataFeeder(PodRefresher),
		},
	}

	drawer = types.Drawer{
		RootPage:  RootPage,
		Shortcuts: Shortcuts,
		ViewMap:   ViewMap,
		PageNav:   PageNav,
		Footers:   Footers,
	}
)
