package madux_module

import (
	"github.com/allmad/madux/madux_conn"
	"github.com/allmad/madux/madux_context"
)

var _ Handler = &Ping{}

type Ping struct {
}

func (s *Ping) Marshal(b *madux_conn.Buffer) error {
	return nil
}

func (s *Ping) Unmarshal(b *madux_conn.Buffer) error {
	return nil
}

func (s *Ping) Handle(ctx *madux_context.T) error {
	return nil
}

func (s *Ping) HandleClient(ctx *madux_context.C) error {
	return nil
}
