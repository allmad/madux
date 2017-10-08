package madux_server

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/allmad/madux/madux_conn"
	"github.com/allmad/madux/madux_context"
	"github.com/allmad/madux/madux_module"
	"github.com/allmad/madux/madux_plugin"
	"github.com/allmad/madux/madux_session"
	"github.com/chzyer/flow"
	"github.com/chzyer/logex"
	"github.com/chzyer/readline"

	"net/http"
	_ "net/http/pprof"
)

type Config struct {
	Network string `name:"net" default:"unix"`
	Host    string `default:"/tmp/madux.sock"`
	Debug   string `default:"/tmp/madux.debug"`
}

func (c *Config) FlaglyDesc() string {
	return "start a daemon process"
}

func (c *Config) FlaglyHandle(f *flow.Flow) error {
	defer f.Close()
	svr := New(f, c)
	return svr.Run()
}

type Server struct {
	cfg      *Config
	flow     *flow.Flow
	plugin   *madux_plugin.Plugin
	sessions madux_session.Sessions
}

func New(f *flow.Flow, cfg *Config) *Server {
	svr := &Server{
		cfg: cfg,
	}
	svr.plugin = madux_plugin.NewPlugin(svr)
	f.ForkTo(&svr.flow, svr.Close)
	return svr
}

func (s *Server) clearSock() {
	if s.cfg.Network != "unix" {
		return
	}
	os.Remove(s.cfg.Host)
}

// ensure that we only start one instance
func (s *Server) check() {
	if s.cfg.Network != "unix" {
		return
	}
	cli, err := madux_conn.Dial(s.cfg.Network, s.cfg.Host)
	if err != nil {
		s.clearSock()
		return
	}

	err = cli.SendMsg(&madux_module.Ping{})
	if err != nil {
		return
	}
}

func (s *Server) handleConn(netConn net.Conn) {
	defer println("exit")
	conn := madux_conn.NewConn(netConn)
	defer conn.Close()
	ctx := &madux_context.T{
		Conn:     conn,
		Sessions: &s.sessions,
	}
	for {
		m, err := conn.ReadMsg()
		if err != nil {
			if !logex.Equal(err, io.EOF) {
				logex.Error(err)
			}
			break
		}

		if err := madux_module.Handle(ctx, m); err != nil {
			logex.Error(err)
			return
		}
	}
}

func (s *Server) Run() error {
	s.check()
	go s.startDebug()
	go http.ListenAndServe("localhost:6060", nil)

	ln, err := net.Listen(s.cfg.Network, s.cfg.Host)
	if err != nil {
		return logex.Trace(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			break
		}
		go s.handleConn(conn)
	}
	return nil
}

func (s *Server) Close() {
	s.flow.Close()
}

func (s *Server) GetSessions() *madux_session.Sessions {
	return &s.sessions
}

func (s *Server) startDebug() {
	cfg := &readline.Config{
		Prompt:      "madux> ",
		HistoryFile: "/tmp/madux.debug.history",
	}
	os.Remove(s.cfg.Debug)
	err := readline.ListenRemote("unix", s.cfg.Debug, cfg, func(r *readline.Instance) {
		for {
			ret := r.Line()
			if ret.CanBreak() {
				break
			} else if ret.CanContinue() {
				continue
			}
			val, err := s.plugin.Eval(ret.Line)
			if err != nil {
				fmt.Fprintln(r, err)
				continue
			}
			fmt.Fprintln(r, val)
		}
	})
	if err != nil {
		fmt.Println(err)
	}
}
