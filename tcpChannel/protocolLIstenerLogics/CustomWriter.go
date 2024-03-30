package protocolLIstenerLogics

import (
	"encoding/binary"
)

type NetworkMessage interface {
	SerializeData(writer *CustomWriter)
}

// type MESSAGE_HANDLER_TYPE func(from string, protoSource APP_PROTO_ID, data []byte)
type MESSAGE_HANDLER_TYPE func(customConn *CustomConnection, protoSource APP_PROTO_ID, data *CustomReader)

type CustomWriter struct {
	data   []byte
	offset int
	order  binary.ByteOrder
}

func NewCustomWriter(capacity int, order binary.ByteOrder) *CustomWriter {
	return &CustomWriter{
		data:   make([]byte, capacity),
		offset: 0,
		order:  order,
	}
}
func NewCustomWriter2(data []byte, offset int, order binary.ByteOrder) *CustomWriter {
	return &CustomWriter{
		data:   data,
		offset: offset,
		order:  order,
	}
}
func NewCustomWriter3(order binary.ByteOrder) *CustomWriter {
	return NewCustomWriter(64, order)
}
func (c *CustomWriter) Write(p []byte) (n int, err error) {
	minCap := c.offset + len(p)
	if minCap >= cap(c.data) {
		aux := c.data
		c.data = make([]byte, minCap+(minCap*1/3))
		copy(c.data, aux)
	}
	for i := range p {
		c.data[c.offset] = p[i]
		c.offset++
	}
	return len(p), nil
}

func (c *CustomWriter) writeNumber(p any) (n int, err error) {
	old := c.offset
	er := binary.Write(c, c.order, p)
	return c.offset - old, er
}
func (c *CustomWriter) WriteString(p string) (n int, err error) {
	old := c.offset
	c.Write([]byte(p))
	if err != nil {
		return 0, err
	}
	return c.offset - old, nil
}
func (c *CustomWriter) SetOffSet(offSet int) int {
	if offSet < cap(c.data) {
		c.offset = offSet
	}

	return c.offset
}

func (c *CustomWriter) OffSet() int {
	return c.offset
}

func (c *CustomWriter) Len() int {
	//len(c.data)
	return c.offset
}

func (c *CustomWriter) Data() []byte {
	return c.data[:c.offset]
}

func (c *CustomWriter) WriteByte(p byte) (n int, err error) {
	return c.writeNumber(p)
}
func (c *CustomWriter) WriteBool(p bool) (n int, err error) {
	return c.writeNumber(p)
}
func (c *CustomWriter) WriteInt8(p int8) (n int, err error) {
	return c.writeNumber(p)
}
func (c *CustomWriter) WriteUInt8(p uint8) (n int, err error) {
	return c.writeNumber(p)
}
func (c *CustomWriter) WriteInt16(p int16) (n int, err error) {
	return c.writeNumber(p)
}
func (c *CustomWriter) WriteUInt16(p uint16) (n int, err error) {
	return c.writeNumber(p)
}
func (c *CustomWriter) WriteInt32(p int32) (n int, err error) {
	return c.writeNumber(p)
}
func (c *CustomWriter) WriteUInt32(p uint32) (n int, err error) {
	return c.writeNumber(p)
}
func (c *CustomWriter) WriteInt64(p int64) (n int, err error) {
	return c.writeNumber(p)
}
func (c *CustomWriter) WriteUInt64(p int8) (n int, err error) {
	return c.writeNumber(p)
}
func (c *CustomWriter) WriteFloat32(p float32) (n int, err error) {
	return c.writeNumber(p)
}
func (c *CustomWriter) WriteFloat64(p float64) (n int, err error) {
	return c.writeNumber(p)
}

/*
case bool, int8, uint8, int16, uint16, int16, int32, uint32, int64, uint64, float32,float64
*/
