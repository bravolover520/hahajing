package kad

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

func min2s(min int) int64 {
	return int64(min * 60)
}

func hr2s(hr int) int64 {
	return int64(hr * 60 * 60)
}

func uint64ToByte(n uint64) []byte {
	b := make([]byte, 8)

	// little endian
	for i := 0; i < 8; i++ {
		b[i] = byte(n)
		n >>= 8
	}

	return b
}

func byteToUint32Slice(b []byte) []uint32 {
	n := make([]uint32, len(b)/4)

	// little endian
	for i := 0; i < len(n); i++ {
		n[i] = uint32(b[4*i+0]) | uint32(b[4*i+1])<<8 | uint32(b[4*i+2])<<16 | uint32(b[4*i+3])<<24
	}

	return n
}

func i2IP(ip uint32) net.IP {
	var bytes [4]byte
	bytes[0] = byte(ip)
	bytes[1] = byte((ip >> 8))
	bytes[2] = byte((ip >> 16))
	bytes[3] = byte((ip >> 24))

	// big endian
	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}

func ip2I(ip net.IP) uint32 {
	start := 0
	if len(ip) > 4 {
		start = 12
	}

	a, b, c, d := ip[start], ip[start+1], ip[start+2], ip[start+3]

	return uint32(a)<<24 | uint32(b)<<16 | uint32(c)<<8 | uint32(d)
}

func iIP2Str(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func random8() uint8 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return uint8(r.Int())
}

func random16() uint16 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return uint16(r.Int())
}

func random32() uint32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return uint32(r.Int31())
}

func random64() uint64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return uint64(r.Int63())
}

func void2Uint32(value interface{}) uint32 {
	switch value.(type) {
	case uint64:
		v := value.(uint64)
		return uint32(v)
	case uint32:
		v := value.(uint32)
		return uint32(v)
	case uint16:
		v := value.(uint16)
		return uint32(v)
	case uint8:
		v := value.(uint8)
		return uint32(v)
	}

	return 0
}
