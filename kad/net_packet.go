package kad

import (
	"bytes"
	"compress/zlib"
	"io"
)

// Packet is control packet, whose eDonkeyID is opKademliaHeader.
type Packet struct {
	pKadID *ID
	ip     uint32 // for sending, it's destination IP, for receiving, it's source IP.
	port   uint16 // UDP port

	opcode byte
	buf    []byte // exclude UDP header(2 bytes: eDonkeyID, opcode)

	receiverVerifyKey uint32

	// For example, I send this packet to destination.
	// @senderVerifyKey will be my UDP Key(random generated during start, don't be confused with struct UDPKey used for Contact) hashed with destination IP by MD5.
	// Destination will fill it as receiver verify key so that I can verify it's destination I sent before.
	// For what meaning of receiver verify key or sender verify key, pay attention to the direction of the packet.
	senderVerifyKey uint32
}

// For send
func (pPacket *Packet) getBuf() []byte {
	buf := make([]byte, len(pPacket.buf)+2)
	buf[0] = opKademliaHeader
	buf[1] = pPacket.opcode

	copy(buf[2:], pPacket.buf)

	return buf
}

func uncompress(compressSrc []byte) (bool, []byte) {
	b := bytes.NewReader(compressSrc)
	r, err := zlib.NewReader(b)
	if err != nil {
		return false, nil
	}

	var out bytes.Buffer
	_, err = io.Copy(&out, r)
	if err != nil {
		return false, nil
	}

	return true, out.Bytes()
}

func doOpKademliaPackedProt(buf []byte) (bool, []byte) {
	if len(buf) == 2 {
		return true, buf
	}

	ok, dstBuf := uncompress(buf[2:])
	if !ok {
		return false, nil
	}

	// restore to org buf format
	newBuf := make([]byte, len(dstBuf)+2)
	copy(newBuf[2:], dstBuf)

	newBuf[0] = opKademliaHeader
	newBuf[1] = buf[1]

	return true, newBuf
}

// For receive
func (pPacket *Packet) setBuf(buf []byte) bool {
	if len(buf) < 2 {
		return false
	}

	if buf[0] != opKademliaHeader && buf[0] != opKademliaPackedProt {
		return false
	}

	if buf[1] != kademlia2HelloRes &&
		buf[1] != kademlia2Res &&
		buf[1] != kademlia2SearchRes {
		return false
	}

	if buf[0] == opKademliaPackedProt {
		var ok bool
		ok, buf = doOpKademliaPackedProt(buf)
		if !ok {
			return false
		}
	}

	pPacket.opcode = buf[1]
	if len(buf) > 2 {
		pPacket.buf = buf[2:]
	} else {
		pPacket.buf = nil
	}

	return true
}
