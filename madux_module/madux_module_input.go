package madux_module

import (
	"github.com/allmad/madux/madux_conn"
	"github.com/allmad/madux/madux_context"
)

var _ Handler = &Input{}

type Input struct {
	SessionId int32
	Data      []byte
}

func (p *Input) Marshal(r *madux_conn.Buffer) error {
	r.PutInt32(p.SessionId)
	r.PutBytes(p.Data)
	return nil
}

func (p *Input) Unmarshal(r *madux_conn.Buffer) error {
	p.SessionId = r.Int32()
	p.Data = r.Bytes()
	return nil
}

func (i *Input) HandleClient(ctx *madux_context.C) error {
	return nil
}

func (i *Input) Handle(ctx *madux_context.T) error {
	sess := ctx.Sessions.GetById(i.SessionId)
	if sess != nil {
		sess.Pty.Write(i.Data)
	}
	return nil
}
