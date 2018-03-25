package com

import (
	"container/list"
	"strings"
	"sync"
)

const maxKeywordsNbr = 10 * 1000

// KeywordListNode is list node of items mappting to keywords.
type KeywordListNode struct {
	e *list.Element

	keyStr string
	items  []*Item
}

// KeywordManager is a manager for user search primary keywords mapping to items from DouBan.
/////////////////////////////////////////////////////////////////////////////////////////////
type KeywordManager struct {
	list    *list.List
	nodeMap map[string]*KeywordListNode // key is string of keywords.

	lock sync.RWMutex
}

// NewKeywordManager x
func NewKeywordManager() *KeywordManager {
	m := KeywordManager{list: list.New(), nodeMap: make(map[string]*KeywordListNode)}
	return &m
}

// get permutation string of @keywords, seperated by space
func getAllKeywords(keywords []string) []string {
	var allKeywords []string
	for i := 0; i < len(keywords); i++ {
		key := keywords[i]
		var subKeywords []string
		for j := 0; j < len(keywords); j++ {
			if j != i {
				subKeywords = append(subKeywords, keywords[j])
			}
		}

		if subKeywords == nil {
			allKeywords = append(allKeywords, key)
		} else {
			allSubKeywords := getAllKeywords(subKeywords)
			for _, keyStr := range allSubKeywords {
				allKeywords = append(allKeywords, key+" "+keyStr)
			}
		}
	}

	return allKeywords
}

func (m *KeywordManager) getNode(keywords []string) *KeywordListNode {
	allKeywords := getAllKeywords(keywords)
	for _, key := range allKeywords {
		if node := m.nodeMap[key]; node != nil {
			return node
		}
	}

	return nil
}

func (m *KeywordManager) setNew(keywords []string, items []*Item) {
	if len(m.nodeMap) == maxKeywordsNbr {
		// remove least recently keywords
		e := m.list.Front()
		node := e.Value.(*KeywordListNode)

		delete(m.nodeMap, node.keyStr)
		m.list.Remove(e)
	}

	keyStr := strings.Join(keywords, " ")
	node := KeywordListNode{keyStr: keyStr, items: items}
	e := m.list.PushBack(&node)
	node.e = e

	m.nodeMap[keyStr] = &node
}

func (m *KeywordManager) setExisting(node *KeywordListNode, items []*Item) {
	m.list.MoveToBack(node.e)

	var newItems []*Item
	for _, item := range items {
		existing := false
		for _, existingItem := range node.items {
			if item.OrgName == existingItem.OrgName && item.Type == existingItem.Type {
				existing = true
				break
			}
		}

		if !existing {
			newItems = append(newItems, item)
		}
	}

	node.items = append(node.items, newItems...)
}

// Set supports appending.
func (m *KeywordManager) Set(keywords []string, items []*Item) {
	if len(items) == 0 {
		return
	}

	m.lock.Lock()

	node := m.getNode(keywords)
	if node == nil {
		m.setNew(keywords, items)
	} else {
		m.setExisting(node, items)
	}

	m.lock.Unlock()
}

// Get is getting items via @keywords
func (m *KeywordManager) Get(keywords []string) []*Item {
	var items []*Item

	m.lock.RLock()

	node := m.getNode(keywords)
	if node != nil {
		items = append(items, node.items...)
	}

	m.lock.RUnlock()

	return items
}

// GetKeyStrs is getting all key strings from least recently to most recently.
func (m *KeywordManager) GetKeyStrs() []string {
	var keyStrs []string

	m.lock.RLock()

	for e := m.list.Front(); e != nil; e = e.Next() {
		node := e.Value.(*KeywordListNode)
		keyStrs = append(keyStrs, node.keyStr)
	}

	m.lock.RUnlock()

	return keyStrs
}
