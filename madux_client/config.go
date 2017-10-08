package madux_client

import "github.com/chzyer/flow"

type Config struct {
	Net  string `default:"unix"`
	Host string `default:"/tmp/madux.sock"`
}

func (c *Config) FlaglyDesc() string {
	return "start a client"
}

func (c *Config) FlaglyHandle(f *flow.Flow) error {
	defer f.Close()

	cli := New(c, f)
	return cli.Run()
}
