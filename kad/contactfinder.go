package kad

import (
	"container/list"
)

const contactFindNbr = 5 // how many contacts used to find more nodes

// ContactFinder x
type ContactFinder struct {
	pList      *list.List                  // we'll take it as a stack
	contactMap map[uint32]*ContactListNode // key is IP

	pPacketProcessor *PacketProcessor
}

func (cf *ContactFinder) start(pPacketProcessor *PacketProcessor) {
	cf.pPacketProcessor = pPacketProcessor

	cf.pList = list.New()
	cf.contactMap = make(map[uint32]*ContactListNode)
}

func (cf *ContactFinder) add(pContact *Contact) {
	// Check existing or not, if existing, just return.
	// It's tricky that contact share the same memory. So content of contact will be changed by other messages.
	pc := cf.contactMap[pContact.ip]
	if pc != nil {
		return
	}

	node := ContactListNode{pContact: pContact}

	// add into map
	cf.contactMap[pContact.ip] = &node

	// add into list
	e := cf.pList.PushBack(&node)
	node.e = e
}

func (cf *ContactFinder) remove(pContact *Contact) {
	// find contact node
	pNode := cf.contactMap[pContact.ip]
	if pNode == nil {
		return
	}

	cf.pList.Remove(pNode.e)
}

func (cf *ContactFinder) tickProcess() {
	// pop contacts and then move to bottom
	listLen := cf.pList.Len()
	for i, c := 0, 0; c < contactFindNbr && i < listLen; i++ {
		e := cf.pList.Back()
		cf.pList.MoveToFront(e)

		pContact := e.Value.(*ContactListNode).pContact
		if pContact.pKadID != nil && pContact.getVersion() >= kademliaVersion2_47a {
			// send kademlia2Req
			targetID := ID{} // random target
			targetID.generate()
			cf.pPacketProcessor.sendFindValue(pContact, &targetID)
			c++
		}
	}
}
