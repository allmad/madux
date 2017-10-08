package madux_context

import (
	"github.com/allmad/madux/madux_conn"
	"github.com/allmad/madux/madux_session"
)

type T struct {
	Conn     *madux_conn.Conn
	Sessions *madux_session.Sessions
}

type C struct {
	Offset int32
}
