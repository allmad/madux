// reference http://rtfm.etla.org/xterm/ctlseq.html
package parser

import (
	"bufio"
	"fmt"
	"io"
)

type item struct {
	typ itemType
	val []byte
}

func (i item) String() string {
	return fmt.Sprintf("{typ: %v, val: %q}",
		i.typ, i.val)
}

const (
	eof Rune = -1
)

type itemType int

const (
	itemNil itemType = iota
	itemError
	itemEOF
	itemPrint
	itemClear
	itemExecute
	itemIgnore
	itemNoAction
	itemCsiDispatch // Control Sequence Introducer
	itemEscDispatch
	itemParam
	itemCollect
	itemEscape
	itemDcsDispatch
	itemOscString // Operating System Command

	// A single (usually optional) numeric parameter, composed of one of more digits.
	itemPs

	// A multiple numeric parameter composed of any number of single numeric parameters, separated by ; character(s). Individual values for the parameters are listed with Ps .
	itemPm

	// A text parameter composed of printable characters.
	itemPt
)

func (i itemType) IsDispatch() bool {
	switch i {
	case itemCsiDispatch, itemEscDispatch, itemDcsDispatch:
		fallthrough
	case itemOscString, itemPrint, itemExecute, itemClear:
		return true
	default:
		return false
	}
}

func (i itemType) String() string {
	switch i {
	case itemNil:
		return "nil"
	case itemError:
		return "error"
	case itemEOF:
		return "eof"
	case itemPrint:
		return "print"
	case itemClear:
		return "clear"
	case itemExecute:
		return "execute"
	case itemIgnore:
		return "ignore"
	case itemNoAction:
		return "noaction"
	case itemCsiDispatch:
		return "csi_dispatch"
	case itemEscDispatch:
		return "esc_dispatch"
	case itemParam:
		return "param"
	case itemCollect:
		return "collect"
	case itemEscape:
		return "escape"
	case itemOscString:
		return "osc_string"
	default:
		return fmt.Sprintf("#%v", int(i))
	}
}

type stateFn func(l *lexer) stateFn

type lexer struct {
	reader *bufio.Reader
	buffer []byte
	state  stateFn
	items  chan item

	lastItemType itemType
	lastSize     int

	syncPos int
}

func Lex(r io.Reader) *lexer {
	l := &lexer{
		reader: bufio.NewReader(r),
		items:  make(chan item),
	}
	go l.run()
	return l
}

func (l *lexer) run() {
	for l.state = stateAnywhere; l.state != nil; {
		l.state = l.state(l)
	}

	if len(l.buffer) > 0 {
		l.items <- item{
			typ: l.lastItemType,
			val: l.buffer,
		}
	}
	close(l.items)
}

func (l *lexer) defval(r Rune) stateFn {
	if r == eof {
		return nil
	}
	panic(fmt.Sprintf("unreachable : %v", r))
}

func (l *lexer) fire(i itemType) {
	l.items <- item{typ: i}
}

func (l *lexer) action(typ itemType, fn stateFn) stateFn {
	l.emit(typ)
	return fn
}

func (l *lexer) skip() {
	l.emit(itemNil)
}

func (l *lexer) emit(i itemType) {
	if i == itemNil {
		l.buffer = l.buffer[:0]
		return
	}

	buf := make([]byte, len(l.buffer))
	copy(buf, l.buffer)
	l.buffer = l.buffer[:0]
	l.lastItemType = i

	it := item{
		typ: i,
		val: buf,
	}
	l.items <- it
}

func (l *lexer) discard() {
	l.buffer = l.buffer[:len(l.buffer)-l.lastSize]
	l.lastSize = 0
}

func (l *lexer) peek() Rune {
	r, _, err := l.reader.ReadRune()
	if err == io.EOF {
		return eof
	}
	l.reader.UnreadRune()
	return Rune(r)
}

func (l *lexer) next() Rune {
	r, size, err := l.reader.ReadRune()
	if err == io.EOF {
		return eof
	}
	l.lastSize = size

	l.buffer = append(l.buffer, []byte(string(r))...)
	return Rune(r)
}

func (l *lexer) doClear() {
	l.emit(itemClear)
}

// -----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
