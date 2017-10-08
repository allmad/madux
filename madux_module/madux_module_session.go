package madux_module

import (
	"os"

	"github.com/allmad/madux/madux_conn"
	"github.com/allmad/madux/madux_context"
)

var _ Handler = &Session{}

type SessionType int32

const (
	SessionTypePull SessionType = iota
	SessionTypePullResp
	SessionTypeList
	SessionTypeListResp
)

type SessionItem struct {
	Id       int32
	Name     string
	Attached bool
}

func (s *SessionItem) Marshal(b *madux_conn.Buffer) error {
	b.PutInt32(s.Id)
	b.PutString(s.Name)
	b.PutBool(s.Attached)
	return nil
}

func (s *SessionItem) Unmarshal(b *madux_conn.Buffer) error {
	s.Id = b.Int32()
	s.Name = b.String()
	s.Attached = b.Bool()
	return nil
}

type SessionList []*SessionItem

type Session struct {
	Id   int32
	Type SessionType

	Start  int32
	Output []byte
	List   []*SessionItem
}

func (s *Session) Marshal(b *madux_conn.Buffer) error {
	b.PutInt32(s.Id)
	b.PutInt32(int32(s.Type))
	b.PutInt32(s.Start)
	b.PutBytes(s.Output)
	b.PutInt32(int32(len(s.List)))
	for _, item := range s.List {
		b.PutMessage(item)
	}
	return nil
}

func (s *Session) Unmarshal(b *madux_conn.Buffer) error {
	s.Id = b.Int32()
	s.Type = SessionType(b.Int32())
	s.Start = b.Int32()
	s.Output = b.Bytes()
	listCnt := b.Int32()
	if listCnt > 0 {
		s.List = make([]*SessionItem, listCnt)
		for idx := range s.List {
			s.List[idx] = &SessionItem{}
			b.Message(s.List[idx])
		}
	}
	return nil
}

func (s *Session) HandleClient(ctx *madux_context.C) error {
	switch s.Type {
	case SessionTypeListResp:
	case SessionTypePullResp:
		ctx.Offset = s.Start
		os.Stdout.Write(s.Output)
	}
	return nil
}

func (s *Session) Handle(ctx *madux_context.T) error {
	switch s.Type {
	case SessionTypePull:
		sess := ctx.Sessions.GetById(s.Id)
		if sess == nil {
			return nil
		}
		if ctx.Conn.Width > 0 {
			sess.Pty.SetSize(ctx.Conn.Width, ctx.Conn.Height)
		}
		output, start := sess.GetOutput(int(s.Start))

		return ctx.Conn.SendMsg(&Session{
			Id:     s.Id,
			Type:   SessionTypePullResp,
			Start:  int32(start),
			Output: output,
		})
	case SessionTypeList:
		ret := make([]*SessionItem, len(ctx.Sessions.List()))
		for idx, sess := range ctx.Sessions.List() {
			ret[idx] = &SessionItem{
				Id:   sess.Id,
				Name: sess.Name,
			}
		}
		return ctx.Conn.SendMsg(&Session{
			Type: SessionTypeListResp,
			List: ret,
		})
	}
	return nil
}
