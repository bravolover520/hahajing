package kad

import (
	"encoding/binary"
)

// Kademlia2HelloResMsg x
type Kademlia2HelloResMsg struct {
	contactID ID
	ip        uint32
	udpPort   uint16
	version   uint8
	verifyKey uint32
}

func (m *Kademlia2HelloResMsg) set(pPacket *Packet) bool {
	// from packet
	m.ip = pPacket.ip
	m.udpPort = pPacket.port
	m.verifyKey = pPacket.senderVerifyKey

	// decode from buf
	bi := ByteIO{buf: pPacket.buf}

	if !bi.check(16 + 2 + 1) {
		return false
	}
	bi.readBytesFast(m.contactID.getHash())
	bi.readUint16() // TCP port
	m.version = bi.readUint8()

	return true
}

// Kademlia2ResMsg x
type Kademlia2ResMsg struct {
	// From contact who respond this message
	ip        uint32
	udpPort   uint16
	verifyKey uint32

	// message content
	targetID ID // target what we're looking for
	contacts []*Contact
}

func (m *Kademlia2ResMsg) set(pPacket *Packet) bool {
	// from packet
	m.ip = pPacket.ip
	m.udpPort = pPacket.port
	m.verifyKey = pPacket.senderVerifyKey

	// decode from buf
	bi := ByteIO{buf: pPacket.buf}

	if !bi.check(16) {
		return false
	}
	bi.readBytesFast(m.targetID.getHash())

	if !bi.check(1) {
		return false
	}
	contactNbr := bi.readUint8()

	// Verify packet is expected size
	if len(pPacket.buf) != 16+1+(16+4+2+2+1)*int(contactNbr) {
		//com.HhjLog.Noticef("Received wrong size(%d) of kademlia2Res packet from %s:%d", len(pPacket.buf), iIP2Str(pPacket.ip), pPacket.port)
		return false
	}

	// loop for each contact
	for i := 0; i < int(contactNbr); i++ {
		kadID := ID{}
		bi.readBytesFast(kadID.getHash())
		ip := bi.readUint32()
		udpPort := bi.readUint16()
		bi.readUint16() // TCP Port
		version := bi.readUint8()

		if version < kademliaVersion2_47a {
			continue
		}

		contact := Contact{
			pKadID:  &kadID,
			ip:      ip,
			updPort: udpPort,
			version: version}

		m.contacts = append(m.contacts, &contact)
	}

	return true
}

// Kademlia2SearchResMsg x
type Kademlia2SearchResMsg struct {
	// From contact who respond this message
	ip        uint32
	udpPort   uint16
	verifyKey uint32

	// message content
	contactID ID // contact ID who send us this RES
	targetID  ID // target what we're looking for(keyword hash)

	// files
	files []*Ed2kFileStruct
}

func (m *Kademlia2SearchResMsg) set(pPacket *Packet) bool {
	// from packet
	m.ip = pPacket.ip
	m.udpPort = pPacket.port
	m.verifyKey = pPacket.senderVerifyKey

	// decode from buf
	bi := ByteIO{buf: pPacket.buf}

	if !bi.check(16 + 16) {
		return false
	}
	bi.readBytesFast(m.contactID.getHash()) // contact ID
	bi.readBytesFast(m.targetID.getHash())  // target ID

	m.setFiles(&bi)

	return true
}

func (m *Kademlia2SearchResMsg) setFiles(bi *ByteIO) bool {
	if !bi.check(2) {
		return false
	}
	count := bi.readUint16() // total results, how many files

	for ; count > 0; count-- {
		if !bi.check(16) {
			return false
		}
		// read all file related infos
		fileHash := bi.readBytes(16)

		pTags := readTags(bi)
		if pTags == nil {
			return false
		}

		// convert to file struct
		fileStruct := Ed2kFileStruct{}
		m.setFileParams(&fileStruct, *pTags)
		copy(fileStruct.Hash[:], fileHash)

		// add into message struct
		m.files = append(m.files, &fileStruct)
	}

	return true
}

func (m *Kademlia2SearchResMsg) setFileSize(pFileStruct *Ed2kFileStruct, pTag *Tag) {
	switch pTag.value.(type) {
	case []byte:
		v := pTag.value.([]byte)
		if len(v) == 8 {
			pFileStruct.Size = binary.LittleEndian.Uint64(v)
		}
	case uint64:
		v := pTag.value.(uint64)
		pFileStruct.Size = uint64(v)
	case uint32:
		v := pTag.value.(uint32)
		pFileStruct.Size = uint64(v)
	case uint16:
		v := pTag.value.(uint16)
		pFileStruct.Size = uint64(v)
	case uint8:
		v := pTag.value.(uint8)
		pFileStruct.Size = uint64(v)
	}
}

func (m *Kademlia2SearchResMsg) setFileParams(pFileStruct *Ed2kFileStruct, tags []*Tag) {

	for _, pTag := range tags {
		switch pTag.name {
		case tagFileName:
			pFileStruct.Name = pTag.value.(string)
		case tagFileSize:
			m.setFileSize(pFileStruct, pTag)
		case tagFileType:
			pFileStruct.Type = pTag.value.(string)
		case tagSources:
			pFileStruct.Avail = void2Uint32(pTag.value)
		case tagMediaLength:
			pFileStruct.MediaLength = void2Uint32(pTag.value)
		}
	}
}
