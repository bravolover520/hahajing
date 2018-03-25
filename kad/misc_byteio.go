package kad

import (
	"encoding/binary"
	"math"
)

// ByteIO x
type ByteIO struct {
	buf []byte
	pos int
}

func (b *ByteIO) getBuf() []byte {
	return b.buf[:b.pos]
}

func (b *ByteIO) check(leftSize int) bool {
	if b.pos+leftSize > len(b.buf) {
		return false
	}

	return true
}

func (b *ByteIO) writeByte(v byte) {
	b.buf[b.pos] = v

	b.pos++
}

func (b *ByteIO) writeUint8(v uint8) {
	b.buf[b.pos] = byte(v)

	b.pos++
}

func (b *ByteIO) writeUint16(v uint16) {
	binary.LittleEndian.PutUint16(b.buf[b.pos:b.pos+2], v)

	b.pos += 2
}

func (b *ByteIO) writeUint32(v uint32) {
	binary.LittleEndian.PutUint32(b.buf[b.pos:b.pos+4], v)

	b.pos += 4
}

func (b *ByteIO) writeUint64(v uint64) {
	binary.LittleEndian.PutUint64(b.buf[b.pos:b.pos+8], v)

	b.pos += 8
}

func (b *ByteIO) writeBytes(bytes []byte) {
	copy(b.buf[b.pos:], bytes)

	b.pos += len(bytes)
}

func (b *ByteIO) readByte() byte {
	v := b.buf[b.pos]
	b.pos++

	return v
}

func (b *ByteIO) readUint8() uint8 {
	v := b.buf[b.pos]
	b.pos++

	return uint8(v)
}

func (b *ByteIO) readUint16() uint16 {
	v := binary.LittleEndian.Uint16(b.buf[b.pos : b.pos+2])
	b.pos += 2

	return v
}

func (b *ByteIO) readUint32() uint32 {
	v := binary.LittleEndian.Uint32(b.buf[b.pos : b.pos+4])
	b.pos += 4

	return v
}

func (b *ByteIO) readFloat32() float32 {
	bits := binary.LittleEndian.Uint32(b.buf[b.pos : b.pos+4])
	b.pos += 4
	return math.Float32frombits(bits)
}

func (b *ByteIO) readUint64() uint64 {
	v := binary.LittleEndian.Uint64(b.buf[b.pos : b.pos+8])
	b.pos += 8

	return v
}

func (b *ByteIO) readBytes(n int) []byte {
	v := make([]byte, n)
	copy(v, b.buf[b.pos:b.pos+n])
	b.pos += n

	return v
}

func (b *ByteIO) readBytesFast(dstBuf []byte) {
	n := len(dstBuf)
	copy(dstBuf, b.buf[b.pos:b.pos+n])
	b.pos += n
}
