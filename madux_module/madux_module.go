package madux_module

import (
	"os"
	"reflect"

	"github.com/allmad/madux/madux_conn"
	"github.com/allmad/madux/madux_context"
)

type Module struct {
	Session
	Input
	Ping
	Config
}

var mod Module

type Handler interface {
	madux_conn.Message
	Handle(*madux_context.T) error
	HandleClient(*madux_context.C) error
}

type modInfo struct {
	Name    string
	Handler Handler
}

func Handle(ctx *madux_context.T, msg madux_conn.Message) error {
	return msg.(Handler).Handle(ctx)
}

// ---------------------------------------------------------------------------

func init() {
	val := reflect.ValueOf(&mod).Elem()
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i).Addr()
		fieldTyp := val.Field(i).Type()
		field, ok := fieldVal.Interface().(Handler)
		if !ok {
			println("not implement interface Handler:", typ.Field(i).Name)
			os.Exit(2)
		}
		typInt := int16(i)
		madux_conn.SetType(field, typInt)
		madux_conn.RegisterNewFn(typInt, func() madux_conn.Message {
			val := reflect.New(fieldTyp)
			handler := val.Interface().(madux_conn.Message)
			return handler
		})
	}
}
