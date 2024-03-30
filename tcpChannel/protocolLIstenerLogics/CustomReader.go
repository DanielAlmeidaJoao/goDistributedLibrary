package protocolLIstenerLogics

import (
	"encoding/binary"
)

type MESSAGE_DESERIALIZER_TYPE func(reader *CustomReader) any

/**
type NetworkMessageDeserializer interface {
	DeserializeData(reader *CustomReader) NetworkMessage //make it return the
} **/

type CustomReader struct {
	data   []byte
	offset int
	order  binary.ByteOrder
}

func (c *CustomReader) Read(p []byte) (n int, err error) {
	c.offset += copy(p, c.data[c.offset:])
	return c.offset, nil
}

/*
*
Writes any number like int32, uint8, ... to the buffer. Numbers like int or float are not accepted
*/
func (b *CustomReader) ReadAnyNumberWithSizeOnTheName(dest any) error {
	err := binary.Read(b, b.order, dest)
	return err
}

func (b *CustomReader) ReadTheRest() []byte {
	return b.data[b.offset:]
}
func (b *CustomReader) ReadString(n int) (string, error) {
	//TODO check for error
	aux := b.data[b.offset : b.offset+n]
	b.offset += len(aux)
	return string(aux), nil
}

func (b *CustomReader) SetOffset(offset int) {
	b.offset = offset
}

func NewCustomReader(data []byte, order binary.ByteOrder) *CustomReader {
	return &CustomReader{
		data:   data,
		offset: 0,
		order:  order,
	}
}

/*
case bool, int8, uint8, int16, uint16, int16, int32, uint32, int64, uint64, float32,float64
*/
func (b *CustomReader) ReadByte() (byte, error) {
	var num byte
	err := binary.Read(b, b.order, &num)
	return num, err
}
func (b *CustomReader) ReadBool() (bool, error) {
	var num bool
	err := binary.Read(b, b.order, &num)
	return num, err
}
func (b *CustomReader) ReadInt8() (int8, error) {
	var num int8
	err := binary.Read(b, b.order, &num)
	return num, err
}
func (b *CustomReader) ReadUint8() (uint8, error) {
	var num uint8
	err := binary.Read(b, b.order, &num)
	return num, err
}
func (b *CustomReader) ReadInt16() (int16, error) {
	var num int16
	err := binary.Read(b, b.order, &num)
	return num, err
}

func (b *CustomReader) ReadUint16() (uint16, error) {
	var num uint16
	err := binary.Read(b, b.order, &num)
	return num, err
}
func (b *CustomReader) ReadInt32() (int32, error) {
	var num int32
	err := binary.Read(b, b.order, &num)
	return num, err
}
func (b *CustomReader) ReadUint32() (uint32, error) {
	var num uint32
	err := binary.Read(b, b.order, &num)
	return num, err
}
func (b *CustomReader) ReadInt64() (int64, error) {
	var num int64
	err := binary.Read(b, b.order, &num)
	return num, err
}
func (b *CustomReader) ReadUInt64() (uint64, error) {
	var num uint64
	err := binary.Read(b, b.order, &num)
	return num, err
}
func (b *CustomReader) ReadFloat32() (float32, error) {
	var num float32
	err := binary.Read(b, b.order, &num)
	return num, err
}
func (b *CustomReader) ReadFloat64() (float64, error) {
	var num float64
	err := binary.Read(b, b.order, &num)
	return num, err
}
