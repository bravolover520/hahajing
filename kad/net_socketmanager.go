package kad

const (
	socketNbr    = 10
	updPortStart = 2000
)

// SocketManager is manager of sockets for distributing sending packets to sockets by round robin.
type SocketManager struct {
	sockets []*Socket
	round   int

	sendCh chan *Packet
}

func (s *SocketManager) start(pPrefs *Prefs, recvCh, sendCh chan *Packet) bool {
	s.sendCh = sendCh // channel for sending packets

	// start sockets
	for i := 0; i < socketNbr; i++ {
		socket := &Socket{no: i}
		sendCh1 := make(chan *Packet, cap(sendCh)/socketNbr)
		udpPort := uint16(updPortStart + i)
		if !socket.start(pPrefs, recvCh, sendCh1, udpPort) {
			return false
		}

		s.sockets = append(s.sockets, socket)
	}

	// loop to distribute sending packets
	go s.sendRoutine()

	return true
}

func (s *SocketManager) sendRoutine() {
	for {
		packet := <-s.sendCh
		s.sockets[s.round].sendCh <- packet

		s.round++
		s.round = s.round % socketNbr
	}
}
