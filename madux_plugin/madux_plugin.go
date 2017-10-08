package madux_plugin

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/allmad/madux/madux_session"

	"qlang.io/cl/qlang"
	_ "qlang.io/lib/builtin"
)

type ServerDelegater interface {
	GetSessions() *madux_session.Sessions
}

type Plugin struct {
	vm        *qlang.Qlang
	delegater ServerDelegater
}

func NewPlugin(delegater ServerDelegater) *Plugin {
	p := &Plugin{qlang.New(), delegater}
	p.init()
	return p
}

func (p *Plugin) init() {
	base := &BasePlugin{p, p.delegater}
	p.Register(&PluginBuiltin{base: base})
	p.Register(&PluginSession{base: base})
}

func (p *Plugin) Register(obj interface{}) error {
	val := reflect.ValueOf(obj)
	typ := val.Type()
	typElem := typ
	if typ.Kind() == reflect.Ptr {
		typElem = typ.Elem()
	}
	name := typElem.Name()
	if !strings.HasPrefix(name, "Plugin") {
		return fmt.Errorf(`unknown plugin, not startwith "Plugin"`)
	}
	name = strings.TrimPrefix(name, "Plugin")
	if name == "Builtin" {
		name = ""
	}
	methods := make(map[string]interface{})
	for i := 0; i < val.NumMethod(); i++ {
		method := val.Method(i)
		methodName := typ.Method(i).Name
		methods[methodName] = method.Interface()
	}
	qlang.Import(name, methods)
	return nil
}

func (p *Plugin) Eval(s string) (interface{}, error) {
	return p.vm.Call([]byte(s), "")
}

func (p *Plugin) Run() error {
	ret, err := p.vm.Call([]byte(`
Session.Create("12312")
a = Session.List()
a[0]
`), "")
	fmt.Println(ret)
	return err
}
