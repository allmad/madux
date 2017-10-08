package madux_conn

type Bytes []byte

func NewBytes(b []byte) *Bytes {
	return (*Bytes)(&b)
}

func (b *Bytes) Marshal(r *Buffer) error {
	r.PutBytes(*b)
	return nil
}

func (b *Bytes) Unmarshal(r *Buffer) error {
	*b = Bytes(r.Bytes())
	return nil
}

type String string

func NewString(str string) *String {
	return (*String)(&str)
}

func (s *String) Marshal(r *Buffer) error {
	r.PutString(string(*s))
	return nil
}

func (s *String) Unmarshal(r *Buffer) error {
	n := r.String()
	*s = String(n)
	return nil
}
