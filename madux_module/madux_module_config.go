package madux_module

import (
	"github.com/allmad/madux/madux_conn"
	"github.com/allmad/madux/madux_context"
)

type Config struct {
	Name   string
	Height int32
	Width  int32
}

func (p *Config) Marshal(r *madux_conn.Buffer) error {
	r.PutString(p.Name)
	r.PutInt32(p.Height)
	r.PutInt32(p.Width)
	return nil
}

func (p *Config) Unmarshal(r *madux_conn.Buffer) error {
	p.Name = r.String()
	p.Height = r.Int32()
	p.Width = r.Int32()
	return nil
}

func (i *Config) HandleClient(ctx *madux_context.C) error {
	return nil
}

func (c *Config) Handle(ctx *madux_context.T) error {
	if c.Name != "" {
		ctx.Conn.SetName(c.Name)
	}
	if c.Height > 0 {
		ctx.Conn.Height = int(c.Height)
	}
	if c.Width > 0 {
		ctx.Conn.Width = int(c.Width)
	}
	return nil
}
