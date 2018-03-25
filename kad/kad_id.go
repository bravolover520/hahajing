package kad

import "encoding/binary"

// ID is KAD ID
type ID struct {
	hash [16]byte
}

func (id *ID) generate() {
	high := uint64ToByte(random64())
	low := uint64ToByte(random64())

	copy(id.hash[:8], low)
	copy(id.hash[8:], high)
}

func (id *ID) getHash() []byte {
	return id.hash[:]
}

func (id *ID) setHash(hash []byte) {
	copy(id.hash[:], hash)
}

func (id *ID) get() [16]byte {
	return id.hash
}

func (id *ID) getXor(pTargetID *ID) *ID {
	xorID := ID{}
	for i := 0; i < 16; i++ {
		xorID.hash[i] = id.hash[i] ^ pTargetID.hash[i]
	}

	return &xorID
}

func (id *ID) get32BitChunk(i int) uint32 {
	return binary.LittleEndian.Uint32(id.hash[i*4 : (i+1)*4])
}
