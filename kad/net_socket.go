package kad

import (
	"crypto/rc4"
	"encoding/binary"
	"hahajing/com"
	"net"
)

const (
	cryptHeaderWithoutPadding        = 8
	magicValueUDPSyncClient   uint32 = 0x395F2EC1
)

// Socket is KAD UDP socket with encryption
type Socket struct {
	no int // No.

	conn           *net.UDPConn
	recvCh, sendCh chan *Packet

	pPrefs *Prefs
}

func (s *Socket) start(pPrefs *Prefs, recvCh, sendCh chan *Packet, udpPort uint16) bool {
	s.pPrefs = pPrefs
	s.recvCh, s.sendCh = recvCh, sendCh

	// init a UDP socket
	var err error
	s.conn, err = net.ListenUDP("udp", &net.UDPAddr{IP: nil, Port: int(udpPort)})
	if err != nil {
		com.HhjLog.Criticalf("Socket: Listen on UDP error: %s", err)
		return false
	}

	// loop to send and receive
	go s.recvRoutine()
	go s.sendRoutine()

	return true
}

func (s *Socket) send(pPacket *Packet) {
	sendbuffer := pPacket.getBuf()
	//sendbuffer = s.encrypt(sendbuffer, pPacket.pKadID, pPacket.receiverVerifyKey, pPacket.senderVerifyKey)
	if sendbuffer == nil {
		return
	}

	remoteAddr := &net.UDPAddr{IP: i2IP(pPacket.ip), Port: int(pPacket.port)}
	_, err := s.conn.WriteToUDP(sendbuffer, remoteAddr)
	if err != nil {
		com.HhjLog.Errorf("Socket: Write to UDP %s:%d error: %s\n", iIP2Str(pPacket.ip), pPacket.port, err)
		return
	}

	socketLog("Socket: Send packet to %s:%d\n", iIP2Str(pPacket.ip), pPacket.port)
}

func (s *Socket) recv() {
	buf := make([]byte, 5000)
	n, remoteAddr, err := s.conn.ReadFromUDP(buf)
	if err != nil {
		//com.HhjLog.Warningf("Socket: Read from UDP error: %s\n", err)
		return
	}

	remoteIP := ip2I(remoteAddr.IP)
	remotePort := uint16(remoteAddr.Port)

	buf, nReceiverVerifyKey, nSenderVerifyKey, bEncrypt := s.decrypt(buf[:n], remoteIP)
	if buf == nil {
		//com.HhjLog.Warningf("Socket: Decrypt received packet from %s:%d failed\n", iIP2Str(remoteIP), remotePort)
		return
	}

	// new packet
	packet := Packet{
		ip:                remoteIP,
		port:              remotePort,
		receiverVerifyKey: nReceiverVerifyKey,
		senderVerifyKey:   nSenderVerifyKey}

	if !packet.setBuf(buf) {
		return
	}

	if bEncrypt {
		socketLog("Socket: Receive encrypt packet from %s:%d\n", iIP2Str(remoteIP), remotePort)
	} else {
		socketLog("Socket: Receive non-encrypt packet from %s:%d\n", iIP2Str(remoteIP), remotePort)
	}

	s.recvCh <- &packet
}

func (s *Socket) encrypt(buf []byte, pClientID *ID, nReceiverVerifyKey, nSenderVerifyKey uint32) []byte {
	// no encrypt
	if pClientID == nil && nReceiverVerifyKey == 0 {
		return buf
	}

	// new crypted buffer
	nCryptHeaderLen := cryptHeaderWithoutPadding + 8
	nCryptedLen := len(buf) + nCryptHeaderLen
	pachCryptedBuffer := make([]byte, nCryptedLen)

	// generate MD5 hash used for RC4 cipher
	nRandomKeyPart := random16()
	bKadRecKeyUsed := false
	md5 := Md5Sum{}
	if pClientID == nil && nReceiverVerifyKey != 0 { // ecrypt by client receiver verify key
		bKadRecKeyUsed = true

		achKeyData := make([]byte, 6)
		binary.LittleEndian.PutUint32(achKeyData, nReceiverVerifyKey)
		binary.LittleEndian.PutUint16(achKeyData[4:], nRandomKeyPart)
		md5.calculate(achKeyData)

	} else { // more changes to ecrypt by client KAD ID
		achKeyData := make([]byte, 18)
		copy(achKeyData, pClientID.getHash())
		binary.LittleEndian.PutUint16(achKeyData[16:], nRandomKeyPart)
		md5.calculate(achKeyData)
	}

	/* header */
	// unciphered part
	if bKadRecKeyUsed {
		pachCryptedBuffer[0] = 2
	} else {
		pachCryptedBuffer[0] = 0
	}

	binary.LittleEndian.PutUint16(pachCryptedBuffer[1:3], nRandomKeyPart)

	// ciphered part
	cipher, err := rc4.NewCipher(md5.getRawHash())
	if err != nil {
		return nil
	}

	iu32Byte := make([]byte, 4)

	// magic value
	binary.LittleEndian.PutUint32(iu32Byte, magicValueUDPSyncClient)
	cipher.XORKeyStream(pachCryptedBuffer[3:7], iu32Byte)

	// pad length(0)
	cipher.XORKeyStream(pachCryptedBuffer[7:8], []byte{0})

	// receiver verify key
	binary.LittleEndian.PutUint32(iu32Byte, nReceiverVerifyKey)
	cipher.XORKeyStream(pachCryptedBuffer[8:12], iu32Byte)

	// sender verify key
	binary.LittleEndian.PutUint32(iu32Byte, nSenderVerifyKey)
	cipher.XORKeyStream(pachCryptedBuffer[12:16], iu32Byte)

	/* body */
	cipher.XORKeyStream(pachCryptedBuffer[16:], buf)

	return pachCryptedBuffer
}

