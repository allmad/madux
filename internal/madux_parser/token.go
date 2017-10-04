package parser

import "fmt"

type Func struct {
	typ     itemType
	name    []byte
	collect []byte
	param   []byte
}

func (f Func) String() string {
	return fmt.Sprintf(
		"&Func{typ: %v, param: %q, coll: %q, name: %q}",
		f.typ, f.param, f.collect, f.name,
	)
}

type Token struct {
	lexch chan item
	out   chan Func
}

func NewToken(lex *lexer) *Token {
	tkn := &Token{
		lexch: lex.items,
		out:   make(chan Func),
	}
	return tkn
}

func (t *Token) Run() {
	ch := make(chan item)
	go t.runMerge(ch)
	go func() {
		var f Func
		defer func() {
			close(t.out)
		}()
		for item := range ch {
			switch item.typ {
			case itemParam:
				f.param = append(f.param, item.val...)
			case itemCollect:
				f.collect = append(f.collect, item.val...)
			case itemClear:
				f.param = nil
				f.collect = nil
			default:
				if item.typ.IsDispatch() {
					f.typ = item.typ
					f.name = item.val
					t.out <- f
					f = Func{}
				}
			}
		}
	}()
}

func (t *Token) runMerge(out chan item) {
	var init bool
	var last item
	for item := range t.lexch {
		if item.typ == itemEOF {
			break
		}
		switch item.typ {
		case itemParam, itemPrint, itemCollect, itemExecute:
		default:
			if init {
				init = false
				out <- last
			}

			out <- item
			continue
		}
		if !init {
			last = item
			init = true
			continue
		}

		if item.typ != last.typ {
			out <- last
			last = item
			continue
		}
		last.val = append(last.val, item.val...)
	}

	out <- last
	close(out)
}
