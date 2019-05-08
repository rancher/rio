package tui

import (
"github.com/rancher/axe/throwing"
"github.com/rancher/axe/throwing/datafeeder"
"github.com/rancher/axe/throwing/types"
)

const (
	serviceKind         = "services"
	routeKind           = "routes"
	podKind             = "pods"
	externalServiceKind = "externalservices"

	stackLabel   = "rio.cattle.io/stack"
	serviceLabel = "rio.cattle.io/service"
)

var (
	defaultBackGroundColor = tcell.ColorBlack

	colorStyles []string

	RootPage = serviceKind

	Shortcuts = [][]string{
		{"Key i", "Inspect"},
		{"Key e", "Edit"},
		{"Key l", "Logs"},
		{"Key x", "Exec"},
		{"Key n", "Create"},
		{"Key d", "Delete"},
		{"Key r", "Refresh"},
		{"Key /", "Search"},
		{"Key p", "View Pods"},
		{"Ket h", "Hit Endpoint"},
	}

	Footers = []types.ResourceView{
		{
			Title: "Services",
			Kind:  serviceKind,
			Index: 1,
		},
		{
			Title: "Routes",
			Kind:  routeKind,
			Index: 2,
		},
	}

	PageNav = map[rune]string{
		'1': serviceKind,
		'2': routeKind,
	}

	tableEventHandler = func(t *throwing.TableView) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				actionView(t)
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
				case 'r':
					t.Refresh()
				case '/':
					t.ShowSearch()
				case 'm':
					t.ShowMenu()
				default:
					t.Navigate(event.Rune())
				case 'p':
					viewPods(t)
				case 'h':
					hit(t)
				}
			}
			return event
		}
	}

	Route = types.ResourceKind{
		Title: "Routes",
		Kind:  "route",
	}

	Service = types.ResourceKind{
		Title: "Services",
		Kind:  "service",
	}

	ExternalService = types.ResourceKind{
		Kind:  "ExternalService",
		Title: "externalservice",
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

	ViewMap = map[string]types.View{
		serviceKind: {
			Actions: append(
				DefaultAction,
				append(
					execAndlog,
					types.Action{
						Name:        "pods",
						Shortcut:    'p',
						Description: "view pods of a service",
					})...,
			),
			Kind:   Service,
			Feeder: datafeeder.NewDataFeeder(ServiceRefresher),
		},
		routeKind: {
			Actions: DefaultAction,
			Kind:    Route,
			Feeder:  datafeeder.NewDataFeeder(RouteRefresher),
		},
		externalServiceKind: {
			Actions: DefaultAction,
			Kind:    Route,
			Feeder:  datafeeder.NewDataFeeder(RouteRefresher),
		},
	}

	drawer = types.Drawer{
		RootPage:  RootPage,
		Shortcuts: Shortcuts,
		ViewMap:   ViewMap,
		PageNav:   PageNav,
		Footers:   Footers,
	}
