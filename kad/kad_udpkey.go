package kad

import (
	"encoding/binary"
	"io"
)

// UDPKey is used to connect me with client. It will be used when I send packet to client.
// It's bound to my public IP. If I change my public IP(e.g. NAT case) or client change key used to calculate its verify key, it will be invalid.
type UDPKey struct {
	key uint32 // contact verify key
	ip  uint32 // My public IP. In eMule source code, it's a little confusing with UDP key used to calculate verify key.
}

func (k *UDPKey) readFromFile(r io.Reader) {
	binary.Read(r, binary.LittleEndian, &k.key)
	binary.Read(r, binary.LittleEndian, &k.ip)
}

func (k *UDPKey) getKeyValue(ip uint32) uint32 {
	if ip == k.ip {
		return k.key
	}

	return 0
}

func (k *UDPKey) reset() {
	k.key = 0
	k.ip = 0
}
