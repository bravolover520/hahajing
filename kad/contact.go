package kad

// how many RTT samples for statistic
const contactRTTSize = 10
const contactHelloReqTimerOut = 5 // second

// Contact x
type Contact struct {
	pKadID  *ID // KAD ID, which will be used to explore more contacts from this contact
	ip      uint32
	updPort uint16

	// It has a verify key used for I sending packet to this contact, who can use it to verify.
	// Verify key is bound to my public IP.
	// If this contact offline and then online or I change my public IP, usually verify key will be invalid.
	// Verify key is composited by 2 parts: remote IP and random 32bit key.
	udpKey UDPKey

	version   uint8
	bVerified bool // Is it verified via HelloReq?

	tLiveExpires int64
	tCreated     int64
	tDead        int64
	tLastLive    int64

	tHelloReq int64 // time sending HelloReq
	tRTT      int64 // Round trip time via hello message
	tRTTs     []int64
}

func (c *Contact) getVersion() uint8 {
	return c.version
}

func (c *Contact) getKadID() *ID {
	return c.pKadID
}

func (c *Contact) getVerifyKey(ip uint32) uint32 {
	return c.udpKey.getKeyValue(ip)
}

func (c *Contact) getIP() uint32 {
	return c.ip
}

func (c *Contact) getUDPPort() uint16 {
	return c.updPort
}

func (c *Contact) resetUDPKey() {
	c.udpKey.reset()
}

func (c *Contact) setLiveExpiresTime(t int64, long bool) {
	if long {
		// Must consider possible HelloReq timeout so that PacketReqGuard can pass it.
		// We use random expires so that HelloReq requests can be distributed averagely in time slots.
		c.tLiveExpires = t + int64(random8()%10) + packetReqLimitTime + 1
	} else {
		c.tLiveExpires = t + contactHelloReqTimerOut
	}
}

// Should be called after setLiveExpiresTime
func (c *Contact) setDeadTime(long bool) {
	if long {
		c.tDead = c.tLiveExpires + contactHelloReqTimerOut*3
	} else {
		c.tDead = c.tLiveExpires + contactHelloReqTimerOut*2
	}
}

func (c *Contact) setHelloReqTime(t int64) {
	c.tHelloReq = t
}

func (c *Contact) setRTT(t int64) {
	if t < c.tHelloReq {
		return
	}

	rtt := t - c.tHelloReq
	c.tRTTs = append(c.tRTTs, rtt)

	if len(c.tRTTs) > contactRTTSize {
		c.tRTTs = c.tRTTs[1:]
	}

	// calculate average RTT
	rtt = 0
	for _, i := range c.tRTTs {
		rtt += i
	}

	c.tRTT = rtt / int64(len(c.tRTTs))
}

func (c *Contact) setCreatedime(t int64) {
	c.tCreated = t
}
