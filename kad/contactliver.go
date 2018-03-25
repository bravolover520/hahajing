package kad

import (
	"time"
)

// ContactLiver x
type ContactLiver struct {
	curTime    int64                         // now + 1
	contactMap map[int64]map[uint32]*Contact // map[time]map[ip]*Contact

	pPacketProcessor *PacketProcessor
}

func (cl *ContactLiver) start(pPacketProcessor *PacketProcessor) {
	cl.pPacketProcessor = pPacketProcessor
	cl.contactMap = make(map[int64]map[uint32]*Contact)
}

func (cl *ContactLiver) add2Map(pContact *Contact) {
	// add into map
	contacts := cl.contactMap[pContact.tLiveExpires]
	if contacts == nil {
		contacts = make(map[uint32]*Contact)
		cl.contactMap[pContact.tLiveExpires] = contacts
	}

	contacts[pContact.ip] = pContact
}

func (cl *ContactLiver) add(t int64, pContact *Contact) {
	if len(cl.contactMap) == 0 {
		cl.curTime = t + 1
	}

	// Here're some cases for this situation
	// If add from file or from finding node, we should check it's live or not. And get its latest details.
	// If add from receiving packet, we don't need to check its live or not.
	// But need to get its details, such as KAD ID. So here we still need to send hello for non-verified contact.
	// Verified means it's from HelloRes.
	if pContact.bVerified {
		pContact.setLiveExpiresTime(t, true) // first
		pContact.setDeadTime(true)

	} else {
		cl.pPacketProcessor.sendMyDetails(kademlia2HelloReq, pContact)
		pContact.setHelloReqTime(t)

		pContact.setLiveExpiresTime(t, false) // first
		pContact.setDeadTime(false)
	}

	// add into map
	// We don't check if its existing or not. Caller should make sure this.
	cl.add2Map(pContact)
}

func (cl *ContactLiver) remove(pContact *Contact) {
	contacts := cl.contactMap[pContact.tLiveExpires]
	delete(contacts, pContact.ip)

	if len(contacts) == 0 {
		delete(cl.contactMap, pContact.tLiveExpires)
	}
}

func (cl *ContactLiver) tickProcess(t int64) []*Contact {
	if len(cl.contactMap) == 0 {
		return nil
	}

	var deadContacts []*Contact
	for ; cl.curTime <= t; cl.curTime++ {
		contacts := cl.contactMap[cl.curTime]
		for _, pContact := range contacts {
			if t >= pContact.tDead { // now it's dead
				//com.HhjLog.Noticef("Contact %s:%d is dead\n", iIP2Str(pContact.ip), pContact.updPort)
				deadContacts = append(deadContacts, pContact)
				continue
			}

			// send new hello
			cl.pPacketProcessor.sendMyDetails(kademlia2HelloReq, pContact)
			pContact.setHelloReqTime(t)

			pContact.setLiveExpiresTime(t, false)

			// add into map again
			cl.add2Map(pContact)
		}

		// clear data of elapsed time from map
		delete(cl.contactMap, cl.curTime)
	}

	return deadContacts
}

func (cl *ContactLiver) update(pContact *Contact, bHelloRes bool) {
	t := time.Now().Unix()

	// receive packet from contact
	if bHelloRes {
		pContact.setRTT(t)
	}

	cl.remove(pContact)

	pContact.setLiveExpiresTime(t, true) // first
	pContact.setDeadTime(true)

	cl.add2Map(pContact)
}
