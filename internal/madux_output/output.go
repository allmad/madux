package madux_output

import "io"

type Moutput struct {
	out io.Writer
}

func New(out io.Writer) *Moutput {
	return &Moutput{
		out: out,
	}
}

func (m *Moutput) ClearScreen() {
	m.out.Write([]byte("\033[2J"))
	m.out.Write([]byte("\033[H"))
}

func (m *Moutput) Restore() {
	m.out.Write([]byte("\033[J"))
}
