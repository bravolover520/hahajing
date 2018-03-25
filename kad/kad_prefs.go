package kad

import (
	"hahajing/com"
	"net"
	"time"
)

const localUDPPort = 1979
const localTCPPort = 1988

// Prefs is my preferences.
type Prefs struct {
	kadID ID

	tcpPort uint16 // useless only for sending packet to client

	udpKey uint32 // used to generate my verify key with destination IP for sending packet. Please note difference with struct UDPKey.

	bFirewalled   bool
	externIP      uint32
	externUDPPort uint16

	localIP      uint32
	localUDPPort uint16 // used to UDP connection listen, if not firewalled, it's same as @externUDPPort

	tLastContact int64 // time of last packet I received from other client, I use it to track if I'm still online
}

func (p *Prefs) start() {
	p.kadID.generate()
	p.udpKey = random32()
	p.tcpPort = localTCPPort

	p.localUDPPort = localUDPPort

	p.initLocalIP()
}

func (p *Prefs) getUDPVerifyKey(targetIP uint32) uint32 {
	ui64Buffer := uint64(p.udpKey)
	ui64Buffer <<= 32
	ui64Buffer |= uint64(targetIP)

	md5 := Md5Sum{}
	md5.calculate(uint64ToByte(ui64Buffer))

	rawHash := md5.getRawHash()
	ui32Hash := byteToUint32Slice(rawHash)

	key := ui32Hash[0]
	for _, hash := range ui32Hash[1:] {
		key ^= hash
	}
	return key%0xFFFFFFFE + 1
}

func (p *Prefs) getKadID() *ID {
	return &p.kadID
}

func (p *Prefs) getPublicIP() uint32 {
	return p.externIP
}

func (p *Prefs) getLocalUDPPort() uint16 {
	return p.localUDPPort
}

func (p *Prefs) getTCPPort() uint16 {
	return p.tcpPort
}

func (p *Prefs) setLastContact() {
	p.tLastContact = time.Now().Unix()
}

func (p *Prefs) getMyConnectOptions() uint8 {
	// Connect options Tag
	// 4 Reserved (!)
	// 1 Direct Callback
	// 1 CryptLayer Required
	// 1 CryptLayer Requested
	// 1 CryptLayer Supported
	const uSupportsCryptLayer uint8 = 1
	const uRequestsCryptLayer uint8 = 0
	const uRequiresCryptLayer uint8 = 0
	// direct callback is only possible if connected to kad, tcp firewalled and verified UDP open (for example on a full cone NAT)
	const uDirectUDPCallback uint8 = 0

	return (uDirectUDPCallback << 3) | (uRequiresCryptLayer << 2) | (uRequestsCryptLayer << 1) | (uSupportsCryptLayer << 0)
}

func (p *Prefs) setPublicIP(ip uint32) {
	p.externIP = ip

	if p.externIP != p.localIP {
		p.bFirewalled = true
	} else {
		p.bFirewalled = false
	}
}

// Get preferred outbound ip of this machine
func (p *Prefs) initLocalIP() {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		com.HhjLog.Critical(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	com.HhjLog.Infof("Local outbound IP: %s\n", localAddr.IP.String())

	p.localIP = ip2I(localAddr.IP)
}
