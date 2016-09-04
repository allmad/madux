package parse

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Rune rune

func (r Rune) rn(start, end rune) bool {
	return rune(r) >= start && rune(r) <= end
}

func (r Rune) isExecute() bool {
	return r.rn(0x00, 0x17) || r.in(0x19) || r.rn(0x1C, 0x1F)
}

func (r Rune) in(n ...rune) bool {
	for idx := range n {
		if rune(r) == n[idx] {
			return true
		}
	}
	return false
}

func (r Rune) isST() bool {
	return r.in(0x9c, 0x18, 0x1A, 0x1b, 0x07)
}

type item struct {
	typ itemType
	val []byte
}

func (i item) String() string {
	return fmt.Sprintf("{typ: %v, val: %v}",
		i.typ, strconv.Quote(string(i.val)),
	)
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
	itemCsiDispatch
	itemEscDispatch
	itemParam
	itemCollect
	itemEscape
	itemCsiEntry
	itemDcsEntry
	itemDcsDispatch
	itemOscString
)

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
	case itemCsiEntry:
		return "csi_entry"
	case itemDcsEntry:
		return "dcs_entry"
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

	l.items <- item{
		typ: l.lastItemType,
		val: l.buffer,
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

func (l *lexer) next() Rune {
	r, size, err := l.reader.ReadRune()
	if err == io.EOF {
		return eof
	}
	l.lastSize = size

	l.buffer = append(l.buffer, []byte(string(r))...)
	return Rune(r)
}

// -----------------------------------------------------------------------------

func stateAnywhere(l *lexer) stateFn {
	switch r := l.next(); {
	case r.in(0x18, 0x1a):
		fallthrough
	case r.rn(0x80, 0x8F), r.rn(0x91, 0x97), r.in(0x99, 0x9A):
		return l.action(itemExecute, stateAnywhere)

	case r.in(0x9c):
		return l.action(itemNoAction, stateAnywhere)

	case r.in(0x1b):
		return l.action(itemNil, stateEscape)
	case r.in(0x90):
		return l.action(itemNil, stateDcsEntry)
	case r.in(0x9b):
		return l.action(itemNil, stateCsiEntry)
	case r.in(0x9d):
		return l.action(itemNil, stateOscString)
	case r.in(0x98, 0x9e, 0x9f):
		return l.action(itemNil, stateSOS_PM_APC)
	default:

		switch {
		case r.isExecute():
			return l.action(itemExecute, stateAnywhere)
		case r == eof:
			return nil
		default: // text
			return l.action(itemPrint, stateAnywhere)
		}
	}
}

func stateEscape(l *lexer) stateFn {
	l.fire(itemClear)

	for {
		switch r := l.next(); {
		case r.isExecute():
			l.emit(itemExecute)
			continue
		case r.in(0x7f):
			l.emit(itemIgnore)
			continue

		case r.rn(0x30, 0x4F), r.rn(0x51, 0x57), r.rn(0x60, 0x7E):
			fallthrough
		case r.in(0x59, 0x5A, 0x5C):
			return l.action(itemEscDispatch, stateAnywhere)

		case r.rn(0x20, 0x2F):
			return l.action(itemCollect, stateEscapeIntermedia)
		case r.in(0x5b):
			return l.action(itemNil, stateCsiEntry)
		case r.in(0x50):
			return l.action(itemNil, stateDcsEntry)
		case r.in(0x58, 0x5E, 0x5F):
			return l.action(itemNil, stateSOS_PM_APC)
		case r.in(0x5d):
			return l.action(itemNil, stateOscString)
		default:
			return l.defval(r)
		}
	}
	return nil
}

func stateEscapeIntermedia(l *lexer) stateFn {
	switch r := l.next(); {
	case r.isExecute():
		return l.action(itemExecute, stateEscapeIntermedia)
	case r.rn(0x20, 0x2f):
		return l.action(itemCollect, stateEscapeIntermedia)
	case r.in(0x7f):
		return l.action(itemIgnore, stateEscapeIntermedia)
	case r.rn(0x30, 0x7e):
		return l.action(itemEscDispatch, stateAnywhere)
	default:
		return l.defval(r)
	}
}

func stateCsiEntry(l *lexer) stateFn {
	l.fire(itemClear)

	for {
		switch r := l.next(); {
		case r.isExecute():
			l.emit(itemExecute)
			continue
		case r.in(0x7f):
			l.emit(itemIgnore)
			continue

		case r.rn(0x40, 0x7e):
			return l.action(itemCsiDispatch, stateAnywhere)

		// action:
		case r.rn(0x30, 0x39), r.in(0x3b):
			return l.action(itemParam, stateCsiParam)
		case r.rn(0x3c, 0x3f):
			return l.action(itemCollect, stateCsiParam)
		case r.rn(0x20, 0x2f):
			return l.action(itemCollect, stateCsiIntermediate)
		case r.in(0x3a):
			return l.action(itemNil, stateCsiIgnore)
		default:
			return l.defval(r)
		}
	}
}

func stateCsiParam(l *lexer) stateFn {
	switch r := l.next(); {
	case r.isExecute():
		return l.action(itemExecute, stateCsiParam)
	case r.rn(0x30, 0x39), r.in(0x3B):
		return l.action(itemParam, stateCsiParam)
	case r.in(0x7f):
		return l.action(itemIgnore, stateCsiParam)
	case r.rn(0x40, 0x7e):
		return l.action(itemCsiDispatch, stateAnywhere)
	case r.rn(0x20, 0x2f):
		return l.action(itemCollect, stateCsiIntermediate)
	case r.in(0x3A), r.rn(0x3C, 0x3F):
		return l.action(itemNil, stateCsiIgnore)
	default:
		return l.defval(r)
	}
}

func stateCsiIntermediate(l *lexer) stateFn {
	switch r := l.next(); {
	case r.isExecute():
		return l.action(itemExecute, stateCsiIntermediate)
	case r.rn(0x20, 0x2f):
		return l.action(itemCollect, stateCsiIntermediate)
	case r.rn(0x30, 0x3f):
		return l.action(itemNil, stateCsiIgnore)
	case r.in(0x7f):
		return l.action(itemIgnore, stateCsiIntermediate)
	case r.rn(0x40, 0x7e):
		return l.action(itemCsiDispatch, stateAnywhere)
	default:
		return l.defval(r)
	}
}

func stateCsiIgnore(l *lexer) stateFn {
	switch r := l.next(); {
	case r.isExecute():
		return l.action(itemExecute, stateCsiIgnore)
	case r.rn(0x20, 0x3f), r.in(0x7f):
		return l.action(itemIgnore, stateCsiIgnore)
	case r.rn(0x40, 0x7e):
		return l.action(itemNil, stateAnywhere)
	default:
		return l.defval(r)
	}
}

func stateDcsEntry(l *lexer) stateFn {
	l.fire(itemClear)

	for {
		switch r := l.next(); {
		case r.isExecute():
			l.emit(itemExecute)
			continue
		case r.in(0x7f):
			l.emit(itemIgnore)
			continue
		// transmit
		case r.rn(0x40, 0x7E):
			return l.action(itemNil, stateDcsPassthrough)
		case r.rn(0x20, 0x2F):
			return l.action(itemCollect, stateDcsIntermediate)
		case r.in(0x3a):
			return l.action(itemNil, stateDcsIgnore)
		case r.rn(0x30, 0x39), r.in(0x3B):
			return l.action(itemParam, stateDcsParam)
		case r.rn(0x3C, 0x3F):
			return l.action(itemCollect, stateDcsParam)
		default:
			return l.defval(r)
		}
	}
}

func stateDcsIntermediate(l *lexer) stateFn {
	switch r := l.next(); {
	case r.rn(0x00, 0x17), r.in(0x19), r.rn(0x1C, 0x1F):
		return l.action(itemExecute, stateDcsIntermediate)
	case r.rn(0x20, 0x2f):
		return l.action(itemCollect, stateDcsIntermediate)
	case r.in(0x7f):
		return l.action(itemIgnore, stateDcsIntermediate)

	// action
	case r.rn(0x30, 0x3f):
		return l.action(itemNil, stateDcsIgnore)
	case r.rn(0x40, 0x7E):
		return l.action(itemNil, stateDcsPassthrough)
	default:
		return l.defval(r)
	}
}

func stateDcsIgnore(l *lexer) stateFn {
	switch r := l.next(); {
	case r.isExecute(), r.rn(0x20, 0x7F):
		return l.action(itemIgnore, stateDcsIgnore)

	// action
	case r.in(0x9c):
		return l.action(itemNil, stateAnywhere)

	default:
		return l.defval(r)
	}
}

func stateDcsParam(l *lexer) stateFn {
	switch r := l.next(); {
	case r.isExecute():
		return l.action(itemIgnore, stateDcsParam)
	case r.rn(0x30, 0x39), r.in(0x3B):
		return l.action(itemParam, stateDcsParam)
	case r.in(0x7c):
		return l.action(itemIgnore, stateDcsParam)

	// action
	case r.in(0x3A), r.rn(0x3C, 0x3F):
		return l.action(itemNil, stateDcsIgnore)
	case r.rn(0x20, 0x2F):
		return l.action(itemCollect, stateDcsIntermediate)
	case r.rn(0x40, 0x7E):
		return l.action(itemNil, stateDcsPassthrough)

	default:
		return l.defval(r)
	}
}

func stateDcsPassthrough(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r.isST():
			l.discard()
			return l.action(itemDcsDispatch, stateAnywhere)
		case r.isExecute(), r.rn(0x20, 0x7E):
			// put
			continue
		case r.in(0x7f):
			l.emit(itemIgnore)
			continue
		default:
			return l.defval(r)
		}
	}
}

func stateOscString(l *lexer) stateFn {
	for {
		r := l.next()
		switch {
		case r.isST():
			l.discard()
			return l.action(itemOscString, stateAnywhere)
		case r.isExecute():
			l.emit(itemNil)
			continue
		case r == eof:
			return nil
		}
	}
}

func stateSOS_PM_APC(l *lexer) stateFn {
	return nil
}

// -----------------------------------------------------------------------------
