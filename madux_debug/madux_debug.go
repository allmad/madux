package madux_debug

import (
	"github.com/chzyer/flow"
	"github.com/chzyer/readline"
)

type Debug struct {
	Addr string `default:"/tmp/madux.debug"`
}

func (d *Debug) FlaglyHandle(f *flow.Flow) error {
	defer f.Close()
	err := readline.DialRemote("unix", d.Addr)
	if err != nil {
		return err
	}
	return nil
}
