package madux_plugin

import "github.com/allmad/madux/madux_session"

type PluginSession struct {
	base *BasePlugin
}

func (p *PluginSession) Export() []interface{} {
	return []interface{}{}
}

func (p *PluginSession) List() []*madux_session.Session {
	return p.base.delegate.GetSessions().List()
}

func (p *PluginSession) Create(name, cmd string) error {
	session, err := madux_session.NewSession(cmd)
	if err != nil {
		return err
	}
	sessions := p.base.delegate.GetSessions()
	sessions.Add(session)
	return nil
}
