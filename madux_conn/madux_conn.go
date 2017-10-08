package madux_conn

import (
	"fmt"
	"net"
	"sync"

	"github.com/chzyer/logex"
)

type Conn struct {
	Name   string
	conn   net.Conn
	buffer *Buffer
	mr     sync.Mutex
	mw     sync.Mutex

	Width  int
	Height int
}

func NewConn(conn net.Conn) *Conn {
	buf := NewBuffer(conn)
	return &Conn{
		conn:   conn,
		buffer: buf,
	}
}

type Message interface {
	Marshal(*Buffer) error
	Unmarshal(*Buffer) error
}

type Item struct {
	Type    int16
	Payload Message
}

func (c *Conn) SetName(name string) {
	c.Name = name
}

func (c *Conn) ReadMsg() (Message, error) {
	var m Item
	c.mr.Lock()
	defer c.mr.Unlock()

	m.Type = c.buffer.Int16()
	if err := c.buffer.Err(); err != nil {
		return nil, err
	}
	fn := mods[m.Type]
	if fn == nil {
		println(123123, m.Type)
		return nil, fmt.Errorf("invalid type: %v", m.Type)
	}
	m.Payload = fn()
	c.buffer.Message(m.Payload)
	if err := c.buffer.Err(); err != nil {
		return nil, err
	}
	return m.Payload, nil
}

func (c *Conn) SendMsg(payload Message) error {
	c.mw.Lock()
	defer c.mw.Unlock()

	c.buffer.PutInt16(handlerMap[getTypeInfo(payload)])
	c.buffer.PutMessage(payload)
	if err := c.buffer.Flush(); err != nil {
		return logex.Trace(err)
	}
	return nil
}

func (c *Conn) Close() error {
	return c.buffer.Close()
}

func Dial(network, addr string) (*Conn, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, logex.Trace(err)
	}
	return NewConn(conn), nil
}
