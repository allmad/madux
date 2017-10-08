package madux_session

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"

	"github.com/allmad/madux/madux_term"
	"github.com/chzyer/logex"
	"github.com/google/shlex"
)

type Session struct {
	Name string
	Id   int32

	cmd    *exec.Cmd
	Pty    *madux_term.Pty
	output *bytes.Buffer
	height int32
	width  int32
}

func NewSession(cmdStr string) (*Session, error) {
	cmdsp, err := shlex.Split(cmdStr)
	if err != nil {
		return nil, logex.Trace(err)
	}
	cmd := exec.Command(cmdsp[0], cmdsp[1:]...)
	pty, err := madux_term.CopyTerm(cmd)
	if err != nil {
		return nil, logex.Trace(err)
	}
	return &Session{
		Pty:    pty,
		cmd:    cmd,
		output: bytes.NewBuffer(nil),
	}, nil
}

func (s *Session) String() string {
	return fmt.Sprintf(`{"Id":%d, "name":%q}`,
		s.Id, s.Name)
}

func (s *Session) Size(w, h int32) {
	s.width = w
	s.height = h
	s.Pty.SetSize(int(w), int(h))
}

func (s *Session) Start() error {
	go io.Copy(s.output, s.Pty)
	if err := s.cmd.Start(); err != nil {
		return err
	}
	return nil
}

func (s *Session) Wait() error {
	if err := s.cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *Session) GetOutput(start int) ([]byte, int) {
	out := s.output.Bytes()
	b := out[start:]
	return b, start + len(b)
}

// ---------------------------------------------------------------------------

type Sessions struct {
	IdSeq int32
	list  []*Session
	m     sync.Mutex
}

func (s *Sessions) GetOutput(idx int) []byte {
	return s.list[idx].output.Bytes()
}

func (s *Sessions) GetById(id int32) *Session {
	for _, item := range s.list {
		if item.Id == id {
			return item
		}
	}
	return nil
}

func (s *Sessions) List() []*Session {
	return s.list
}

func (s *Sessions) Add(sess *Session) {
	sess.Id = atomic.AddInt32(&s.IdSeq, 1)
	err := sess.Start()
	if err != nil {
		logex.Error(err)
		return
	}
	go func() {
		sess.Wait()
		s.m.Lock()
		for idx := range s.list {
			if s.list[idx] == sess {
				s.list = append(s.list[:idx], s.list[idx+1:]...)
				break
			}
		}
		s.m.Unlock()
	}()

	s.m.Lock()
	s.list = append(s.list, sess)
	s.m.Unlock()
}
