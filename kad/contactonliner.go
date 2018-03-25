package kad

import (
	"container/list"
	"time"
)

// ContactContainer x
type ContactContainer struct {
	e        *list.Element
	pContact *Contact
}

// CreatedElementValue x
type CreatedElementValue struct {
	tCreated   int64
	containers map[uint32]*ContactContainer // key is IP
}

// ContactOnliner x
type ContactOnliner struct {
	pCreatedList *list.List
	contactMap   map[uint32]*ContactContainer // key is IP

	pPacketReqGuard *PacketReqGuard
}

func (co *ContactOnliner) start(pPacketReqGuard *PacketReqGuard) {
	co.pPacketReqGuard = pPacketReqGuard

	co.pCreatedList = list.New()
	co.contactMap = make(map[uint32]*ContactContainer)
}

func (co *ContactOnliner) add(t int64, pContact *Contact) {
	pContact.setCreatedime(t)

	container := ContactContainer{pContact: pContact}

	// add into contact map
	co.contactMap[pContact.ip] = &container

	// add into list
	e := co.pCreatedList.Back()
	if e == nil || t > e.Value.(*CreatedElementValue).tCreated {
		v := CreatedElementValue{tCreated: t, containers: make(map[uint32]*ContactContainer)}
		e = co.pCreatedList.PushBack(&v)
	}

	// should have same created time, but we don't check this

	// add into map with same created time
	pv := e.Value.(*CreatedElementValue)
	pv.containers[pContact.ip] = &container
	container.e = e
}

func (co *ContactOnliner) remove(pContact *Contact) {
	// remove from contact map
	pContainer := co.contactMap[pContact.ip]
	if pContainer == nil {
		return
	}

	delete(co.contactMap, pContact.ip)

	// remove element from list
	pv := pContainer.e.Value.(*CreatedElementValue)
	delete(pv.containers, pContact.ip)

	if len(pv.containers) == 0 {
		co.pCreatedList.Remove(pContainer.e)
	}
}

func (co *ContactOnliner) getSearchContacts(pSearch *Search) []*Contact {
	t := time.Now().Unix()

	// Assume longer online contact will still stay online
	var contacts []*Contact
	for e := co.pCreatedList.Front(); e != nil; e = e.Next() {
		pv := e.Value.(*CreatedElementValue)
		for _, container := range pv.containers {
			if container.pContact.getKadID() == nil {
				continue
			}

			// check can we pass from packet request guard
			tolerance := pSearch.calcSearchTolerance(container.pContact)
			opcode := kademlia2SearchKeyReq
			if tolerance > searchTolerance {
				opcode = kademlia2Req
			}
			if !co.pPacketReqGuard.canPass(t, container.pContact.ip, opcode) {
				continue
			}

			contacts = append(contacts, container.pContact)
			if len(contacts) == bootstrapSearchContactNbr {
				break
			}
		}
	}

	return contacts
}
