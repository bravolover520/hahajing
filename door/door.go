package door

import (
	"hahajing/com"
	"sync"
	"time"
)

const (
	keywordManagerUpdateCheckTimer = 10 // minute
	keywordUpdateTimer             = 1  // second

	keywordManagerUpdateHour = 1 // 几点开始更新

	keywordCheckReqChSize = 1000
)

// KeywordCheckReq x
type KeywordCheckReq struct {
	MyKeyword *com.MyKeyword

	ResCh chan *KeywordCheckRes
}

// KeywordCheckRes x
type KeywordCheckRes struct {
	Items    []*com.Item
	ErrorStr string
}

// Door x
type Door struct {
	douBan     DouBan
	douBanLock sync.Mutex

	mtime MTime

	guard Guard

	keywordManager    *com.KeywordManager
	KeywordCheckReqCh chan *KeywordCheckReq
}

// Start x
func (d *Door) Start(keywordManager *com.KeywordManager) {
	d.keywordManager = keywordManager
	d.KeywordCheckReqCh = make(chan *KeywordCheckReq, keywordCheckReqChSize)

	if !d.douBan.start() {
		com.HhjLog.Panic("DouBan start failed!")
	}

	go d.processRoutine()
}

func (d *Door) keywordCheckReqRoutine(req *KeywordCheckReq, code int) {
	keywords := req.MyKeyword.SearchKeywords

	var items []*com.Item
	switch code {
	case douBanCode:
		d.douBanLock.Lock()
		items = d.douBan.search(keywords, 1)
		d.douBanLock.Unlock()
	case mtimeCode:
		items = d.mtime.search(keywords)
	}

	// firstly, store to keyword manager
	d.keywordManager.Set(keywords, items)

	items = com.FilterItems(items, req.MyKeyword)
	req.ResCh <- &KeywordCheckRes{Items: items}
}

func (d *Door) processKeywordCheckReq(req *KeywordCheckReq) {
	t := time.Now().Unix()
	code, ok := d.guard.add(t)
	if !ok {
		req.ResCh <- &KeywordCheckRes{Items: nil, ErrorStr: "系统忙，请等会儿重试！"}
		return
	}

	go d.keywordCheckReqRoutine(req, code)
}

func (d *Door) processRoutine() {
	timer := time.NewTicker(keywordManagerUpdateCheckTimer * time.Minute)

	preHour := time.Now().Hour()
	finishCh := make(chan bool)
	updating := false

	for {
		select {
		case pKeywordCheckReq := <-d.KeywordCheckReqCh:
			d.processKeywordCheckReq(pKeywordCheckReq)

		case t := <-timer.C:
			hour := t.Hour()

			if !updating && preHour < keywordManagerUpdateHour && hour >= keywordManagerUpdateHour {
				updating = true
				go d.updateKeywordManager(finishCh)
			}

			preHour = hour
		case <-finishCh:
			updating = false
		}
	}
}

func (d *Door) updateKeywordManager(finishCh chan bool) {
	com.HhjLog.Notice("Start updating Keyword Manager...")

	d.walkKeywordManager()

	finishCh <- true

	com.HhjLog.Notice("Finish updating Keyword Manager!")
}

func (d *Door) walkKeywordManager() {
	keyStrs := d.keywordManager.GetKeyStrs()
	for _, keyStr := range keyStrs {
		// sync from DouBan
		keywords := com.Split2Keywords(keyStr)

		d.douBanLock.Lock()
		items := d.douBan.search(keywords, 1)
		d.douBanLock.Unlock()

		d.keywordManager.Set(keywords, items)

		time.Sleep(keywordUpdateTimer * time.Second)
	}
}
