package parser

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

// string terminator
// ^G, ST
func (r Rune) isST() bool {
	return r.in(0x9c, 0x18, 0x1A, 0x1b, 0x07)
}
