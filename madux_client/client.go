package madux_client

import (
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/allmad/madux/madux_conn"
	"github.com/allmad/madux/madux_context"
	"github.com/allmad/madux/madux_module"
	"github.com/allmad/madux/madux_term"
	"github.com/chzyer/flow"
	"github.com/chzyer/logex"
)

type Client struct {
	cfg  *Config
	flow *flow.Flow
	conn *madux_conn.Conn
}

func New(cfg *Config, f *flow.Flow) *Client {
	return &Client{
		flow: f,
		cfg:  cfg,
	}
}

func (c *Client) Run() error {
	go http.ListenAndServe("localhost:6061", nil)

	state, err := madux_term.MakeRaw(0)
	if err != nil {
		return err
	}
	w, h, err := madux_term.GetSize(0)
	if err != nil {
		return err
	}
	defer func() {
		c.flow.Wait()
		madux_term.Restore(0, state)
	}()
	_, _ = w, h

	conn, err := madux_conn.Dial(c.cfg.Net, c.cfg.Host)
	if err != nil {
		return logex.Trace(err)
	}
	c.conn = conn
	conn.SendMsg(&madux_module.Config{
		Height: int32(h),
		Width:  int32(w),
	})

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			err2 := conn.SendMsg(&madux_module.Input{
				SessionId: 1,
				Data:      buf[:n],
			})

			if err != nil || err2 != nil {
				logex.Error(err, err2)
				break
			}
		}
	}()
	ctx := &madux_context.C{}
	go c.loop(ctx)

	go func() {
		for range time.Tick(100 * time.Millisecond) {
			err := conn.SendMsg(&madux_module.Session{
				Id:    1,
				Start: ctx.Offset,
				Type:  madux_module.SessionTypePull,
			})
			if err != nil {
				logex.Error(err)
			}
		}
	}()
	// os.Stdout.Write(payload.Data)
	return nil
}

func (c *Client) loop(ctx *madux_context.C) {
	for {
		msg, err := c.conn.ReadMsg()
		if err != nil {
			break
		}
		err = msg.(madux_module.Handler).HandleClient(ctx)
		if err != nil {
			logex.Error(err)
			break
		}
	}
}

func (c *Client) Fork() error {
	cmd := exec.Command(os.Args[0], "server")
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setsid = true
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}
