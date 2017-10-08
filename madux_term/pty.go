package madux_term

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/pkg/term/termios"
)

type Pty struct {
	cmd  *exec.Cmd
	w, h int

	pty *os.File
	tty *os.File
}

func CopyTerm(cmd *exec.Cmd) (*Pty, error) {
	pty, err := NewPty()
	if err != nil {
		return nil, err
	}
	pty.SetCmd(cmd)
	return pty, nil
}

func NewPty() (*Pty, error) {
	pty, tty, err := termios.Pty()
	if err != nil {
		return nil, err
	}
	return &Pty{
		pty: pty,
		tty: tty,
	}, nil
}

func (p *Pty) setAttr(attr *syscall.Termios) error {
	return termios.Tcsetattr(p.tty.Fd(), termios.TCSADRAIN, attr)
}

func (p *Pty) CopyAttr(fd uintptr) error {
	var attr syscall.Termios
	if err := termios.Tcgetattr(fd, &attr); err != nil {
		return err
	}
	return p.setAttr(&attr)
}

func (p *Pty) copySize(fd int) error {
	w, h, err := GetSize(fd)
	if err != nil {
		return err
	}
	return p.SetSize(w, h-1)
}

func (p *Pty) SetSize(w, h int) error {
	if p.w == w && p.h == h {
		return nil
	}
	println(1123)
	if err := SetSize(p.tty.Fd(), w, h); err != nil {
		return err
	}
	p.w = w
	p.h = h
	p.cmd.Process.Signal(syscall.SIGWINCH)
	return nil
}

func (p *Pty) SetCmd(c *exec.Cmd) {
	p.cmd = c
	c.Stdin = p.tty
	c.Stdout = p.tty
	c.Stderr = p.tty
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Setctty = true
	c.SysProcAttr.Setsid = true
}

func (p *Pty) Read(b []byte) (int, error) {
	return p.pty.Read(b)
}

func (p *Pty) Write(b []byte) (int, error) {
	return p.pty.Write(b)
}

func (p *Pty) Close() {
	p.tty.Close()
	p.pty.Close()
}

func (p *Pty) WindowChange(w, h int) error {
	return SetSize(p.pty.Fd(), w, h)
}
