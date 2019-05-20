package throwing

import (
	"github.com/gdamore/tcell"
)

var (
	EscapeEventHandler = func(app *AppView) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
				app.showMenu = false
				app.SwitchPage(app.currentPage, app.tableViews[app.currentPage])
			}
			return event
		}
	}

	searchDoneEventHandler = func(app *AppView) func(key tcell.Key) {
		return func(key tcell.Key) {
			switch key {
			case tcell.KeyEscape:
				app.SetFocus(app.content)
				app.searchView.InputField.SetText("")
			case tcell.KeyEnter:
				t := app.tableViews[app.currentPage]
				t.UpdateWithSearch(app.searchView.InputField.GetText())
				app.searchView.InputField.SetText("")
				t.Refresh()
			}
		}
	}
)
