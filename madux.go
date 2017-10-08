package main

import (
	"os"

	"github.com/allmad/madux/madux_client"
	"github.com/allmad/madux/madux_debug"
	"github.com/allmad/madux/madux_server"
	"github.com/chzyer/flagly"
	"github.com/chzyer/flow"
	"github.com/chzyer/logex"
)

type Madux struct {
	Server *madux_server.Config `flagly:"handler"`
	Client *madux_client.Config `flagly:"handler"`
	Debug  *madux_debug.Debug   `flagly:"handler"`
}

func main() {
	fset := flagly.New(os.Args[0])
	f := flow.New()
	fset.Context(f)
	if err := fset.Compile(&Madux{}); err != nil {
		panic(err)
	}

	if err := fset.Run(os.Args[1:]); err != nil {
		println(err.Error())
		os.Exit(1)
	}

	if err := f.Wait(); err != nil {
		logex.Error(err)
		os.Exit(1)
	}
}
