package madux_conn

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

type Buffer struct {
	buf    []byte
	rd     *bufio.Reader
	w      *bufio.Writer
	source io.ReadWriteCloser
	bo     binary.ByteOrder
	err    error
}

func NewBuffer(source io.ReadWriteCloser) *Buffer {
	return &Buffer{
		source: source,
		rd:     bufio.NewReader(source),
		w:      bufio.NewWriter(source),
		bo:     binary.BigEndian,
	}
}

func (r *Buffer) Buffered() int {
	return r.w.Buffered()
}

func (r *Buffer) getbuf(n int) []byte {
	if len(r.buf) < n {
		r.buf = make([]byte, n)
	}
	return r.buf[:n]
}

func (r *Buffer) Int16() int16 {
	return int16(r.Uint16())
}

func (r *Buffer) seterr(err error) {
	if r.err == nil {
		r.err = err
	}
}

func (r *Buffer) Message(msg Message) {
	r.seterr(msg.Unmarshal(r))
}

func (r *Buffer) Uint16() uint16 {
	buf := r.getbuf(2)
	_, err := r.rd.Read(buf)
	r.seterr(err)
	return r.bo.Uint16(buf)
}

func (r *Buffer) Flush() error {
	return r.w.Flush()
}

func (r *Buffer) PutMessage(msg Message) {
	r.seterr(msg.Marshal(r))
}

func (r *Buffer) String() string {
	size := int(r.Int32())
	if size > 1<<20 {
		panic(fmt.Sprintf("String: too much size: %v", size))
	}
	if size == 0 {
		return ""
	}
	buf := r.getbuf(size)
	_, err := r.rd.Read(buf)
	r.seterr(err)
	return string(buf)
}

func (r *Buffer) PutBytes(b []byte) {
	r.PutInt32(int32(len(b)))
	if len(b) > 0 {
		_, err := r.w.Write(b)
		r.seterr(err)
	}
}

func (r *Buffer) PutBool(b bool) {
	buf := r.getbuf(1)
	if b {
		buf[0] = 1
	}
	_, err := r.w.Write(buf)
	r.seterr(err)
}

func (r *Buffer) Bool() bool {
	buf := r.getbuf(1)
	_, err := r.rd.Read(buf)
	r.seterr(err)
	return buf[0] == 1
}

func (r *Buffer) PutString(s string) {
	r.PutInt32(int32(len(s)))
	if len(s) > 0 {
		_, err := r.w.WriteString(s)
		r.seterr(err)
	}
}

func (r *Buffer) Int32() int32 {
	buf := r.getbuf(4)
	_, err := r.rd.Read(buf)
	r.seterr(err)
	return int32(r.bo.Uint32(buf))
}

func (r *Buffer) PutInt32(v int32) {
	buf := r.getbuf(4)
	r.bo.PutUint32(buf, uint32(v))
	_, err := r.w.Write(buf)
	r.seterr(err)
}

func (r *Buffer) PutInt16(v int16) {
	buf := r.getbuf(2)
	r.bo.PutUint16(buf, uint16(v))
	_, err := r.w.Write(buf)
	r.seterr(err)
	return
}

func (r *Buffer) Err() error {
	return r.err
}

func (r *Buffer) Bytes() []byte {
	length := r.Int32()
	if length == 0 {
		return nil
	}
	ret := make([]byte, length)
	_, err := r.rd.Read(ret)
	r.seterr(err)
	return ret
}

func (r *Buffer) Close() error {
	return r.source.Close()
}
