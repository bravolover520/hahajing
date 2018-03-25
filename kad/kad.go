package kad

import (
	"time"
)

const (
	kadTimer               = 1
	kadPacketReqGuardTimer = 60

	kadSearchReqChSize = 1000
)

// Kad x
type Kad struct {
	prefs           Prefs
	contactManager  ContactManager
	packetProcesser PacketProcessor
	packetReqGuard  PacketReqGuard
	searchManager   SearchManager

	socketManager  SocketManager
	recvCh, sendCh chan *Packet

	// externs
	SearchReqCh chan *SearchReq
}

// Start x
func (k *Kad) Start() bool {
	k.SearchReqCh = make(chan *SearchReq, kadSearchReqChSize)

	socketChSize := bootstrapSearchContactNbr * int(kademliaFindNode) * int(kademliaFindNode)
	k.recvCh = make(chan *Packet, socketChSize)
	k.sendCh = make(chan *Packet, socketChSize)

	// start should be from bottom to up layer
	k.prefs.start()
	k.socketManager.start(&k.prefs, k.recvCh, k.sendCh)
	k.packetReqGuard.start()
	k.packetProcesser.start(&k.prefs, &k.contactManager, &k.searchManager, &k.packetReqGuard, k.sendCh)
	k.searchManager.start(&k.packetProcesser, &k.contactManager.onliner)

	k.contactManager.start(&k.prefs, &k.packetProcesser, &k.packetReqGuard)

	go k.scheduleRoutine()

	return true
}

func (k *Kad) scheduleRoutine() {
	tick := time.NewTicker(kadTimer * time.Second)
	packetReqGuardTimer := time.NewTicker(kadPacketReqGuardTimer * time.Second)

	for {
		select {
		case pPacket := <-k.recvCh:
			k.packetProcesser.processPacket(pPacket)
		case <-tick.C:
			k.contactManager.tickProcess()
			k.searchManager.tickProcess()
		case <-packetReqGuardTimer.C:
			k.packetReqGuard.timerProcess()

		case pSearchReq := <-k.SearchReqCh:
			k.searchManager.newSearch(pSearchReq)
		}
	}
}
