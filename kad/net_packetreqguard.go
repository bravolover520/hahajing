package kad

import (
	"time"
)

const packetReqLimitTime = 60 // second

// limits of Kademlia requests per time interval
var packetReqLimits = map[byte]int{
	kademlia2HelloReq:     3,
	kademlia2Req:          10,
	kademlia2SearchKeyReq: 3}

// PacketReqPerIP is KAD requests counting per IP.
type PacketReqPerIP struct {
	reqs map[byte][]int64 // opcode: [time]
}

func (p *PacketReqPerIP) canPass(t int64, opcode byte) (int, bool) {
	times := p.reqs[opcode]
	limit := packetReqLimits[opcode]
	count := 1 // assume this one is added.
	i := len(times) - 1
	for ; i >= 0; i-- {
		if t-times[i] > packetReqLimitTime {
			break
		}

		count++
		if count > limit {
			return 0, false
		}
	}

	return i, true // -1: empty
}

func (p *PacketReqPerIP) add(t int64, opcode byte) bool {
	i, pass := p.canPass(t, opcode)
	if !pass {
		return false
	}

	times := append(p.reqs[opcode], t)

	// cut, we don't need more
	if i < 0 {
		i = 0
	}
	p.reqs[opcode] = times[i:]

	return true
}

// PacketReqGuard is guard for monitoring each KAD request for each remote IP so that remote KAD client will not drop our packets.
// Here IPs are not synchronized with these in ContactManager. They're independ.
type PacketReqGuard struct {
	reqs map[uint32]*PacketReqPerIP // remote IP: *PacketReqPerIP

	curTime     int64
	trackReqs   map[uint32]int64          // remote IP: request time
	expiresReqs map[int64]map[uint32]bool // time: remote IP
}

func (g *PacketReqGuard) start() {
	g.reqs = make(map[uint32]*PacketReqPerIP)

	g.trackReqs = make(map[uint32]int64)
	g.expiresReqs = make(map[int64]map[uint32]bool)
}

func (g *PacketReqGuard) add(t int64, remoteIP uint32, opcode byte) bool {
	// we only care about requests
	if opcode != kademlia2HelloReq &&
		opcode != kademlia2Req &&
		opcode != kademlia2SearchKeyReq {
		return true
	}

	if len(g.expiresReqs) == 0 {
		g.curTime = t + 1
	}

	reqs := g.reqs[remoteIP]
	if reqs == nil {
		reqs = &PacketReqPerIP{reqs: make(map[byte][]int64)}
		g.reqs[remoteIP] = reqs
	}

	if !reqs.add(t, opcode) {
		return false
	}

	// remove
	expiresTime, ok := g.trackReqs[remoteIP]
	if ok {
		ips := g.expiresReqs[expiresTime]
		delete(ips, remoteIP)
		if len(ips) == 0 {
			delete(g.expiresReqs, expiresTime)
		}
	}

	// add
	expiresTime = t + packetReqLimitTime
	g.trackReqs[remoteIP] = expiresTime

	ips := g.expiresReqs[expiresTime]
	if ips == nil {
		ips = make(map[uint32]bool)
		g.expiresReqs[expiresTime] = ips
	}
	ips[remoteIP] = true

	return true
}

func (g *PacketReqGuard) timerProcess() {
	if len(g.expiresReqs) == 0 {
		return
	}

	t := time.Now().Unix()

	for ; g.curTime <= t; g.curTime++ {
		ips := g.expiresReqs[g.curTime]
		for ip := range ips {
			delete(g.trackReqs, ip)
			delete(g.reqs, ip)
		}

		delete(g.expiresReqs, g.curTime)
	}
}

func (g *PacketReqGuard) canPass(t int64, remoteIP uint32, opcode byte) bool {
	reqs := g.reqs[remoteIP]
	if reqs == nil {
		return true
	}

	_, pass := reqs.canPass(t, opcode)
	return pass
}
