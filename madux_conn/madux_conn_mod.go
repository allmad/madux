package madux_conn

import "unsafe"

var mods = map[int16]func() Message{}

func RegisterNewFn(typ int16, fn func() Message) {
	mods[typ] = fn
}

var handlerMap = map[int64]int16{}

func getTypeInfo(h Message) int64 {
	typeInfo := (*struct{ Type uintptr })(unsafe.Pointer(&h)).Type
	return int64(typeInfo)
}

func SetType(h Message, val int16) {
	handlerMap[getTypeInfo(h)] = val
}

func GetType(h Message) int16 {
	return handlerMap[getTypeInfo(h)]
}
