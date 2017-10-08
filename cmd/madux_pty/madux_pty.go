package main

import (
	"io"
	"os"
	"os/exec"

	"github.com/allmad/madux/go/debug"
	"github.com/allmad/madux/go/term"
	"github.com/chzyer/logex"
)

func main() {
	state, err := term.MakeRaw(0)
	if err != nil {
		logex.Fatal(err)
	}
	defer term.Restore(0, state)

	pty, err := term.NewPty()
	if err != nil {
		logex.Error(err)
		return
	}
	// pty.CopyAttr(os.Stdin.Fd())
	w, h, _ := term.GetSize(0)
	// pty.SetSize(49, 10)
	pty.SetSize(w, h)

	col, err := debug.NewConsole()
	if err != nil {
		logex.Error(err)
		return
	}
	defer col.Close()
	go col.Run()

	// cmd := exec.Command("vim", "/tmp/madux_ex.go")
	cmd := exec.Command("zsh")
	pty.SetCmd(cmd)
	go io.Copy(pty, os.Stdin)
	go col.CopyTo(os.Stdout)
	go func() {
		fd, _ := os.OpenFile("/tmp/madux", os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0666)
		buf := make([]byte, 10240)
		for {
			n, err := pty.Read(buf)
			if err != nil {
				break
			}
			fd.Write(buf[:n])
			os.Stdout.Write(buf[:n])
		}
	}()

	if err := cmd.Run(); err != nil {
		logex.Error(err)
		return
	}
}
