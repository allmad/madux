package madux_plugin

import (
	"testing"

	"github.com/chzyer/test"
)

func TestPlugin(t *testing.T) {
	defer test.New(t)

	p := NewPlugin(nil)
	p.Run()
}
