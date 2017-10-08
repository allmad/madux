package main

import (
	"bytes"
	"net"
	"os"
	"strconv"

	"github.com/allmad/madux/go/term"
	"github.com/chzyer/logex"
	"github.com/chzyer/readline"
)

func main() {
	if err := process(); err != nil {
		logex.Error(err)
	}
}

func process() error {
	state, err := term.MakeRaw(0)
	if err != nil {
		return err
	}
	defer term.Restore(0, state)

	conn, err := net.Dial("unix", "/tmp/madux.0")
	if err != nil {
		return err
	}
	defer conn.Close()

	rl, _ := readline.New("> ")

	buf := make([]byte, 1024)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}
		if bytes.Contains(buf[:n], []byte{0x3}) {
			break
		}
		if bytes.Contains(buf[:n], []byte{0x1}) {
		readline:
			line, err := rl.Readline()
			if err != nil {
				print("(raw mode)")
				continue
			}
			unq, err := strconv.Unquote(`"` + line + `"`)
			if err != nil {
				logex.Error(err)
				print("(raw mode)")
				continue
			}
			conn.Write([]byte(unq))
			goto readline
		}
		conn.Write(buf[:n])
	}
	return nil
}
