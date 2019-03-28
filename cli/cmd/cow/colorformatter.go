package cow

import (
	"sort"

	"github.com/gdamore/tcell"

	"github.com/alecthomas/chroma/styles"
)

var defaultBackGroundColor = tcell.ColorBlack

var colorStyles []string

func init() {
	for style := range styles.Registry {
		colorStyles = append(colorStyles, style)
	}
	sort.Strings(colorStyles)
}
