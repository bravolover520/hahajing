package kad

import (
	"encoding/binary"
	"hahajing/com"
	"os"
	"time"
)

const contactTotalNbr = 1000

// ContactManager x, controlling the time entry
type ContactManager struct {
	pPerfs           *Prefs
	pPacketProcessor *PacketProcessor

	liver   ContactLiver
	onliner ContactOnliner
	finder  ContactFinder

	contactMap map[uint32]*Contact // key is IP
}

func (cm *ContactManager) start(pPerfs *Prefs, pPacketProcessor *PacketProcessor, pPacketReqGuard *PacketReqGuard) bool {
	cm.pPerfs = pPerfs
	cm.pPacketProcessor = pPacketProcessor

	cm.liver.start(pPacketProcessor)
	cm.onliner.start(pPacketReqGuard)
	cm.finder.start(pPacketProcessor)

	cm.contactMap = make(map[uint32]*Contact)

	return cm.readFile()
}

func (cm *ContactManager) readFile() bool {
	path := com.GetConfigPath()
	nodeFileName := path + "/config/kad/nodes.dat"

	f, err := os.Open(nodeFileName)
	if err != nil {
		return false
	}
	defer f.Close()

	var numContacts uint32
	binary.Read(f, binary.LittleEndian, &numContacts)

	var version uint32
	binary.Read(f, binary.LittleEndian, &version)
	if !(version >= 1 && version <= 3) {
		return false
	}

	binary.Read(f, binary.LittleEndian, &numContacts)
	for ; numContacts > 0; numContacts-- {
		var kadID ID
		var ip uint32
		var udpPort uint16
		var tcpPort uint16

		binary.Read(f, binary.LittleEndian, kadID.getHash())
		binary.Read(f, binary.LittleEndian, &ip)
		binary.Read(f, binary.LittleEndian, &udpPort)
		binary.Read(f, binary.LittleEndian, &tcpPort)

		var contactVerion uint8
		var byType byte
		if version >= 1 {
			binary.Read(f, binary.LittleEndian, &contactVerion)
		} else {
			binary.Read(f, binary.LittleEndian, &byType)
		}

		var kadUDPKey UDPKey
		var bVerified bool
		if version >= 2 {
			kadUDPKey.readFromFile(f)
			binary.Read(f, binary.LittleEndian, &bVerified)
		}

		// IP Appears invalid
		if byType >= 4 {
			continue
		}

		// always take it as not verified
		cm.addContact(&kadID, ip, udpPort, contactVerion, &kadUDPKey, false)
	}

	if len(cm.contactMap) == 0 { // no any contacts
		return false
	}

	return true
}

/*
	Note the overwriting case, it's very tricky.
	This is only one entry to add contact in routing zone.
*/
func (cm *ContactManager) addContact(
	pKadID *ID, // nil: unknown ID
	ip uint32,
	updPort uint16,
	version uint8, // 0: unknown
	pKadUDPKey *UDPKey,
	bVerified bool) (bool, *Contact) {

	if len(cm.contactMap) >= contactTotalNbr {
		return false, nil
	}

	// we don't like guy with old version
	if version != 0 && version < minSupportContactVersion {
		return false, nil
	}

	bNew := true
	pContact := cm.contactMap[ip]
	if pContact == nil { // new one
		pContact = &Contact{
			pKadID:    pKadID,
			ip:        ip,
			updPort:   updPort,
			version:   version,
			udpKey:    *pKadUDPKey,
			bVerified: bVerified}

		cm.contactMap[ip] = pContact // new one into map

		t := time.Now().Unix()
		cm.liver.add(t, pContact)
		cm.onliner.add(t, pContact)

	} else { // existing contact, overwriting it.
		if pKadID != nil { // sometimes contact don't respond to me with its KAD ID.
			pContact.pKadID = pKadID
		}
		pContact.updPort = updPort

		if version > 0 {
			pContact.version = version
		}

		pContact.udpKey = *pKadUDPKey
		if !pContact.bVerified {
			pContact.bVerified = bVerified
		}

		bNew = false
	}

	return bNew, pContact
}

func (cm *ContactManager) addKademlia2HelloRes(pMsg *Kademlia2HelloResMsg) {
	bNew, pContact := cm.addContact(
		&pMsg.contactID,
		pMsg.ip,
		pMsg.udpPort,
		pMsg.version,
		&UDPKey{key: pMsg.verifyKey, ip: cm.pPerfs.getPublicIP()},
		true) // It's verified.
	if pContact == nil {
		return
	}

	// It's good place we start node finding, because we have it's details.
	cm.finder.add(pContact)

	// It's live, just update it. So that we don't need to send hello to it.
	if !bNew {
		cm.liver.update(pContact, true)
	}
}

func (cm *ContactManager) addContactFromKademlia2Res(pKadID *ID, ip uint32, udpPort uint16, version uint8) {
	pContact := cm.contactMap[ip]
	if pContact == nil { // new one, just add it. Don't care bout successful or not.
		cm.addContact(
			pKadID,
			ip,
			udpPort,
			version,
			&UDPKey{key: 0, ip: 0},
			false)
		return
	}

	// Alreay existing, we should take carefully what we'll be updated.
	// Should we trust it? No
	// In case of routing zone is full, here's different implementation with KAD protocol that
	// we don't send hello to least recently live node to check it's live or not.
	// And then decide to what to do.
}

func (cm *ContactManager) addKademlia2Res(pMsg *Kademlia2ResMsg) {
	// update contact who send this RES to us
	bNew, pContact := cm.addContact(
		nil,
		pMsg.ip,
		pMsg.udpPort,
		0,
		&UDPKey{key: pMsg.verifyKey, ip: cm.pPerfs.getPublicIP()},
		false)

	// we notify it's live if already existing
	if pContact != nil && !bNew {
		cm.liver.update(pContact, false)
	}

	// add contacts
	for _, p := range pMsg.contacts {
		cm.addContactFromKademlia2Res(p.pKadID, p.ip, p.updPort, p.version)
	}
}

func (cm *ContactManager) tickProcess() {
	t := time.Now().Unix()

	// live is always check firstly so that we can remove dead contacts early.
	deadContacts := cm.liver.tickProcess(t)

	// remove dead contacts
	for _, pContact := range deadContacts {
		delete(cm.contactMap, pContact.ip)
		cm.onliner.remove(pContact)
		cm.finder.remove(pContact)
	}

	// we still need find more nodes
	if len(cm.contactMap) < contactTotalNbr {
		cm.finder.tickProcess()
	} else {
		//com.HhjLog.Infof("Reach limit of contacts!")
	}
}
