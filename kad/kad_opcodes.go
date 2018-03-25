package kad

import "strconv"

const (
	// KAD version
	kademliaVersion1_46c     uint8 = 0x01 /*45b - 46c*/
	kademliaVersion2_47a     uint8 = 0x02 /*47a*/
	kademliaVersion3_47b     uint8 = 0x03 /*47b*/
	kademliaVersion5_48a     uint8 = 0x05 // -0.48a
	kademliaVersion6_49aBeta uint8 = 0x06 // -0.49aBETA1, needs to support: OP_FWCHECKUDPREQ (!), obfuscation, direct callbacks, source type 6, UDP firewallcheck
	kademliaVersion7_49a     uint8 = 0x07 // -0.49a needs to support OP_KAD_FWTCPCHECK_ACK, KADEMLIA_FIREWALLED2_REQ
	kademliaVersion8_49b     uint8 = 0x08 // TAG_KADMISCOPTIONS, KADEMLIA2_HELLO_RES_ACK
	kademliaVersion9_50a     uint8 = 0x09 // handling AICH hashes on keyword storage
	kademliaVersion          uint8 = 0x09 // Change CT_EMULE_MISCOPTIONS2 if Kadversion becomes >= 15 (0x0F)

	// eDonkeyID
	opKademliaHeader     byte = 0xE4 // eDonkeyID, for control packet, which starts with eDonkeyID followed by opcode
	opKademliaPackedProt byte = 0xE5
	opPackedPort         byte = 0xD4
	opEmulePort          byte = 0xC5
	opUDPReservedPort1   byte = 0xA3 // reserved for later UDP headers (important for EncryptedDatagramSocket)
	opUDPReservedPort2   byte = 0xB2 // reserved for later UDP headers (important for EncryptedDatagramSocket)

	// KADEMLIA (opcodes) (udp)
	kademlia2HelloReq    byte = 0x11
	kademlia2HelloRes    byte = 0x19 //
	kademlia2HelloResAck byte = 0x22 // <NodeID><uint8 tags>

	kademlia2Req byte = 0x21 // use to find nodes via target(KAD ID)
	kademlia2Res byte = 0x29 //

	kademlia2SearchKeyReq    byte = 0x33 // search keyword
	kademlia2SearchSourceReq byte = 0x34 //
	kademlia2SeachNotesReq   byte = 0x35 //
	kademlia2SearchRes       byte = 0x3B //

	kademlia2PublishKeyReq    byte = 0x43 //
	kademlia2PublishSourceReq byte = 0x44 //
	kademlia2PublishNotesReq  byte = 0x45 //

	kademliaFirewalledReq  byte = 0x50 // <TCPPORT (sender) [2]>
	kademliaFirewalled2Req byte = 0x53 // <TCPPORT (sender) [2]><userhash><connectoptions 1>
	kademliaFirewalledRes  byte = 0x58 // <IP (sender) [4]>, both REQs will use as the RES. We use to explore our external IP.

	kademlia2Ping byte = 0x60 // (null)
	kademlia2Pong byte = 0x61 // (null), we use to explore our external UDP port.

	// KADEMLIA (parameter), used in kademlia2RES
	kademliaFindValue     uint8 = 0x02
	kademliaFindNode      uint8 = 0x0B
	kademliaFindValueMore uint8 = kademliaFindNode
)

const minSupportContactVersion = kademliaVersion3_47b

func getOpcodeStr(opcode byte) string {
	var opcodeStr = strconv.Itoa(int(opcode))
	switch opcode {
	case kademlia2HelloReq:
		opcodeStr = "kademlia2HelloReq"
	case kademlia2HelloRes:
		opcodeStr = "kademlia2HelloRes"
	case kademlia2HelloResAck:
		opcodeStr = "kademlia2HelloResAck"

	case kademlia2Req:
		opcodeStr = "kademlia2Req"
	case kademlia2Res:
		opcodeStr = "kademlia2Res"

	case kademlia2SearchKeyReq:
		opcodeStr = "kademlia2SearchKeyReq"
	case kademlia2SearchRes:
		opcodeStr = "kademlia2SearchRes"
	}

	return opcodeStr
}

func getVersionStr(version uint8) string {
	var versionStr = strconv.Itoa(int(version))
	switch version {
	case kademliaVersion1_46c:
		versionStr = "V46c"
	case kademliaVersion2_47a:
		versionStr = "V47a"
	case kademliaVersion3_47b:
		versionStr = "V47b"
	case kademliaVersion5_48a:
		versionStr = "V48a"
	case kademliaVersion6_49aBeta:
		versionStr = "V49aBeta"
	case kademliaVersion7_49a:
		versionStr = "V49a"
	case kademliaVersion8_49b:
		versionStr = "V49b"
	case kademliaVersion9_50a:
		versionStr = "V50a"
	}

	return versionStr
}