func (s *Socket) decrypt(bufIn []byte, remoteIP uint32) ([]byte, uint32, uint32, bool) {
	if len(bufIn) <= cryptHeaderWithoutPadding {
		return nil, 0, 0, false
	}

	// no encrypted packet
	if bufIn[0] == opKademliaHeader ||
		bufIn[0] == opKademliaPackedProt ||
		bufIn[0] == opPackedPort ||
		bufIn[0] == opEmulePort ||
		bufIn[0] == opUDPReservedPort1 ||
		bufIn[0] == opUDPReservedPort2 {
		return bufIn, 0, 0, false
	}

	// might be an encrypted packet, try to decrypt
	// we only care about KAD packet
	var cipher *rc4.Cipher
	var err error
	ok := false
	for i := 0; i < 2; i++ {
		md5 := Md5Sum{}

		if i == 0 {
			// kad packet with NodeID as key
			achKeyData := make([]byte, 18)
			copy(achKeyData, s.pPrefs.getKadID().getHash())
			copy(achKeyData[16:], bufIn[1:3]) // random key part sent from remote client
			md5.calculate(achKeyData)
		} else {
			// kad packet with ReceiverKey as key
			achKeyData := make([]byte, 6)
			binary.LittleEndian.PutUint32(achKeyData[:4], s.pPrefs.getUDPVerifyKey(remoteIP))
			copy(achKeyData[4:], bufIn[1:3]) // random key part sent from remote client
			md5.calculate(achKeyData)
		}

		// try to decrypt
		cipher, err = rc4.NewCipher(md5.getRawHash())
		if err != nil {
			return nil, 0, 0, false
		}

		magicValueByte := make([]byte, 4)
		cipher.XORKeyStream(magicValueByte, bufIn[3:7])
		magicValue := binary.LittleEndian.Uint32(magicValueByte)

		// decrypt successfully
		if magicValue == magicValueUDPSyncClient {
			ok = true
			break
		}
	}

	if !ok {
		return nil, 0, 0, false
	}

	// pad length
	nResult := len(bufIn)

	byPadLenByte := []byte{0}
	cipher.XORKeyStream(byPadLenByte, bufIn[7:8])
	byPadLen := byPadLenByte[0]

	nResult -= cryptHeaderWithoutPadding
	if nResult <= int(byPadLen) {
		return nil, 0, 0, false
	}

	nResult -= int(byPadLen)
	if nResult <= 8 {
		return nil, 0, 0, false
	}

	// verify key
	var nReceiverVerifyKey, nSenderVerifyKey uint32
	ui32Byte := make([]byte, 4)
	start := cryptHeaderWithoutPadding + byPadLen

	cipher.XORKeyStream(ui32Byte, bufIn[start:start+4])
	nReceiverVerifyKey = binary.LittleEndian.Uint32(ui32Byte)

	cipher.XORKeyStream(ui32Byte, bufIn[start+4:start+8])
	nSenderVerifyKey = binary.LittleEndian.Uint32(ui32Byte)

	nResult -= 8

	// body
	bufOut := make([]byte, nResult)
	cipher.XORKeyStream(bufOut, bufIn[len(bufIn)-nResult:])

	return bufOut, nReceiverVerifyKey, nSenderVerifyKey, true
}

func (s *Socket) recvRoutine() {
	for {
		s.recv()
	}
}

func (s *Socket) sendRoutine() {
	for {
		packet := <-s.sendCh

		s.send(packet)
	}
}

func socketLog(format string, args ...interface{}) {
	if bEnableSocketLog {
		com.HhjLog.Infof(format, args...)
	}
}
