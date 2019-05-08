package throwing

import (
	"github.com/rivo/tview"
)

type PageTrack struct {
	PageName string
	tview.Primitive
}

type PrimitiveQueue struct {
	*AppView
	items []PageTrack
}

func (p *PrimitiveQueue) Enqueue(t PageTrack) {
	p.items = append(p.items, t)
}

func (p *PrimitiveQueue) Dequeue() PageTrack {
	return p.last(true)
}

func (p *PrimitiveQueue) Empty() bool {
	return len(p.items) == 0
}

func (p *PrimitiveQueue) last(dequeue bool) PageTrack {
	if p.Empty() {
		return PageTrack{
			Primitive: p.AppView.tableViews[p.RootPage],
			PageName:  p.RootPage,
		}
	}
	item := p.items[len(p.items)-1]
	if dequeue {
		p.items = p.items[0 : len(p.items)-1]
	}
	return item
}

func (p *PrimitiveQueue) Last() PageTrack {
	return p.last(false)
}
