package parse

type Token struct {
	lexch chan item
	out   chan *item
}

func NewToken(lex *lexer) *Token {
	tkn := &Token{
		lexch: lex.items,
		out:   make(chan *item),
	}
	return tkn
}

func (t *Token) Run() {
	var last *item
	for item := range t.lexch {
		if item.typ == itemEOF {
			break
		}
		if item.typ == itemClear {
			continue
		}
		if last == nil {
			newLast := item
			last = &newLast
			continue
		}
		if item.typ != last.typ {
			t.out <- last
			newLast := item
			last = &newLast
			continue
		}
		last.val = append(last.val, item.val...)
	}

	t.out <- last
	close(t.out)
}
