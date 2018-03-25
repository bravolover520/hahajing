package kad

import (
	"container/list"
)

// ContactListNode is list node of contact.
type ContactListNode struct {
	e        *list.Element
	pContact *Contact
}

// ContactTimeListElementValue is element value of list.
type ContactTimeListElementValue struct {
	t       int64
	nodeMap map[uint32]*ContactListNode // key is IP, node with same time will be in this map.
}

// ContactTimeList is a list with time. Note that time is always growing.
type ContactTimeList struct {
	pList   *list.List
	nodeMap map[uint32]*ContactListNode // key is IP
}

func (ctl *ContactTimeList) start() {
	ctl.pList = list.New()
	ctl.nodeMap = make(map[uint32]*ContactListNode)
}

func (ctl *ContactTimeList) remove(pContact *Contact) {
	node := ctl.nodeMap[pContact.ip]
	if node == nil {
		return
	}

	// remove from list
	pv := node.e.Value.(*ContactTimeListElementValue)
	delete(pv.nodeMap, pContact.ip)

	if len(pv.nodeMap) == 0 {
		ctl.pList.Remove(node.e)
	}
}

func (ctl *ContactTimeList) add(t int64, pContact *Contact) {
	// remove first
	ctl.remove(pContact)

	// add as new one
	node := ContactListNode{pContact: pContact}

	// add into map
	ctl.nodeMap[pContact.ip] = &node

	// add into list
	e := ctl.pList.Back()
	if e == nil || t > e.Value.(*ContactTimeListElementValue).t {
		v := ContactTimeListElementValue{t: t}
		e = ctl.pList.PushBack(&v)
	}

	// add into map with same time
	pv := e.Value.(*ContactTimeListElementValue)
	if pv.nodeMap == nil {
		pv.nodeMap = make(map[uint32]*ContactListNode)
	}
	pv.nodeMap[pContact.ip] = &node
	node.e = e
}
