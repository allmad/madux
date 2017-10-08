package madux_window

import "github.com/allmad/madux/go/madux_output"

type Window struct {
	mop *madux_output.Moutput
}

func New(mop *madux_output.Moutput) *Window {
	return &Window{
		mop: mop,
	}
}

func (w *Window) Init() {
	w.mop.ClearScreen()
}
