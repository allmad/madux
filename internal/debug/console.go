package debug

import (
	"fmt"
	"io"
	"net"
	"os"
)

type Console struct {
	ln net.Listener
	pr io.ReadCloser
	pw io.WriteCloser
}

func NewConsole() (*Console, error) {
	fname := fmt.Sprintf("/tmp/madux.%v", 0)
	os.Remove(fname)

	ln, err := net.Listen("unix", fname)
	if err != nil {
		return nil, err
	}
	pr, pw := io.Pipe()
	return &Console{
		ln: ln,
		pr: pr,
		pw: pw,
	}, nil
}

func (c *Console) Close() {
	os.Remove("/tmp/maudx.0")
}

func (c *Console) Run() {
	for {
		conn, err := c.ln.Accept()
		if err != nil {
			break
		}
		go c.handleConn(conn)
	}
}

func (c *Console) CopyTo(w io.Writer) {
	io.Copy(w, c.pr)
}

func (c *Console) handleConn(conn net.Conn) {
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}
		c.pw.Write(buf[:n])
	}
}
